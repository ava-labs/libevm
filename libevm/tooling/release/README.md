# Release

In the following, we create a release candidate version `v1.13.14-0.2.0.rc.4` and your Github username is `myusername`. We set environment variables accordingly:

```bash
VERSION=v1.13.14-0.2.0.rc.4
USERNAME=myusername
```

1. Create two branches, usually from the tip of the `main` branch, but the base commit MAY be from anywhere on `main`:

    ```bash
    git fetch origin main:main
    git checkout main
    git checkout -b "release/$VERSION"
    git push -u origin "release/$VERSION"
    git checkout -b "$USERNAME/release/$VERSION"
    ```

    The `$USERNAME/release/$VERSION` branch will be used to add "release modifications" and will target the branch `release/$VERSION` in a pull request.
1. Run the script `./cherrypick.sh` which cherry picks all Geth commits listed in [cherrypicks](cherrypicks). You may have to resolve conflicts. If you encounter a CI error NOT from the `go_test_tooling` job, ideally [amend](https://git-scm.com/book/en/v2/Git-Tools-Rewriting-History) the cherry-picked commit(s) responsible for each issue.
1. Modify [params/version.libevm.go](/params/version.libevm.go)
    - Change the `LibEVMReleaseType` to the correct release type, for example `ReleaseCandidate`
    - If planning a release candidate: set `libEVMReleaseCandidate` to the correct number; in this case `4`
    - If needed: change the `LibEVMVersionMajor`, `LibEVMVersionMinor` and `LibEVMVersionPatch` numbers
    - Change the `LibEVMVersion` to the correct final version string `$VERSION` (`version.libevm_test.go` will ensure that this is correctly formatted)
1. Commit your modifications to [params/version.libevm.go](/params/version.libevm.go) with a commit title `chore: release v1.13.14-0.2.0.rc.4`:

    ```bash
    git commit -m "chore: release $VERSION"
    ```

1. Push your modified branch to the remote `git push -u origin $USERNAME/release/$VERSION`
1. Open a pull request from your modified branch `$USERNAME/release/v1.13.14-0.2.0.rc.4` and targeting `release/v1.13.14-0.2.0.rc.4`:
    1. You can create the pull request for example using

        ```bash
        open -a "Google Chrome" https://github.com/ava-labs/libevm/compare/release/$VERSION...$USERNAME/release/$VERSION
        ```

    1. Set the title of the pull request to

        ```txt
        chore: release `v1.13.14-0.2.0.rc.4`
        ```

1. Wait for all the checks to pass on the pull request
1. Fast forward merge your branch in the release branch:

    ```bash
    git checkout "release/$VERSION"
    git merge --ff-only "$USERNAME/release/$VERSION"
    git push
    ```

    This will also close the ongoing pull request.
1. Finally create your tag and push it using the release branch:

    ```bash
    git tag "$VERSION"
    git push origin "$VERSION"
    ```

1. (optional) you can then cleanup with:

    ```bash
    git branch -D "release/$VERSION" "$USERNAME/release/$VERSION"
    git push -d origin "$USERNAME/release/$VERSION"
    ```
