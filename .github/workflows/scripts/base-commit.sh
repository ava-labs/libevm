#!/bin/bash

# This script finds the base commit libevm changes should be considered from,
# and returns it to the caller.
# This base commit is the first child commit with its message containing
# "rename Go module + update internal import paths" of the last
# go-ethereum (aka geth) tag present in the main branch of libevm.
# If an error occurs, this one is echo-ed and the script exits with code 1.

git remote add geth https://github.com/ethereum/go-ethereum.git
git fetch geth 'refs/tags/*:refs/tags/*'

geth_tags=$(git ls-remote --tags --sort=-version:refname geth "v[1-9]*" | awk '{print $2}' | sed 's/refs\/tags\///')

base_geth_tag_hash=
for geth_tag in $geth_tags; do
  geth_tag_hash="$(git rev-parse $geth_tag)"
  if git merge-base --is-ancestor $geth_tag_hash "origin/main"; then
    base_geth_tag_hash="$geth_tag_hash"
    break
  fi
done

if [ -z "$base_geth_tag_hash" ]; then
  echo "No geth tag found in libevm main branch."
  exit 1
fi

base_commit_hash="$(git log --oneline --grep="rename Go module + update internal import paths" --after "$base_geth_tag_hash" -n 1 origin/main | head -n 1 | awk '{print $1}')"
if [ -z "$base_commit_hash" ]; then
  echo "No child commit found after tag $base_geth_tag_hash on branch main."
  exit 1
fi

echo $base_commit_hash | tr -d '\n'