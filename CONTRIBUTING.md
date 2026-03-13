# Contributing

## Development

```sh
git clone https://github.com/skyfriends/grove.git
cd grove
go build -o grove .
./grove help
```

## Making changes

1. Fork the repo and create a branch
2. Make your changes
3. Test locally: `go build -o grove . && ./grove`
4. Open a PR

## Commit messages

Use [conventional commits](https://www.conventionalcommits.org):

```
feat: add new command
fix: handle missing config gracefully
docs: update install instructions
```

These drive the automated changelog in releases.

## Releases

Releases are automated via [GoReleaser](https://goreleaser.com). When a version tag is pushed (`v0.2.0`), CI builds cross-platform binaries and publishes a GitHub release with a changelog.

To test the release build locally:

```sh
make snapshot
```
