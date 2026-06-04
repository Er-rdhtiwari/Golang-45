# Day 2 — Go CLI Entrypoint, Config Loading, Flags, Environment Variables, and Project Setup

Today we focus on something every backend/cloud project needs:

> How does user input enter your Go program, become validated configuration, and get passed cleanly into application logic?

For your Slack/Tekton/Kubernetes notifier project, this is the stage where the CLI receives values like:

```bash
go run ./cmd/notifier \
  --pipeline-run failed-build-123 \
  --namespace dev \
  --slack-channel "#alerts"
```

or from environment variables:

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."
export K8S_NAMESPACE="dev"
```

Then your program converts those raw inputs into a structured `Config` object.

---

# 1. `package main` and `func main`

Every executable Go program starts with:

```go
package main
```

and must contain:

```go
func main() {
    // program starts here
}
```

## What is `package main`?

In Go, code is organized into packages.

A package can be:

```go
package config
package slack
package router
package model
package main
```

But only `package main` can build into an executable binary.

Example:

```go
package main

import "fmt"

func main() {
    fmt.Println("Slack Tekton notifier starting...")
}
```

When you run:

```bash
go run main.go
```

Go looks for `func main()` inside `package main`.

---

## Python comparison

In Python, you often write:

```python
def main():
    print("Starting app")

if __name__ == "__main__":
    main()
```

In Go, this becomes:

```go
package main

func main() {
    // start app
}
```

Go does not need `if __name__ == "__main__"`.

---

# 2. What should `main.go` do?

A beginner mistake is putting everything inside `main.go`.

Bad example:

```go
func main() {
    // parse flags
    // read env vars
    // validate config
    // call Kubernetes
    // call Slack
    // format messages
    // handle shell scripts
    // log everything
}
```

This becomes hard to test and maintain.

A better production-style `main.go` is small:

```go
func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("config error: %v", err)
    }

    app := app.New(cfg)

    if err := app.Run(); err != nil {
        log.Fatalf("application error: %v", err)
    }
}
```

The job of `main.go` is mostly:

1. Start the program.
2. Load configuration.
3. Initialize dependencies.
4. Run the application.
5. Handle fatal errors.

That is it.

---

# 3. `go.mod` and `go.sum`

## What is `go.mod`?

`go.mod` defines your Go module.

Create it with:

```bash
go mod init github.com/yourname/slack-tekton-notifier
```

Example `go.mod`:

```go
module github.com/yourname/slack-tekton-notifier

go 1.22

require (
    github.com/slack-go/slack v0.12.5
)
```

The module name is your project’s import path.

So if your module is:

```go
module github.com/yourname/slack-tekton-notifier
```

Then inside your project you can import your own packages like:

```go
import "github.com/yourname/slack-tekton-notifier/internal/config"
```

---

## What is `go.sum`?

`go.sum` stores checksums of downloaded dependencies.

It helps Go verify that dependencies have not been changed unexpectedly.

You usually do not edit `go.sum` manually.

Commands like these update it:

```bash
go mod tidy
go test ./...
go run ./cmd/notifier
```

---

## Python comparison

Python has files like:

```text
requirements.txt
pyproject.toml
poetry.lock
Pipfile.lock
```

Go has:

```text
go.mod
go.sum
```

Rough comparison:

| Python                         | Go                      |
| ------------------------------ | ----------------------- |
| `requirements.txt`             | `go.mod`                |
| `poetry.lock` / `Pipfile.lock` | `go.sum`                |
| `pip install`                  | `go mod tidy`           |
| import from package            | import from module path |

---

# 4. Imports and module names

A simple Go file:

```go
package main

import (
    "fmt"
    "log"

    "github.com/yourname/slack-tekton-notifier/internal/config"
)
```

Imports can be:

## Standard library imports

```go
import "fmt"
import "os"
import "flag"
import "log"
```

These come with Go.

## Third-party imports

```go
import "github.com/slack-go/slack"
```

These come from external modules.

## Local project imports

```go
import "github.com/yourname/slack-tekton-notifier/internal/config"
```

These refer to packages inside your own module.

---

# 5. Recommended project setup

For your notifier project, a clean structure could look like this:

```text
slack-tekton-notifier/
├── go.mod
├── go.sum
├── cmd/
│   └── notifier/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── app/
│   │   └── app.go
│   ├── slack/
│   │   └── client.go
│   ├── tekton/
│   │   └── reader.go
│   ├── router/
│   │   └── router.go
│   └── model/
│       └── event.go
├── scripts/
│   └── collect_failure_trace.sh
└── README.md
```

Important folders:

```text
cmd/
```

Contains executable entrypoints.

```text
internal/
```

Contains application packages that should not be imported by other external projects.

```text
internal/config/
```

Responsible for loading and validating configuration.

```text
internal/app/
```

Coordinates the high-level application flow.

---

# 6. CLI flags in Go

Go has a standard library package called `flag`.

Example:

```go
package main

import (
    "flag"
    "fmt"
)

