# Siwap Frontend

Vue 3 + TypeScript frontend for the Siwap desktop app.

## Development

```sh
pnpm install
pnpm build
```

From the repository root, Wails v3 bindings are generated into `frontend/bindings`:

```sh
$(go env GOPATH)/bin/wails3 generate bindings -clean=true -ts ./...
```
