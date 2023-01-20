#!/bin/bash
set -o errexit -o nounset -o pipefail
command -v shellcheck > /dev/null && shellcheck "$0"

if [ $# -ne 1 ]; then
  echo "Usage: ./download_releases.sh RELEASE_TAG"
  exit 1
fi

tag="TBD"

for contract in hackatom reflect; do
  url="https://github.com/CosmWasm/cosmwasm/releases/download/$tag/${contract}.wasm"
  echo "Downloading $url ..."
  wget -O "${contract}.wasm" "$url"
done

tag="$1"
url="https://github.com/crescent-network/crescent-cosmwasm/releases/download/$tag/crescent_examples.wasm"
echo "Downloading $url ..."
wget -O "crescent_reflect.wasm" "$url"

rm -f version.txt
echo "$tag" >version.txt