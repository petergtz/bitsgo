# Debug the bit-service with vscode

## pre requiste
Go for Visual Studio Code

## setup
1. install dlv
    ```
    bash$ go get -u github.com/derekparker/delve/cmd/dlv
    ```
    verify delve
    ```
    bash$ dlv version
    Delve Debugger
    Version: 0.12.2
    Build: v0.12.2
    ```

1. generate a bit-service config

    From inside the folder $GOPATH/src/github.com/petergtz/bitsgo/standalone execute the script `generate_debug_config.sh`
    This creates a config file for the bit-service which will be used from vscode.

1. Configure VSCODE
    Put this launch config into `$GOPATH/.vscode` folder.

    ```
    {
        // Use IntelliSense to learn about possible attributes.
        // Hover to view descriptions of existing attributes.
        // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
        "version": "0.2.0",
        "configurations": [

            {
                "name": "Launch",
                "type": "go",
                "request": "launch",
                "mode": "debug",
                "remotePath": "",
                "port": 2345,
                "host": "127.0.0.1",
                "program":"${workspaceRoot}/src/github.com/petergtz/bitsgo/cmd/bitsgo/main.go",
                "env": {},
                "args": ["-c", "${workspaceRoot}/src/github.com/petergtz/bitsgo/standalone/vscode_debug_config.yml"],
                "showLog": true
            }
        ]
    }
    ```
    Next step is open vscode and switch to the debug config and launch the bits-service from inside the vscode.
    Now you can set some breakpoints and debug some areas of the bits-serivce.
