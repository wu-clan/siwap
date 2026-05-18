# Build Assets

This directory contains desktop build metadata and icons used by Wails v3.

- `config.yml` contains project metadata and development-mode commands
- `appicon.png` is the source app icon
- `darwin/` contains macOS plist assets
- `linux/` contains Linux desktop/package metadata
- `windows/` contains Windows manifest, icon, and installer assets

Use the root Taskfile for normal development commands:

```sh
wails3 task dev
wails3 task build
wails3 task package
```