func main() {
    namespace := flag.String("namespace", "default", "Kubernetes namespace")
    pipelineRun := flag.String("pipeline-run", "", "Tekton PipelineRun name")
    slackChannel := flag.String("slack-channel", "#alerts", "Slack channel name")

    flag.Parse()

    fmt.Println("Namespace:", *namespace)
    fmt.Println("PipelineRun:", *pipelineRun)
    fmt.Println("Slack channel:", *slackChannel)
}
```

Run it:

```bash
go run main.go --namespace dev --pipeline-run failed-build-123 --slack-channel "#deployments"
```

Output:

```text
Namespace: dev
PipelineRun: failed-build-123
Slack channel: #deployments
```

---

## Important detail: pointers

This line:

```go
namespace := flag.String("namespace", "default", "Kubernetes namespace")
```

returns a pointer:

```go
*string
```

So you access the value using:

```go
*namespace
```

Why?

Because the flag package stores values and updates them after `flag.Parse()`.

---

## Python comparison

Python with `argparse`:

```python
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("--namespace", default="default")
parser.add_argument("--pipeline-run", required=True)

args = parser.parse_args()

print(args.namespace)
print(args.pipeline_run)
```

Go equivalent:

```go
namespace := flag.String("namespace", "default", "Kubernetes namespace")
pipelineRun := flag.String("pipeline-run", "", "Tekton PipelineRun name")

flag.Parse()

fmt.Println(*namespace)
fmt.Println(*pipelineRun)
```

---

# 7. Environment variables

Environment variables are commonly used for deployment-specific values.

Examples:

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
export K8S_NAMESPACE="dev"
export LOG_LEVEL="debug"
```

In Go, use the `os` package:

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    webhookURL := os.Getenv("SLACK_WEBHOOK_URL")

    if webhookURL == "" {
        fmt.Println("SLACK_WEBHOOK_URL is not set")
        return
    }

    fmt.Println("Slack webhook URL loaded")
}
```

---

## `os.Getenv` vs `os.LookupEnv`

`os.Getenv` returns an empty string if the variable is missing.

```go
value := os.Getenv("LOG_LEVEL")
```

But this cannot tell the difference between:

```bash
LOG_LEVEL=""
```

and:

```text
LOG_LEVEL is not set
```

For stricter checking, use:

```go
value, ok := os.LookupEnv("LOG_LEVEL")
if !ok {
    fmt.Println("LOG_LEVEL was not set")
}
```

---

# 8. Flags vs environment variables

A production CLI often supports both.

Example:

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."
go run ./cmd/notifier --namespace dev --pipeline-run failed-build-123
```

Good rule:

| Input type            | Best for                            |
| --------------------- | ----------------------------------- |
| CLI flags             | Values that change per run          |
| Environment variables | Secrets and deployment-level config |
| Config files          | Larger structured config            |
| Defaults              | Safe fallback values                |

For your notifier:

| Value             | Recommended source           |
| ----------------- | ---------------------------- |
| PipelineRun name  | CLI flag                     |
| Namespace         | CLI flag or env              |
| Slack webhook URL | Environment variable         |
| Slack channel     | CLI flag or env              |
| Log level         | Environment variable or flag |
| Script path       | Default or flag              |

---

# 9. ASCII flow: user input → flags/env → config → app logic

```text
+------------------+
| User / Cron / CI |
+------------------+
          |
          v
+-----------------------------+
| CLI command                 |
|                             |
| ./notifier                  |
| --pipeline-run build-123    |
| --namespace dev             |
+-----------------------------+
          |
          v
+-----------------------------+
| Flags                       |
| - pipeline-run              |
| - namespace                 |
| - slack-channel             |
+-----------------------------+
          |
          v
+-----------------------------+
| Environment Variables       |
| - SLACK_WEBHOOK_URL         |
| - LOG_LEVEL                 |
| - KUBECONFIG                |
+-----------------------------+
          |
          v
+-----------------------------+
| Config Loader               |
| - reads flags               |
| - reads env vars            |
| - applies defaults          |
| - validates required values |
+-----------------------------+
          |
          v
+-----------------------------+
| Structured Config           |
|                             |
| Config{                     |
|   Namespace: "dev",         |
|   PipelineRun: "build-123", |
|   SlackWebhookURL: "...",   |
| }                           |
+-----------------------------+
          |
          v
+-----------------------------+
| App Logic                   |
| - read Tekton failure       |
| - collect Kubernetes trace  |
| - format Slack message      |
| - send Slack notification   |
+-----------------------------+
```

---

# 10. Config loader design

Instead of reading flags and env vars everywhere, create one config package.

Bad approach:

```go
func sendSlackMessage() {
    webhook := os.Getenv("SLACK_WEBHOOK_URL")
    // send message
}

func readTektonFailure() {
    namespace := os.Getenv("K8S_NAMESPACE")
    // read Kubernetes
}
```

This spreads configuration across the codebase.

Better approach:

```go
type Config struct {
    Namespace       string
    PipelineRunName string
    SlackWebhookURL string
    SlackChannel    string
    LogLevel        string
    ScriptPath      string
}
```

Then load it once:

```go
cfg, err := config.LoadConfig()
```

Then pass it into your app:

```go
application := app.New(cfg)
application.Run()
```

---

# 11. Reusable abstraction: `Config` struct and `LoadConfig`

Create:

```text
internal/config/config.go
```

Example:

