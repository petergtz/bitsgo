package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/petergtz/bitsgo/ccupdater"

	"github.com/benbjohnson/clock"
	"github.com/petergtz/bitsgo"
	"github.com/petergtz/bitsgo/blobstores/azure"
	"github.com/petergtz/bitsgo/blobstores/decorator"
	"github.com/petergtz/bitsgo/blobstores/gcp"
	"github.com/petergtz/bitsgo/blobstores/local"
	"github.com/petergtz/bitsgo/blobstores/openstack"
	"github.com/petergtz/bitsgo/blobstores/s3"
	"github.com/petergtz/bitsgo/blobstores/webdav"
	"github.com/petergtz/bitsgo/config"
	log "github.com/petergtz/bitsgo/logger"
	"github.com/petergtz/bitsgo/middlewares"
	"github.com/petergtz/bitsgo/pathsigner"
	"github.com/petergtz/bitsgo/routes"
	"github.com/petergtz/bitsgo/statsd"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configPath = kingpin.Flag("config", "specify config to use").Required().Short('c').String()
)

func main() {
	kingpin.Parse()

	config, e := config.LoadConfig(*configPath)

	if e != nil {
		log.Log.Fatalw("Could not load config.", "error", e)
	}
	log.Log.Infow("Logging level", "log-level", config.Logging.Level)
	logger := createLoggerWith(config.Logging.Level)
	log.SetLogger(logger)

	metricsService := statsd.NewMetricsService()

	appStashBlobstore := createAppStashBlobstore(config.AppStash, metricsService)
	packageBlobstore, signPackageURLHandler := createBlobstoreAndSignURLHandler(config.Packages, config.PublicEndpointUrl(), config.Port, config.Secret, "packages", metricsService)
	dropletBlobstore, signDropletURLHandler := createBlobstoreAndSignURLHandler(config.Droplets, config.PublicEndpointUrl(), config.Port, config.Secret, "droplets", metricsService)
	buildpackBlobstore, signBuildpackURLHandler := createBlobstoreAndSignURLHandler(config.Buildpacks, config.PublicEndpointUrl(), config.Port, config.Secret, "buildpacks", metricsService)
	buildpackCacheBlobstore, signBuildpackCacheURLHandler := createBuildpackCacheSignURLHandler(config.Droplets, config.PublicEndpointUrl(), config.Port, config.Secret, metricsService)

	handler := routes.SetUpAllRoutes(
		config.PrivateEndpointUrl().Host,
		config.PublicEndpointUrl().Host,
		middlewares.NewBasicAuthMiddleWare(basicAuthCredentialsFrom(config.SigningUsers)...),
		&local.SignatureVerificationMiddleware{&pathsigner.PathSignerValidator{config.Secret, clock.New()}},
		signPackageURLHandler,
		signDropletURLHandler,
		signBuildpackURLHandler,
		signBuildpackCacheURLHandler,
		bitsgo.NewAppStashHandler(appStashBlobstore, config.AppStash.MaxBodySizeBytes()),
		bitsgo.NewResourceHandlerWithUpdater(
			packageBlobstore,
			createUpdater(config.CCUpdater),
			"package",
			metricsService,
			config.Packages.MaxBodySizeBytes()),
		bitsgo.NewResourceHandler(buildpackBlobstore, "buildpack", metricsService, config.Buildpacks.MaxBodySizeBytes()),
		bitsgo.NewResourceHandler(dropletBlobstore, "droplet", metricsService, config.Droplets.MaxBodySizeBytes()),
		bitsgo.NewResourceHandler(buildpackCacheBlobstore, "buildpack_cache", metricsService, config.BuildpackCache.MaxBodySizeBytes()))

	address := os.Getenv("BITS_LISTEN_ADDR")
	if address == "" {
		address = "0.0.0.0"
	}

	log.Log.Infow("Starting server", "port", config.Port)
	httpServer := &http.Server{
		Handler: negroni.New(
			middlewares.NewMetricsMiddleware(metricsService),
			middlewares.NewZapLoggerMiddleware(log.Log),
			negroni.Wrap(handler)),
		Addr:         fmt.Sprintf("%v:%v", address, config.Port),
		WriteTimeout: 60 * time.Minute,
		ReadTimeout:  60 * time.Minute,
		ErrorLog:     zap.NewStdLog(logger),
	}
	e = httpServer.ListenAndServeTLS(config.CertFile, config.KeyFile)
	log.Log.Fatalw("http server crashed", "error", e)
}

