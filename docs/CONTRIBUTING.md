# Development

## Building

`make build` should create a binary for your current OS and arch into the `bin` directory.

## Releasing

We're using [goreleaser](https://goreleaser.com/intro/) to automate building and archiving. If enough changes have
accumulated that merit a release, you can:

1. Tag a new version, e.g. `git tag v0.33.0`.
2. See locally what goreleaser would do with it using `make release`.
3. Check the `dist` directory for the output archives.
4. If you're satisfied, push the tag to trigger the release pipeline: `git push v0.33.0`.
