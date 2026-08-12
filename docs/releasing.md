# Releasing Mayfly

Mayfly uses Go module versioning and Semantic Versioning. Release tags have the
form `vMAJOR.MINOR.PATCH`; the module remains in the unstable `v0` series until
its public API is declared stable.

## Version policy

- Increment `PATCH` for backward-compatible fixes and documentation changes.
- Increment `MINOR` for backward-compatible features. While the module is at
  `v0`, use a minor release for any intentional public API break and call it
  out prominently in the changelog.
- Use a prerelease suffix such as `v0.5.0-rc.1` for release candidates.
- Starting with `v1.0.0`, increment `MAJOR` for breaking changes. A future v2
  must also change the module path to `github.com/cwbudde/mayfly/v2`.
- Never move or replace a published tag. Publish a new version for corrections.

## Release checklist

1. Choose the next version from the changes since the latest tag.
2. Move the relevant entries from `Unreleased` in `CHANGELOG.md` into a dated
   version section and update its comparison links.
3. Run `just release-check version=MAJOR.MINOR.PATCH`.
4. Confirm the package overview with `go doc .` and verify all README and docs
   links resolve in the repository browser.
5. Commit the release preparation with `chore: prepare vMAJOR.MINOR.PATCH`.
6. Run `just release version=MAJOR.MINOR.PATCH` to create the annotated tag.
7. Push the commit and tag: `git push origin main vMAJOR.MINOR.PATCH`.
8. Confirm the GitHub release-validation workflow succeeds.
9. Ask the public Go proxy to fetch the immutable tag:
   `GOPROXY=proxy.golang.org go list -m github.com/cwbudde/mayfly@vMAJOR.MINOR.PATCH`.
10. Verify the version and rendered package documentation at
    <https://pkg.go.dev/github.com/cwbudde/mayfly>.

Go modules are published from repository tags rather than uploaded directly to
pkg.go.dev. Fetching a pushed tag through the public proxy makes it discoverable
by the Go package site.

## Validation workflow

`.github/workflows/release.yml` runs for SemVer-shaped tags and can also be
started manually. It validates the version, module metadata, license and
changelog, then runs static analysis and the complete test suite. The workflow
does not create tags or GitHub releases; those remain deliberate maintainer
actions.