func createLoggerWith(logLevel string) *zap.Logger {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zapLogLevelFrom(logLevel)
	logger, e := loggerConfig.Build()
	if e != nil {
		log.Log.Panic(e)
	}
	return logger
}

func zapLogLevelFrom(configLogLevel string) zap.AtomicLevel {
	switch strings.ToLower(configLogLevel) {
	case "", "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		log.Log.Fatal("Invalid log level in config", "log-level", configLogLevel)
		return zap.NewAtomicLevelAt(-1)
	}
}

func basicAuthCredentialsFrom(configCredententials []config.Credential) (basicAuthCredentials []middlewares.Credential) {
	basicAuthCredentials = make([]middlewares.Credential, len(configCredententials))
	for i := range configCredententials {
		basicAuthCredentials[i] = middlewares.Credential(configCredententials[i])
	}
	return
}

func createBlobstoreAndSignURLHandler(
	blobstoreConfig config.BlobstoreConfig,
	publicEndpoint *url.URL,
	port int,
	secret string,
	resourceType string,
	metricsService bitsgo.MetricsService) (bitsgo.Blobstore, *bitsgo.SignResourceHandler) {

	localResourceSigner := createLocalResourceSigner(publicEndpoint, port, secret, resourceType)
	if blobstoreConfig.BlobstoreType == config.Local {
		log.Log.Infow("Creating local blobstore", "path-prefix", blobstoreConfig.LocalConfig.PathPrefix)
		return decorator.ForBlobstoreWithPathPartitioning(
				decorator.ForBlobstoreWithMetricsEmitter(
					local.NewBlobstore(*blobstoreConfig.LocalConfig),
					metricsService,
					resourceType)),
			bitsgo.NewSignResourceHandler(localResourceSigner, localResourceSigner)
	}

	var blobstore decorator.Blobstore = createNonLocalBlobstore(blobstoreConfig)
	var signer bitsgo.ResourceSigner = createNonLocalBlobstore(blobstoreConfig)

	if blobstoreConfig.BlobstoreType == config.WebDAV {
		blobstore = decorator.ForBlobstoreWithPathPrefixing(blobstore, resourceType+"/")
		signer = decorator.ForResourceSignerWithPathPrefixing(signer, resourceType+"/")
	}

	return decorator.ForBlobstoreWithPathPartitioning(
			decorator.ForBlobstoreWithMetricsEmitter(blobstore, metricsService, resourceType)),
		bitsgo.NewSignResourceHandler(
			decorator.ForResourceSignerWithPathPartitioning(signer),
			localResourceSigner)
}

func createBuildpackCacheSignURLHandler(blobstoreConfig config.BlobstoreConfig,
	publicEndpoint *url.URL,
	port int,
	secret string,
	metricsService bitsgo.MetricsService) (bitsgo.Blobstore, *bitsgo.SignResourceHandler) {

	if blobstoreConfig.BlobstoreType == config.Local {
		log.Log.Infow("Creating local blobstore", "path-prefix", blobstoreConfig.LocalConfig.PathPrefix)
		localResourceSigner := createLocalResourceSigner(publicEndpoint, port, secret, "buildpack_cache/entries")
		return decorator.ForBlobstoreWithPathPartitioning(
				decorator.ForBlobstoreWithPathPrefixing(
					decorator.ForBlobstoreWithMetricsEmitter(
						local.NewBlobstore(*blobstoreConfig.LocalConfig),
						metricsService,
						"buildpack_cache"),
					"buildpack_cache/")),
			bitsgo.NewSignResourceHandler(localResourceSigner, localResourceSigner)
	}
	prefix := "buildpack_cache/"
	if blobstoreConfig.BlobstoreType == config.WebDAV {
		prefix = "droplets/buildpack_cache/"
	}

	return decorator.ForBlobstoreWithPathPartitioning(
			decorator.ForBlobstoreWithPathPrefixing(
				decorator.ForBlobstoreWithMetricsEmitter(
					createNonLocalBlobstore(blobstoreConfig),
					metricsService,
					"buildpack_cache"),
				prefix)),
		bitsgo.NewSignResourceHandler(
			decorator.ForResourceSignerWithPathPartitioning(
				decorator.ForResourceSignerWithPathPrefixing(
					createNonLocalBlobstore(blobstoreConfig),
					"buildpack_cache")),
			createLocalResourceSigner(publicEndpoint, port, secret, "buildpack_cache/entries"))
}

