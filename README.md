# Siwap

[简体中文](README.zh-CN.md)

Siwap is a desktop launcher for project-based AI assistant sessions. It stays as a compact left-side sidebar, lets you choose a project, Git worktree, terminal, and AI assistant, then opens the session in your preferred terminal.

![Siwap demo](siwap.gif)

## Features

- Project and Git worktree context switching
- Built-in Claude Code, Codex, and OpenCode launchers
- Custom AI assistants and custom terminal profiles
- Session focus, close, filtering, and failed-launch retention
- Always-on-top, focus-loss hiding, left-edge reveal, and configurable summon shortcut
- System, light, and dark appearance with Chinese and English language support

## Tech Stack

- Wails v3 alpha
- Go backend
- Vue 3 + TypeScript
- Vite Plus
- Tailwind CSS
- shadcn-vue
- vue-i18n
- pnpm

## Requirements

- Go
- Wails CLI
- Node.js with pnpm
- Git
- Platform runtime dependencies required by Wails

Install Wails CLI if needed:

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha.91
```

## Development

Install frontend dependencies:

```sh
cd frontend
pnpm install
```

Run in development mode:

```sh
wails3 task dev
```

Run checks:

```sh
wails3 generate bindings -clean=true -ts ./...
go test ./...
go build ./...
cd frontend && pnpm build
```

Build the desktop app:

```sh
wails3 task build
```

Create a platform package:

```sh
wails3 task package
```

## Scripts

```sh
scripts/dev.sh
scripts/release.sh <version>
```

`dev.sh` runs frontend install/build, Go tests, and a Wails v3 build. `release.sh` builds a release binary and can sign macOS artifacts when `SIWAP_CODESIGN_IDENTITY` is configured.
