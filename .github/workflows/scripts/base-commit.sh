#!/bin/bash

# This script finds the base commit libevm changes should be considered from,
# and returns it to the caller.
# This base commit is the first child commit with its message containing
# "rename Go module + update internal import paths" of the last
# go-ethereum (aka geth) tag present in the main branch of libevm.
# If an error occurs, this one is echo-ed and the script exits with code 1.


base_geth_tag_hash="$($(dirname "$0")/geth-commit.sh)"

base_commit_hash="$(git log --oneline --grep="rename Go module + update internal import paths" --after "$base_geth_tag_hash" -n 1 origin/main | head -n 1 | awk '{print $1}')"
if [ -z "$base_commit_hash" ]; then
  echo "No child commit found after tag $base_geth_tag_hash on branch main."
  exit 1
fi

echo $base_commit_hash | tr -d '\n'