# `semver-checker`

## Introduction
This application queries a Gitlab instance for a specific version of a generic package, or latest if the given version string is 'latest'.

## TL;DR
1. Set up your `~/.semver-checker.yaml`:
    ```yaml
    host: <GITLAB HOST URL>
    token: <ACCESS TOKEN WITH REGISTRY READ RIGHTS>
    project: <PATH GROUP/PROJECT WHERE PACKAGES ARE STORED>
    ```
2. Build this app.
3. Run this app with a package name & version to search for:
    ```sh
    $ semver-checker package@1.0.0
    $ semver-checker package@latest
    # Omitting the version returns the latest.
    $ semver-checker package
    ```