```go
package config

import (
    "errors"
    "flag"
    "os"
)

type Config struct {
    Namespace       string
    PipelineRunName string
    SlackWebhookURL string
    SlackChannel    string
    LogLevel        string
    ScriptPath      string
}

func LoadConfig() (Config, error) {
    namespace := flag.String("namespace", getEnvOrDefault("K8S_NAMESPACE", "default"), "Kubernetes namespace")
    pipelineRun := flag.String("pipeline-run", "", "Tekton PipelineRun name")
    slackChannel := flag.String("slack-channel", getEnvOrDefault("SLACK_CHANNEL", "#alerts"), "Slack channel")
    logLevel := flag.String("log-level", getEnvOrDefault("LOG_LEVEL", "info"), "Log level")
    scriptPath := flag.String("script-path", getEnvOrDefault("TRACE_SCRIPT_PATH", "./scripts/collect_failure_trace.sh"), "Failure trace script path")

    flag.Parse()

    cfg := Config{
        Namespace:       *namespace,
        PipelineRunName: *pipelineRun,
        SlackWebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
        SlackChannel:    *slackChannel,
        LogLevel:        *logLevel,
        ScriptPath:      *scriptPath,
    }

    if err := cfg.Validate(); err != nil {
        return Config{}, err
    }

    return cfg, nil
}

func (c Config) Validate() error {
    if c.PipelineRunName == "" {
        return errors.New("pipeline-run is required")
    }

    if c.SlackWebhookURL == "" {
        return errors.New("SLACK_WEBHOOK_URL environment variable is required")
    }

    if c.Namespace == "" {
        return errors.New("namespace cannot be empty")
    }

    if c.SlackChannel == "" {
        return errors.New("slack channel cannot be empty")
    }

    return nil
}

func getEnvOrDefault(key string, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}
```

---

# 12. Example `main.go`

Create:

```text
cmd/notifier/main.go
```

```go
package main

import (
    "log"

    "github.com/yourname/slack-tekton-notifier/internal/app"
    "github.com/yourname/slack-tekton-notifier/internal/config"
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
```

Notice how small this is.

`main.go` does not know how Slack works.

`main.go` does not know how Tekton works.

`main.go` does not know how shell scripts work.

It only wires things together.

---

# 13. Example `app.go`

Create:

```text
internal/app/app.go
```

```go
package app

import (
    "fmt"

    "github.com/yourname/slack-tekton-notifier/internal/config"
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
    fmt.Println("Namespace:", a.cfg.Namespace)
    fmt.Println("PipelineRun:", a.cfg.PipelineRunName)
    fmt.Println("Slack channel:", a.cfg.SlackChannel)

    // Later:
    // 1. Read Tekton PipelineRun status
    // 2. If failed, collect Kubernetes failure trace
    // 3. Format Slack message
    // 4. Send Slack notification

    return nil
}
```

This creates a clean separation:

```text
main.go      -> starts the app
config.go    -> loads and validates config
app.go       -> coordinates business workflow
```

---

# 14. Pseudocode for config loading

```text
FUNCTION LoadConfig:

    read namespace from:
        CLI flag --namespace
        fallback env K8S_NAMESPACE
        fallback default "default"

    read pipeline run from:
        CLI flag --pipeline-run
        no default

    read Slack webhook URL from:
        env SLACK_WEBHOOK_URL

    read Slack channel from:
        CLI flag --slack-channel
        fallback env SLACK_CHANNEL
        fallback default "#alerts"

    read log level from:
        CLI flag --log-level
        fallback env LOG_LEVEL
        fallback default "info"

    create Config struct

    validate Config:
        pipeline run must not be empty
        Slack webhook URL must not be empty
        namespace must not be empty
        Slack channel must not be empty

    return Config
```

---

# 15. Default values and validation

Defaults are useful for non-critical values.

Example:

```go
namespace := getEnvOrDefault("K8S_NAMESPACE", "default")
logLevel := getEnvOrDefault("LOG_LEVEL", "info")
slackChannel := getEnvOrDefault("SLACK_CHANNEL", "#alerts")
```

But not every value should have a default.

For example, this should not have a fake default:

```go
SLACK_WEBHOOK_URL
```

Bad:

```go
webhook := getEnvOrDefault("SLACK_WEBHOOK_URL", "dummy")
```

Better:

```go
webhook := os.Getenv("SLACK_WEBHOOK_URL")
if webhook == "" {
    return errors.New("SLACK_WEBHOOK_URL is required")
}
```

---

## Good defaults

```text
namespace       -> default
log level       -> info
Slack channel   -> #alerts
script path     -> ./scripts/collect_failure_trace.sh
timeout         -> 30 seconds
```

## Required values

```text
Slack webhook URL
PipelineRun name
Kubernetes access credentials
```

---

# 16. Why config should be separate from business logic

Business logic should answer:

```text
What should the application do?
```

Config logic should answer:

```text
Where do values come from?
```

These should not be mixed.

Bad:

```go
func SendSlackAlert(message string) error {
    webhookURL := os.Getenv("SLACK_WEBHOOK_URL")

    if webhookURL == "" {
        return errors.New("missing webhook")
    }

    // send Slack message
}
```

Better:

```go
type SlackClient struct {
    webhookURL string
}

func NewSlackClient(webhookURL string) *SlackClient {
    return &SlackClient{webhookURL: webhookURL}
}
```

Then:

```go
client := slack.NewClient(cfg.SlackWebhookURL)
```

Now your Slack client does not care whether the webhook came from:

```text
environment variable
secret manager
config file
Kubernetes Secret
CI/CD variable
```

It just receives the value.

---

# 17. Secrets vs normal config

This is very important in production.

## Normal config

Normal config is not sensitive.

Examples:

```text
K8S_NAMESPACE=dev
LOG_LEVEL=info
SLACK_CHANNEL=#alerts
TRACE_SCRIPT_PATH=./scripts/collect_failure_trace.sh
```

