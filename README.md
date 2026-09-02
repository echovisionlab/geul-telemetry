# Geul telemetry

Shared logging, tracing, audit, security, request-correlation, and redaction
contracts for Geul services.

- Go module: `github.com/echovisionlab/geul-telemetry`
- npm package: `@echovisionlab/geul-telemetry`

```sh
pnpm add @echovisionlab/geul-telemetry
```

Browser code should import the `/actor`, `/redaction`, or `/trace` subpaths.
The package root also includes Node request context and is server-only.

## Development

```sh
corepack enable
pnpm install --frozen-lockfile
pnpm format:check
pnpm lint
pnpm typecheck
pnpm test:coverage
go test -race ./...
go vet ./...
```

## Release

One Release Please version and `v*` tag cover both Go and npm consumers. npm
publication uses GitHub Actions trusted publishing without a repository npm
token.

## License

PolyForm Noncommercial 1.0.0. Commercial use requires a separate license from
Echo Vision Lab. See [LICENSE.md](LICENSE.md).
