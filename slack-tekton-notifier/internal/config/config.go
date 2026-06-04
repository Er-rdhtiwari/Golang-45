package config

import (
    "errors"
    "flag"
    "os"
    "time"
)

type Config struct {
    Kubernetes KubernetesConfig
    Slack      SlackConfig
    Runtime    RuntimeConfig
}

type KubernetesConfig struct {
    Namespace       string
    PipelineRunName string
    KubeconfigPath  string
}

type SlackConfig struct {
    WebhookURL string
    Channel    string
}

type RuntimeConfig struct {
    LogLevel   string
    ScriptPath string
    Timeout    time.Duration
}

func LoadConfig() (Config, error) {
    namespace := flag.String("namespace", getEnvOrDefault("K8S_NAMESPACE", "default"), "Kubernetes namespace")
    pipelineRun := flag.String("pipeline-run", "", "Tekton PipelineRun name")
    kubeconfig := flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "Path to kubeconfig file")

    slackChannel := flag.String("slack-channel", getEnvOrDefault("SLACK_CHANNEL", "#alerts"), "Slack channel")
    logLevel := flag.String("log-level", getEnvOrDefault("LOG_LEVEL", "info"), "Log level")
    scriptPath := flag.String("script-path", getEnvOrDefault("TRACE_SCRIPT_PATH", "./scripts/collect_failure_trace.sh"), "Trace script path")
    timeoutSeconds := flag.Int("timeout-seconds", 30, "Application timeout in seconds")

    flag.Parse()

    cfg := Config{
        Kubernetes: KubernetesConfig{
            Namespace:       *namespace,
            PipelineRunName: *pipelineRun,
            KubeconfigPath:  *kubeconfig,
        },
        Slack: SlackConfig{
            WebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
            Channel:    *slackChannel,
        },
        Runtime: RuntimeConfig{
            LogLevel:   *logLevel,
            ScriptPath: *scriptPath,
            Timeout:    time.Duration(*timeoutSeconds) * time.Second,
        },
    }

    if err := cfg.Validate(); err != nil {
        return Config{}, err
    }

    return cfg, nil
}

func (c Config) Validate() error {
    if c.Kubernetes.PipelineRunName == "" {
        return errors.New("pipeline-run is required")
    }

    if c.Kubernetes.Namespace == "" {
        return errors.New("namespace cannot be empty")
    }

    if c.Slack.WebhookURL == "" {
        return errors.New("SLACK_WEBHOOK_URL environment variable is required")
    }

    if c.Slack.Channel == "" {
        return errors.New("Slack channel cannot be empty")
    }

    if c.Runtime.Timeout <= 0 {
        return errors.New("timeout must be greater than zero")
    }

    return nil
}

func getEnvOrDefault(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}