These can usually appear in logs, config files, or documentation.

## Secrets

Secrets are sensitive.

Examples:

```text
SLACK_WEBHOOK_URL
SLACK_BOT_TOKEN
KUBECONFIG credentials
API keys
database passwords
OAuth tokens
```

Do not print secrets.

Bad:

```go
fmt.Println("Slack webhook:", cfg.SlackWebhookURL)
```

Better:

```go
fmt.Println("Slack webhook configured:", cfg.SlackWebhookURL != "")
```

---

## Production-style rule

You can log:

```text
Slack webhook configured: true
```

Do not log:

```text
https://hooks.slack.com/services/actual-secret-value
```

---

# 18. How CLI input becomes structured application input

Raw CLI command:

```bash
./notifier --namespace dev --pipeline-run failed-build-123
```

Raw environment:

```bash
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
SLACK_CHANNEL=#build-alerts
LOG_LEVEL=debug
```

Becomes:

```go
Config{
    Namespace:       "dev",
    PipelineRunName: "failed-build-123",
    SlackWebhookURL: "https://hooks.slack.com/services/...",
    SlackChannel:    "#build-alerts",
    LogLevel:        "debug",
    ScriptPath:      "./scripts/collect_failure_trace.sh",
}
```

Then the rest of the app receives one clean object:

```go
app.New(cfg)
```

This is much better than passing many separate strings everywhere.

---

# 19. Production-style config design

A more mature config could look like this:

```go
package config

import "time"

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
```

This is cleaner for larger apps.

Usage:

```go
cfg.Slack.WebhookURL
cfg.Kubernetes.Namespace
cfg.Runtime.LogLevel
```

Instead of:

```go
cfg.SlackWebhookURL
cfg.K8sNamespace
cfg.LogLevel
```

---

## Example production-style loader

```go
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
```

---

# 20. Designing a small but production-like CLI entrypoint

Your CLI entrypoint should be boring.

That is a good thing.

A production-like `main.go` follows this shape:

```go
func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("config error: %v", err)
    }

    application, err := app.New(cfg)
    if err != nil {
        log.Fatalf("failed to initialize app: %v", err)
    }

    if err := application.Run(); err != nil {
        log.Fatalf("app failed: %v", err)
    }
}
```

This pattern appears in many backend services too.

A backend service might do:

```text
load config
connect to database
create router
start HTTP server
```

Your CLI does:

```text
load config
create Slack client
create Tekton reader
run one notification workflow
exit
```

The shape is similar.

---

# 21. CLI-based vs service-like behavior

Your project is CLI-based because:

```text
It starts, performs one job, then exits.
```

Example:

```bash
./notifier --pipeline-run failed-build-123
```

But it resembles a service because internally it has:

```text
configuration
routing
clients
models
validation
logging
external integrations
failure handling
```

A backend service usually keeps running:

```text
start HTTP server
wait for requests
process many events
```

Your CLI processes one event:

```text
start
read config
process one PipelineRun
send Slack message
exit
```

But the internal architecture can still be production-like.

---

# 22. Hands-on: build/refactor config loader

Your task today:

Create this file:

```text
internal/config/config.go
```

Add:

```go
package config

import (
    "errors"
    "flag"
    "os"
)

type Config struct {
    Namespace       string
    PipelineRunName string
    SlackWebhookURL string
    SlackChannel    string
    LogLevel        string
}

func LoadConfig() (Config, error) {
    namespace := flag.String("namespace", getEnvOrDefault("K8S_NAMESPACE", "default"), "Kubernetes namespace")
    pipelineRun := flag.String("pipeline-run", "", "Tekton PipelineRun name")
    slackChannel := flag.String("slack-channel", getEnvOrDefault("SLACK_CHANNEL", "#alerts"), "Slack channel")
    logLevel := flag.String("log-level", getEnvOrDefault("LOG_LEVEL", "info"), "Log level")

    flag.Parse()

    cfg := Config{
        Namespace:       *namespace,
        PipelineRunName: *pipelineRun,
        SlackWebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
        SlackChannel:    *slackChannel,
        LogLevel:        *logLevel,
    }

    if err := cfg.Validate(); err != nil {
        return Config{}, err
    }

    return cfg, nil
}

func (c Config) Validate() error {
    if c.Namespace == "" {
        return errors.New("namespace is required")
    }

    if c.PipelineRunName == "" {
        return errors.New("pipeline-run is required")
    }

    if c.SlackWebhookURL == "" {
        return errors.New("SLACK_WEBHOOK_URL is required")
    }

    if c.SlackChannel == "" {
        return errors.New("slack channel is required")
    }

    return nil
}

func getEnvOrDefault(key, fallback string) string {
    value := os.Getenv(key)
    if value == "" {
        return fallback
    }
    return value
}
```

Then create:

```text
cmd/notifier/main.go
```

```go
package main

import (
    "fmt"
    "log"

    "github.com/yourname/slack-tekton-notifier/internal/config"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    fmt.Println("Config loaded successfully")
    fmt.Println("Namespace:", cfg.Namespace)
    fmt.Println("PipelineRun:", cfg.PipelineRunName)
    fmt.Println("Slack channel:", cfg.SlackChannel)
    fmt.Println("Log level:", cfg.LogLevel)
    fmt.Println("Slack webhook configured:", cfg.SlackWebhookURL != "")
}
```

Run:

```bash
export SLACK_WEBHOOK_URL="dummy-webhook-for-local-test"

go run ./cmd/notifier \
  --namespace dev \
  --pipeline-run failed-build-123 \
  --slack-channel "#build-alerts"
```

Expected output:

```text
Config loaded successfully
Namespace: dev
PipelineRun: failed-build-123
Slack channel: #build-alerts
Log level: info
Slack webhook configured: true
```

---

# 23. Testing missing required values

Run without `pipeline-run`:

```bash
export SLACK_WEBHOOK_URL="dummy"

go run ./cmd/notifier --namespace dev
```

Expected error:

```text
failed to load config: pipeline-run is required
```

Run without Slack webhook:

```bash
unset SLACK_WEBHOOK_URL

go run ./cmd/notifier --namespace dev --pipeline-run failed-build-123
```

Expected error:

```text
failed to load config: SLACK_WEBHOOK_URL is required
```

This is good behavior.

Your app should fail early before trying to contact Kubernetes or Slack.

---

# 24. Common mistakes

## Mistake 1: Putting all logic in `main.go`

Bad:

```go
func main() {
    // config
    // Slack
    // Kubernetes
    // formatting
    // shell execution
}
```

Better:

```go
func main() {
    cfg, err := config.LoadConfig()
    // initialize app
    // run app
}
```

---

## Mistake 2: Logging secrets

Bad:

```go
log.Println(cfg.SlackWebhookURL)
```

Better:

```go
log.Println("Slack webhook configured:", cfg.SlackWebhookURL != "")
```

---

## Mistake 3: No validation

Bad:

```go
cfg := Config{
    PipelineRunName: *pipelineRun,
}
return cfg, nil
```

Better:

```go
if err := cfg.Validate(); err != nil {
    return Config{}, err
}
```

---

## Mistake 4: Too many global variables

Bad:

```go
var Namespace string
var SlackWebhookURL string
var PipelineRun string
```

Better:

```go
type Config struct {
    Namespace       string
    SlackWebhookURL string
    PipelineRunName string
}
```

---

## Mistake 5: Reading environment variables deep inside business logic

Bad:

```go
func NotifySlack() {
    webhook := os.Getenv("SLACK_WEBHOOK_URL")
}
```

Better:

```go
func NewSlackClient(webhookURL string) *SlackClient {
    return &SlackClient{webhookURL: webhookURL}
}
```

---

## Mistake 6: Using defaults for secrets

Bad:

```go
webhook := getEnvOrDefault("SLACK_WEBHOOK_URL", "test")
```

Better:

```go
webhook := os.Getenv("SLACK_WEBHOOK_URL")
if webhook == "" {
    return errors.New("SLACK_WEBHOOK_URL is required")
}
```

---

# 25. Debugging checklist

When your CLI does not work, check these:

## Command correctness

```bash
go run ./cmd/notifier --namespace dev --pipeline-run failed-build-123
```

Check spelling:

```text
--pipeline-run
```

not:

```text
--pipelinerun
--pipeline
--pipeline_run
```

---

## Environment variables

Check:

```bash
echo $SLACK_WEBHOOK_URL
echo $SLACK_CHANNEL
echo $K8S_NAMESPACE
```

---

## Module path

Check `go.mod`:

```go
module github.com/yourname/slack-tekton-notifier
```

Your import must match:

```go
import "github.com/yourname/slack-tekton-notifier/internal/config"
```

---

## Run from project root

Prefer:

```bash
go run ./cmd/notifier
```

from the root directory.

Avoid running randomly from inside nested folders until you understand module paths well.

---

## Dependencies

Run:

```bash
go mod tidy
```

---

## Compile everything

Run:

```bash
go test ./...
```

Even before writing tests, this checks that packages compile.

---

# 26. Go syntax and conventions compared with Python

## Variables

Python:

```python
namespace = "dev"
```

Go:

```go
namespace := "dev"
```

or:

```go
var namespace string = "dev"
```

---

## Structs vs dictionaries/classes

Python dictionary:

```python
config = {
    "namespace": "dev",
    "pipeline_run": "build-123",
}
```

Go struct:

```go
type Config struct {
    Namespace       string
    PipelineRunName string
}
```

Usage:

```go
cfg := Config{
    Namespace:       "dev",
    PipelineRunName: "build-123",
}
```

---

## Error handling

Python often uses exceptions:

```python
if not pipeline_run:
    raise ValueError("pipeline_run is required")
```

Go returns errors:

```go
if c.PipelineRunName == "" {
    return errors.New("pipeline-run is required")
}
```

Usage:

```go
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal(err)
}
```

---

## Packages

Python:

```python
from config.loader import load_config
```

Go:

```go
import "github.com/yourname/slack-tekton-notifier/internal/config"
```

Then:

```go
cfg, err := config.LoadConfig()
```

---

## Public vs private names

In Go, capitalization matters.

Public/exported:

```go
type Config struct {}
func LoadConfig() {}
```

Private/unexported:

```go
func getEnvOrDefault() {}
```

Python uses naming convention:

```python
def _get_env_or_default():
    pass
```

Go enforces this through capitalization.

---

# 27. How this fits your Slack/Tekton/Kubernetes notifier

Your Day 1 architecture had this flow:

```text
CLI -> config -> model -> router -> Slack client -> shell script -> Tekton -> Kubernetes -> failure trace -> Slack message
```

Today focuses on the beginning:

```text
CLI -> config
```

That part is small but very important.

If configuration is messy, every later layer becomes messy.

A clean config loader gives later packages a stable contract:

