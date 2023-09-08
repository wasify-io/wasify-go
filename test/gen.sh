#!/bin/sh

DIR="_data"

cd "$DIR"

for gofile in *.go; do
    base_name=$(basename "$gofile" .go)

    # Use tinygo to build the .wasm file
    tinygo build -scheduler=none -o "${base_name}.wasm" -target wasi "$gofile"
done

cd -

echo "WASM generation complete."