func createLocalResourceSigner(publicEndpoint *url.URL, port int, secret string, resourceType string) bitsgo.ResourceSigner {
	return &local.LocalResourceSigner{
		DelegateEndpoint:   fmt.Sprintf("%v://%v:%v", publicEndpoint.Scheme, publicEndpoint.Host, port),
		Signer:             &pathsigner.PathSignerValidator{secret, clock.New()},
		ResourcePathPrefix: "/" + resourceType + "/",
	}
}

func createAppStashBlobstore(blobstoreConfig config.BlobstoreConfig, metricsService bitsgo.MetricsService) bitsgo.NoRedirectBlobstore {
	if blobstoreConfig.BlobstoreType == config.Local {
		log.Log.Infow("Creating local blobstore", "path-prefix", blobstoreConfig.LocalConfig.PathPrefix)
		return decorator.ForBlobstoreWithPathPartitioning(decorator.ForBlobstoreWithPathPrefixing(
			local.NewBlobstore(*blobstoreConfig.LocalConfig),
			"app_bits_cache/"))
	}
	prefix := "app_bits_cache/"
	if blobstoreConfig.BlobstoreType == config.WebDAV {
		prefix = "app_stash/app_bits_cache/"
	}
	return decorator.ForBlobstoreWithPathPartitioning(
		decorator.ForBlobstoreWithPathPrefixing(
			decorator.ForBlobstoreWithMetricsEmitter(
				createNonLocalBlobstore(blobstoreConfig),
				metricsService,
				"app_stash"),
			prefix))
}

type blobstoreResourceSigner interface {
	decorator.Blobstore
	bitsgo.ResourceSigner
}

func createNonLocalBlobstore(blobstoreConfig config.BlobstoreConfig) blobstoreResourceSigner {
	switch blobstoreConfig.BlobstoreType {
	case config.Local:
		panic("Should not get here")
	case config.AWS:
		log.Log.Infow("Creating S3 blobstore", "bucket", blobstoreConfig.S3Config.Bucket)
		return s3.NewBlobstore(*blobstoreConfig.S3Config)
	case config.Google:
		log.Log.Infow("Creating GCP blobstore", "bucket", blobstoreConfig.GCPConfig.Bucket)
		return gcp.NewBlobstore(*blobstoreConfig.GCPConfig)
	case config.Azure:
		log.Log.Infow("Creating Azure blobstore", "container", blobstoreConfig.AzureConfig.ContainerName)
		return azure.NewBlobstore(*blobstoreConfig.AzureConfig)
	case config.OpenStack:
		log.Log.Infow("Creating Openstack blobstore", "container", blobstoreConfig.OpenstackConfig.ContainerName)
		return openstack.NewBlobstore(*blobstoreConfig.OpenstackConfig)
	case config.WebDAV:
		log.Log.Infow("Creating Webdav blobstore",
			"public-endpoint", blobstoreConfig.WebdavConfig.PublicEndpoint,
			"private-endpoint", blobstoreConfig.WebdavConfig.PrivateEndpoint)
		return webdav.NewBlobstore(*blobstoreConfig.WebdavConfig)
	default:
		log.Log.Fatalw("blobstoreConfig is invalid.", "blobstore-type", blobstoreConfig.BlobstoreType)
		return nil // satisfy compiler
	}
}

func createUpdater(ccUpdaterConfig *config.CCUpdaterConfig) bitsgo.Updater {
	if ccUpdaterConfig == nil {
		return &bitsgo.NullUpdater{}
	}
	return ccupdater.NewCCUpdater(
		ccUpdaterConfig.Endpoint,
		ccUpdaterConfig.Method,
		ccUpdaterConfig.ClientCertFile,
		ccUpdaterConfig.ClientKeyFile,
		ccUpdaterConfig.CACertFile)
}
