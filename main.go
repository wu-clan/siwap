package main

import (
	"embed"
	"log"

	"siwap/internal/desktop"
)

var version = "dev"

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	desktop.Version = version
	if err := desktop.Run(assets); err != nil {
		log.Fatal(err)
	}
}