```go
type Config struct {
    Namespace       string
    PipelineRunName string
    SlackWebhookURL string
    SlackChannel    string
}
```

Now the rest of your system does not care whether values came from:

```text
CLI flags
environment variables
Kubernetes Secret
GitHub Actions variables
IBM Cloud Code Engine env vars
```

The rest of the system only receives clean application input.

---

# 28. Interview questions with sample answers

## 1. What is the role of `main.go` in a Go application?

Sample answer:

`main.go` is the executable entrypoint. It should usually stay small. Its job is to load configuration, initialize dependencies, call the application runner, and handle fatal errors. Business logic should live in separate packages so it can be tested and reused.

---

## 2. What is the difference between `go.mod` and `go.sum`?

Sample answer:

`go.mod` defines the module name, Go version, and dependency requirements. `go.sum` stores checksums for dependencies to ensure reproducible and secure builds. Developers usually edit `go.mod` indirectly through commands like `go get` and `go mod tidy`, and they generally do not manually edit `go.sum`.

---

## 3. When would you use CLI flags versus environment variables?

Sample answer:

CLI flags are useful for values that change per execution, such as a PipelineRun name or namespace. Environment variables are useful for deployment-level settings and secrets, such as Slack webhook URLs or API tokens. A good config loader can combine both, applying defaults and validation.

---

## 4. Why should configuration be separate from business logic?

Sample answer:

Separating configuration from business logic makes the application easier to test, maintain, and deploy. Business logic should not know whether a value came from a flag, environment variable, config file, or secret manager. It should receive a validated config object and focus only on application behavior.

---

## 5. How would you prevent secrets from leaking in logs?

Sample answer:

I would avoid printing secret values directly. Instead of logging the Slack webhook URL or API token, I would log whether it is configured. For example, `Slack webhook configured: true`. I would also keep secrets in environment variables, Kubernetes Secrets, or a secret manager rather than hardcoding them.

---

# 29. Mini assignment for Day 2

Build this:

```text
internal/config/config.go
cmd/notifier/main.go
```

Your CLI should support:

```bash
--namespace
--pipeline-run
--slack-channel
--log-level
```

Your environment variables should support:

```bash
SLACK_WEBHOOK_URL
K8S_NAMESPACE
SLACK_CHANNEL
LOG_LEVEL
```

Validation rules:

```text
pipeline-run is required
SLACK_WEBHOOK_URL is required
namespace cannot be empty
slack channel cannot be empty
```

Run test commands:

```bash
export SLACK_WEBHOOK_URL="dummy"

go run ./cmd/notifier \
  --namespace dev \
  --pipeline-run failed-build-123 \
  --slack-channel "#alerts"
```

Expected result:

```text
Config loaded successfully
```

Then test failure:

```bash
unset SLACK_WEBHOOK_URL

go run ./cmd/notifier \
  --namespace dev \
  --pipeline-run failed-build-123
```

Expected result:

```text
failed to load config: SLACK_WEBHOOK_URL is required
```

---

# 30. Key takeaway

A production-like Go CLI should not begin with business logic.

It should begin with clean input handling:

```text
raw flags/env vars
        |
        v
validated Config struct
        |
        v
application logic
```

Your goal is to make the rest of the application depend on this:

```go
cfg config.Config
```

not this:

```go
os.Getenv(...)
flag.String(...)
hardcoded values
global variables
```

That is the first step toward writing Go code that feels like real backend/cloud production code.
# DSA — Strings Basics in Go

Today we’ll revise **strings in Go** from a DSA/interview perspective.

Strings look simple, but in Go there are a few important details:

```go
s := "hello"
```

A Go string is:

```text
immutable sequence of bytes
```

That means:

1. A string cannot be changed in-place.
2. `len(s)` gives the number of bytes, not always the number of characters.
3. For normal English strings, bytes and characters usually match.
4. For Unicode strings like `"नमस्ते"` or `"😊"`, bytes and characters are different.

---

# 1. Basic string declaration

```go
package main

import "fmt"

func main() {
    name := "golang"

    fmt.Println(name)
}
```

Output:

```text
golang
```

Python comparison:

```python
name = "golang"
print(name)
```

Go uses:

```go
name := "golang"
```

Python uses:

```python
name = "golang"
```

---

# 2. Strings are immutable

In Python, strings are also immutable.

Python:

```python
s = "hello"
# s[0] = "H"  # error
```

Go:

```go
s := "hello"
// s[0] = 'H' // error
```

You cannot modify a character directly inside a string.

Bad Go:

```go
s := "hello"
s[0] = 'H' // not allowed
```

Why?

Because strings are immutable.

To modify a string, convert it to a mutable structure first.

For ASCII strings, you can use `[]byte`:

```go
package main

import "fmt"

func main() {
    s := "hello"

    chars := []byte(s)
    chars[0] = 'H'

    result := string(chars)

    fmt.Println(result)
}
```

Output:

```text
Hello
```

For Unicode-safe modification, use `[]rune`:

```go
s := "नमस्ते"
runes := []rune(s)
```

---

# 3. Indexing strings

You can access a string by index:

```go
package main

import "fmt"

func main() {
    s := "hello"

    fmt.Println(s[0])
    fmt.Println(s[1])
}
```

Output:

```text
104
101
```

This surprises beginners.

Why not `h` and `e`?

Because `s[0]` gives a **byte**, and Go prints byte values as numbers by default.

ASCII values:

