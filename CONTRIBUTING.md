# Contributing to evmgo

Thanks for contributing to `evmgo`.

## Getting Started

1. Fork the repository.
2. Clone your fork.
3. Install a Go version that matches `go.mod`.
4. Run the test suite before making changes:

```bash
go test ./...
```

If you want to build the CLI locally:

```bash
go build ./cmd/evmgo
```

## Before Opening An Issue

Use GitHub Issues for:

- bug reports
- feature requests

Use GitHub Discussions for:

- usage questions
- troubleshooting help
- general support

Discussions: https://github.com/itzfelixv/evmgo/discussions

Before opening a new issue, search existing issues and discussions first.

## Before Opening A Pull Request

Before you open a PR:

1. Update from the current `master` branch.
2. Run:

```bash
go test ./...
```

3. Describe what changed and how you verified it.
4. Update docs when user-facing behavior changes.

Major features are welcome. You do not need to artificially split work just to
fit a repository rule, but explain the user need and behavior change clearly.

## Commit Messages

Conventional commits are recommended for consistency, but they are not a hard
requirement.

Examples:

- `feat: add calldata decode helper`
- `fix: preserve tx json contract in tx view core`
- `docs: update tx command examples`
- `test: cover revert reason decoding`

## Review Expectations

Maintainers may ask for follow-up changes before merge.

Reviews are based on:

- correctness
- project fit
- test coverage and verification
- clarity of user-facing behavior

Not every proposal will merge as-is, especially if the scope, UX, or long-term
maintenance cost is unclear.
