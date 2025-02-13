#!/usr/bin/env bash

set -eu;
set -o pipefail;

SELF_DIR=$(dirname $0)
# The format of the `cherrypicks` file is guaranteed by a test so we can use simple parsing here.
CHERRY_PICKS=$(cat "${SELF_DIR}/cherrypicks" | grep -Pv "^#" | awk '{print $1}')
RELEASE_REF="main";

commits=()
for commit in ${CHERRY_PICKS}; do
    git merge-base --is-ancestor "${commit}" "${RELEASE_REF}" && \
        echo "Skipping ${commit} already in history" && \
        continue;

    echo "Cherry-picking ${commit}";
    commits+=("${commit}");
done

git cherry-pick "${commits[@]}";
