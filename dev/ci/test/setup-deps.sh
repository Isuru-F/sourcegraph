#!/usr/bin/env bash

set -euxo pipefail

asdf install
yarn
yarn generate

curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