```text
h -> 104
e -> 101
l -> 108
o -> 111
```

To print as a character:

```go
fmt.Printf("%c\n", s[0])
fmt.Printf("%c\n", s[1])
```

Output:

```text
h
e
```

Python comparison:

```python
s = "hello"
print(s[0])  # h
```

In Python, indexing returns a one-character string.

In Go, indexing returns a byte.

---

# 4. String length

```go
s := "hello"
fmt.Println(len(s))
```

Output:

```text
5
```

For English letters, this is simple.

But remember:

```go
s := "😊"
fmt.Println(len(s))
```

Output is usually:

```text
4
```

Why?

Because `len(s)` returns bytes, and the emoji uses multiple bytes in UTF-8.

To count Unicode characters, use:

```go
package main

import (
    "fmt"
    "unicode/utf8"
)

func main() {
    s := "😊"

    fmt.Println(len(s))
    fmt.Println(utf8.RuneCountInString(s))
}
```

Output:

```text
4
1
```

Python comparison:

```python
s = "😊"
print(len(s))  # 1
```

Important interview point:

```text
Go len(string) = number of bytes
Python len(string) = number of characters
```

---

# 5. Bytes and runes

Go has two important types for string problems:

```go
byte
rune
```

## `byte`

A `byte` is an alias for `uint8`.

Use it for ASCII-style problems:

```text
a-z
A-Z
0-9
simple English strings
```

Example:

```go
s := "abc"
b := s[0]

fmt.Printf("%c\n", b)
```

---

## `rune`

A `rune` represents a Unicode code point.

Use it when the string may contain Unicode characters.

Example:

```go
s := "हेलो"

for i, ch := range s {
    fmt.Println(i, ch)
}
```

Better print:

```go
for i, ch := range s {
    fmt.Printf("byte index: %d, character: %c\n", i, ch)
}
```

---

# 6. Looping through strings

There are two common ways.

## Method 1: index loop

```go
s := "hello"

for i := 0; i < len(s); i++ {
    fmt.Printf("%c\n", s[i])
}
```

This loops byte by byte.

Good for ASCII.

---

## Method 2: range loop

```go
s := "hello"

for i, ch := range s {
    fmt.Println(i, string(ch))
}
```

This loops rune by rune.

Good for Unicode.

For DSA beginner problems with lowercase English letters, the byte loop is usually fine.

---

# 7. Concatenating strings

```go
first := "hello"
second := "world"

result := first + " " + second

fmt.Println(result)
```

Output:

```text
hello world
```

Python comparison:

```python
result = first + " " + second
```

Same idea.

---

## Important DSA note

Repeated string concatenation inside a loop can be inefficient.

Bad for large strings:

```go
result := ""

for i := 0; i < len(s); i++ {
    result += string(s[i])
}
```

Why?

Because strings are immutable. Every concatenation creates a new string.

Better:

```go
var builder strings.Builder

for i := 0; i < len(s); i++ {
    builder.WriteByte(s[i])
}

result := builder.String()
```

Need import:

```go
import "strings"
```

Python comparison:

```python
parts = []

for ch in s:
    parts.append(ch)

result = "".join(parts)
```

Go equivalent:

```go
var builder strings.Builder
builder.WriteByte(ch)
result := builder.String()
```

---

# 8. Common string functions

Go has the `strings` package.

```go
import "strings"
```

Examples:

```go
strings.Contains("hello", "ell")      // true
strings.HasPrefix("hello", "he")      // true
strings.HasSuffix("hello", "lo")      // true
strings.ToLower("GoLang")             // "golang"
strings.ToUpper("GoLang")             // "GOLANG"
strings.TrimSpace(" hello ")          // "hello"
strings.Split("a,b,c", ",")           // []string{"a", "b", "c"}
strings.Join([]string{"a", "b"}, "-") // "a-b"
```

Python comparison:

| Python               | Go                           |
| -------------------- | ---------------------------- |
| `"ell" in s`         | `strings.Contains(s, "ell")` |
| `s.startswith("he")` | `strings.HasPrefix(s, "he")` |
| `s.endswith("lo")`   | `strings.HasSuffix(s, "lo")` |
| `s.lower()`          | `strings.ToLower(s)`         |
| `s.upper()`          | `strings.ToUpper(s)`         |
| `s.strip()`          | `strings.TrimSpace(s)`       |
| `s.split(",")`       | `strings.Split(s, ",")`      |
| `"-".join(arr)`      | `strings.Join(arr, "-")`     |

---

# 9. ASCII character checks

For many beginner DSA problems, inputs are lowercase English letters.

Example:

```go
ch := s[i]

if ch >= 'a' && ch <= 'z' {
    fmt.Println("lowercase letter")
}
```

Remember:

```go
'a'
```

is a rune/character literal.

```go
"a"
```

is a string literal.

Important difference:

```go
'a' // character
"a" // string
```

Python does not separate these strongly:

```python
'a'  # string of length 1
"a"  # also string of length 1
```

In Go:

```go
'a' // rune
"a" // string
```

---

# 10. Easy string problem

## Problem: Reverse a String

Given a string `s`, return the reversed string.

Example:

```text
Input:  "hello"
Output: "olleh"
```

Another example:

```text
Input:  "golang"
Output: "gnalog"
```

For now, assume the input contains only lowercase English letters.

---

# 11. Brute-force thinking

We can build a new string from the end to the beginning.

Pseudo-code:

