#!/bin/sh

DIR="$(pwd)/test/_data"

for gofile in "${DIR}"/*.go; do
    base_name=$(basename "$gofile" .go)

    # Use tinygo to build the .wasm file
    tinygo build -scheduler=none -o "${DIR}/${base_name}.wasm" -target wasi "$gofile"
done

echo "WASM generation complete."
