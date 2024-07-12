# Introduction
This application queries the given Gitlab instance for a specific version of a generic package, or latest if the given version string is 'latest'.
# TL;DR
1. Set up your `~/.semver-checker.yaml`:
    ```yaml
    host: <YOUR GITLAB HOST URL>
    token: <YOUR ACCESS TOKEN WITH REGISTRY READ RIGHTS>
    projects: <YOUR PROJECT WHERE PACKAGES ARE STORED>
    ```
3. Build this app.
4. Run this app with a package name & version to search for.
