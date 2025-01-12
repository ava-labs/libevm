#!/bin/bash

# This script finds the base commit corresponding to a geth tag
# from which libevm was last based on, and returns it to the caller.
# If an error occurs, this one is echo-ed and the script exits with code 1.

git remote add geth https://github.com/ethereum/go-ethereum.git
git fetch geth 'refs/tags/*:refs/tags/*'

geth_tags=$(git ls-remote --tags --sort=-version:refname geth "v[1-9]*" | awk '{print $2}' | sed 's/refs\/tags\///')

base_geth_tag_hash=
for geth_tag in $geth_tags; do
  geth_tag_hash="$(git rev-parse --short $geth_tag)"
  if git merge-base --is-ancestor $geth_tag_hash "origin/main"; then
    base_geth_tag_hash="$geth_tag_hash"
    break
  fi
done

if [ -z "$base_geth_tag_hash" ]; then
  echo "No geth tag found in libevm main branch."
  exit 1
fi

echo $base_geth_tag_hash | tr -d '\n'
