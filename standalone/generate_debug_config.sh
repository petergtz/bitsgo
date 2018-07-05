#!/bin/bash

cat <<EOF > go-patch.yml
- type: replace
  path: /cert_file
  value: (( concat \$GOPATH "/src/github.com/petergtz/bitsgo/standalone/cert_file" ))

- type: replace
  path: /key_file
  value: (( concat \$GOPATH "/src/github.com/petergtz/bitsgo/standalone/key_file" ))
EOF

spruce merge --go-patch config.yml go-patch.yml > vscode_debug_config.yml
rm go-patch.yml
