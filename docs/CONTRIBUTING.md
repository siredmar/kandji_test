# Development

## Building

`make build` should create a binary for your current OS and arch into the `bin` directory.

## Releasing

The main release branch is `master`. For preparing releases the `develop` branch is used. This branch is used to accumulate changes and test them before merging them into `master`.
For releasing `semantic-release` is used which automatically creates a new tag and release notes based on the commit messages.

The release process is as follows:
1. commit to `develop`. This creates a new version with the `-develop` suffix.
2. once you are happy with the changes, merge `develop` into `master`. This creates a new version and does the actual release on Slack.

## Manual releasing (e.g. for EIS)

You also can run the release process manually. This is useful if you want to release a version for EIS. To do this, run the following command. 
You have to provide it a version and the path to the changelog file.

```bash
.buildkite/steps/build_release.sh <version> <changelog file>
```

The results will be located in `dist/`.