```text
result = empty string

for i from len(s)-1 down to 0:
    add s[i] to result

return result
```

Go code:

```go
package main

import "fmt"

func reverseString(s string) string {
    result := ""

    for i := len(s) - 1; i >= 0; i-- {
        result += string(s[i])
    }

    return result
}

func main() {
    fmt.Println(reverseString("hello"))
}
```

Output:

```text
olleh
```

This works, but it is not ideal.

Why?

Because strings are immutable. Every time we do:

```go
result += string(s[i])
```

Go creates a new string.

Time complexity can become worse than expected for large strings.

---

# 12. Better solution using `strings.Builder`

```go
package main

import (
    "fmt"
    "strings"
)

func reverseString(s string) string {
    var builder strings.Builder

    for i := len(s) - 1; i >= 0; i-- {
        builder.WriteByte(s[i])
    }

    return builder.String()
}

func main() {
    fmt.Println(reverseString("hello"))
    fmt.Println(reverseString("golang"))
}
```

Output:

```text
olleh
gnalog
```

## Complexity

Let `n` be the length of the string.

```text
Time Complexity:  O(n)
Space Complexity: O(n)
```

We need new space because strings are immutable.

---

# 13. Unicode-safe reverse

The previous solution is fine for ASCII.

But it may break for Unicode:

```go
reverseString("😊a")
```

Because the emoji is multiple bytes.

Unicode-safe version:

```go
package main

import "fmt"

func reverseStringUnicode(s string) string {
    runes := []rune(s)

    left := 0
    right := len(runes) - 1

    for left < right {
        runes[left], runes[right] = runes[right], runes[left]
        left++
        right--
    }

    return string(runes)
}

func main() {
    fmt.Println(reverseStringUnicode("hello"))
    fmt.Println(reverseStringUnicode("😊a"))
}
```

Output:

```text
olleh
a😊
```

For interviews, mention:

```text
If the input is ASCII, []byte or strings.Builder is enough.
If the input may contain Unicode, use []rune.
```

---

# 14. Python version of the same problem

Python simple version:

```python
def reverse_string(s):
    return s[::-1]

print(reverse_string("hello"))
```

Output:

```text
olleh
```

Manual Python version:

```python
def reverse_string(s):
    result = []

    for i in range(len(s) - 1, -1, -1):
        result.append(s[i])

    return "".join(result)
```

Go equivalent:

```go
func reverseString(s string) string {
    var builder strings.Builder

    for i := len(s) - 1; i >= 0; i-- {
        builder.WriteByte(s[i])
    }

    return builder.String()
}
```

Python has very compact slicing:

```python
s[::-1]
```

Go does not have this shortcut. In Go, we usually write the loop manually.

---

# 15. Common mistakes in Go string DSA

## Mistake 1: Thinking `len(s)` always means character count

```go
fmt.Println(len("😊")) // 4, not 1
```

For Unicode character count:

```go
utf8.RuneCountInString("😊")
```

---

## Mistake 2: Trying to mutate a string directly

Bad:

```go
s := "hello"
s[0] = 'H'
```

Good:

```go
chars := []byte(s)
chars[0] = 'H'
s = string(chars)
```

---

## Mistake 3: Confusing single quotes and double quotes

```go
'a' // rune
"a" // string
```

This is valid:

```go
if s[i] == 'a' {
    fmt.Println("found a")
}
```

This is wrong:

```go
if s[i] == "a" {
    fmt.Println("found a")
}
```

Why?

Because:

```go
s[i]
```

is a byte, but:

```go
"a"
```

is a string.

---

## Mistake 4: Using `+` repeatedly for large strings

Bad:

```go
result := ""

for i := 0; i < len(s); i++ {
    result += string(s[i])
}
```

Better:

```go
var builder strings.Builder

for i := 0; i < len(s); i++ {
    builder.WriteByte(s[i])
}
```

---

# 16. Mini practice problem

## Problem

Write a function that checks whether a string is a palindrome.

A palindrome reads the same forward and backward.

Examples:

```text
"madam" -> true
"racecar" -> true
"hello" -> false
```

Try solving with two pointers:

```text
left = 0
right = len(s) - 1

while left < right:
    if s[left] != s[right]:
        return false

    left++
    right--

return true
```

Go skeleton:

```go
package main

import "fmt"

func isPalindrome(s string) bool {
    left := 0
    right := len(s) - 1

    for left < right {
        if s[left] != s[right] {
            return false
        }

        left++
        right--
    }

    return true
}

func main() {
    fmt.Println(isPalindrome("madam"))
    fmt.Println(isPalindrome("racecar"))
    fmt.Println(isPalindrome("hello"))
}
```

Expected output:

```text
true
true
false
```

Complexity:

```text
Time Complexity:  O(n)
Space Complexity: O(1)
```

This is a great beginner DSA string problem because it teaches:

```text
indexing
two pointers
byte comparison
loop conditions
early return
```

---

# 17. Key takeaway

For Go string DSA, remember this mental model:

```text
string  -> immutable bytes
byte    -> ASCII-style character handling
rune    -> Unicode-safe character handling
[]byte  -> mutable byte array
[]rune  -> mutable Unicode character array
```

Python hides many of these details.

Go makes them explicit.

That explicitness is useful in backend/cloud work because you often deal with:

```text
logs
CLI inputs
JSON payloads
environment variables
Kubernetes names
Slack messages
file paths
network data
```

So for your current Go learning path, strings are not just DSA. They are also a core backend skill.
