# Contributing to VibeLog

Thanks for taking the time to contribute.

## Before you start

- Open an issue for larger changes so the approach can be discussed first.
- Keep the existing architecture: `cmd/` for the entrypoint, `internal/` for
  packages, tests next to the code they cover.
- Run `CGO_ENABLED=0 go test ./...` and `cd frontend && npm run typecheck`
  before opening a pull request.
- Never commit secrets, `.env` files, databases, or build artifacts.

## License of contributions

By submitting a pull request, patch, or any other contribution, you agree that:

1. Your contribution is licensed under the repository's current source
   license (see `LICENSE`).
2. You grant Muhammad Navid Athallah a perpetual, worldwide, non-exclusive,
   royalty-free, irrevocable license to use, reproduce, modify, distribute,
   sublicense, and relicense your contribution, including under a commercial
   license.
3. You have the right to grant this license for the contribution.

If you do not agree with these terms, please do not submit a contribution.

## Security

Do not open a public issue for a vulnerability. Follow `SECURITY.md`.
