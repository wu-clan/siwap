# Siwap

[English](README.md)

Siwap 是一个桌面端 AI 助手会话启动器。应用以紧凑的左侧侧边栏运行，选择项目、Git 工作树、终端和 AI 助手后，即可在偏好的终端中打开会话。

## 功能

- 项目与 Git 工作树上下文切换
- 内置 Claude Code、Codex、OpenCode 启动器
- 支持自定义 AI 助手和自定义终端配置
- 支持会话聚焦、关闭、筛选和失败启动保留
- 支持置顶、失焦隐藏、左侧边缘呼出和自定义呼出快捷键
- 支持系统、明亮、深色外观，以及中文和英文语言

## 技术栈

- Wails v3 alpha
- Go 后端
- Vue 3 + TypeScript
- Vite Plus
- Tailwind CSS
- shadcn-vue
- vue-i18n
- pnpm

## 环境要求

- Go
- Wails CLI
- Node.js 与 pnpm
- Git
- Wails 所需的平台运行时依赖

如需安装 Wails CLI：

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha.91
```

## 开发

安装前端依赖：

```sh
cd frontend
pnpm install
```

启动开发模式：

```sh
wails3 task dev
```

运行检查：

```sh
wails3 generate bindings -clean=true -ts ./...
go test ./...
go build ./...
cd frontend && pnpm build
```

构建桌面应用：

```sh
wails3 task build
```

创建平台安装包：

```sh
wails3 task package
```

## 脚本

```sh
scripts/dev.sh
scripts/release.sh <version>
```

`dev.sh` 会安装并构建前端、运行 Go 测试和 Wails v3 build。`release.sh` 会构建发布二进制；配置 `SIWAP_CODESIGN_IDENTITY` 后可签名 macOS 产物。
