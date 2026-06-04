package app

import (
	"fmt"

	"github.com/Er-rdhtiwari/Golang-45/slack-tekton-notifier/internal/config"
)

type App struct {
	cfg config.Config
}

func New(cfg config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run() error {
	fmt.Println("Notifier starting...")
	fmt.Println("Namespace:", a.cfg.Kubernetes.Namespace)
	fmt.Println("PipelineRun:", a.cfg.Kubernetes.PipelineRunName)
	fmt.Println("Slack channel:", a.cfg.Slack.Channel)
	fmt.Println("Log level:", a.cfg.Runtime.LogLevel)
	fmt.Println("Slack webhook configured:", a.cfg.Slack.WebhookURL != "")

	// Later:
	// 1. Read Tekton PipelineRun status
	// 2. If failed, collect Kubernetes failure trace
	// 3. Format Slack message
	// 4. Send Slack notification

	return nil
}
