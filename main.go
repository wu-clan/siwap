package main

import (
	"embed"
	"log"

	"siwap/internal/desktop"
)

var version = "dev"

//go:embed all:frontend/dist
var assets embed.FS

// main 初始化版本号并启动桌面应用
func main() {
	desktop.Version = version
	if err := desktop.Run(assets); err != nil {
		log.Fatal(err)
	}
}
