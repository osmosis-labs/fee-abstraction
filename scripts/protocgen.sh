#!/usr/bin/env sh

set -eo pipefail

# get protoc executions
echo "Generating gogo proto code"
cd proto
proto_dirs=$(find ./feeabstraction/feeabs/v1beta1/ -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep go_package $file &>/dev/null; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cd ..

# move proto files to the right places
#
# Note: Proto files are suffixed with the current binary version.
cp -r github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types/* ./x/feeabs/types/
rm -rf github.com
