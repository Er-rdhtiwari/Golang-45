package main

import (
    "log"

    "github.com/Er-rdhtiwari/Golang-45/slack-tekton-notifier/internal/app"
    "github.com/Er-rdhtiwari/Golang-45/slack-tekton-notifier/internal/config"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    application := app.New(cfg)

    if err := application.Run(); err != nil {
        log.Fatalf("application failed: %v", err)
    }
}