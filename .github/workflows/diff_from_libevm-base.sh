#!/usr/bin/env bash

set -eux;

git checkout "${1}" --;
git diff --diff-filter=a --word-diff --unified=0 libevm-base.."${1}" \
    ':(exclude).golangci.yml' \
    ':(exclude).github/**';
