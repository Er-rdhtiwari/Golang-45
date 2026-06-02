# Day 1 — Slack/Tekton/Kubernetes Notifier Architecture Revision

Think of this project as a **cloud-native DevOps notifier**:

> A Go CLI tool observes or triggers a Tekton pipeline running on Kubernetes, detects success or failure, collects useful failure context, formats it, and sends a Slack message to help engineers react quickly.

It is small enough to be a learning project, but it resembles a real backend/cloud workflow because it connects configuration, domain models, routing logic, external APIs, shell execution, Kubernetes resources, error handling, logging, and notifications.

---

# 1. What problem the project solves

In a real cloud team, builds and deployments fail often because of:

* broken tests
* bad container images
* missing Kubernetes secrets
* failed Tekton Tasks
* wrong RBAC permissions
* bad YAML
* unavailable clusters
* network issues
* Slack webhook/API failures

Without a notifier, someone must manually inspect Tekton, Kubernetes pods, logs, and events.

Your notifier solves this by:

1. Running from the CLI or CI/CD environment.
2. Reading configuration.
3. Executing or observing a Tekton pipeline.
4. Detecting pipeline success or failure.
5. Pulling failure trace information from Tekton/Kubernetes.
6. Sending a clean Slack message with status and debugging context.

The goal is not just “send Slack message.”

The real goal is:

> Turn noisy infrastructure failure into a clear, actionable engineering signal.

---

# 2. Full end-to-end flow

Requested flow:

```text
CLI
 -> config
 -> model
 -> router
 -> Slack client
 -> shell script
 -> Tekton
 -> Kubernetes
 -> failure trace
 -> Slack message
```

A realistic interpretation:

1. **CLI** receives user input.
2. **Config loader** reads YAML/env/flags.
3. **Model layer** creates typed Go structs.
4. **Router** decides which workflow to run.
5. **Slack client** may send an initial “pipeline started” message.
6. **Shell script runner** calls `tkn`, `kubectl`, or another script.
7. **Tekton** creates PipelineRuns and TaskRuns.
8. **Kubernetes** schedules pods and containers.
9. **Failure tracer** inspects PipelineRuns, TaskRuns, pods, logs, and events.
10. **Slack client** sends final success/failure message.

---

# ASCII architecture diagram

```text
+----------------------+
|      User / CI        |
|  ./notifier run ...   |
+----------+-----------+
           |
           v
+----------------------+
|       main.go         |
|  small entrypoint     |
+----------+-----------+
           |
           v
+----------------------+
|      CLI Layer        |
| flags, args, command  |
+----------+-----------+
           |
           v
+----------------------+
|    Config Loader      |
| YAML + env + defaults |
+----------+-----------+
           |
           v
+----------------------+
|     Domain Model      |
| PipelineRequest       |
| NotificationEvent     |
| FailureTrace          |
+----------+-----------+
           |
           v
+----------------------+
|        Router         |
| decides workflow      |
| run / notify / trace  |
+----+-------------+---+
     |             |
     |             |
     v             v
+---------+   +----------------+
| Slack   |   | Script Runner  |
| Client  |   | bash/tkn/kubectl|
+----+----+   +--------+-------+
     |                 |
     |                 v
     |        +----------------+
     |        |     Tekton     |
     |        | PipelineRun    |
     |        | TaskRun        |
     |        +--------+-------+
     |                 |
     |                 v
     |        +----------------+
     |        |  Kubernetes    |
     |        | Pods/Logs/Events|
     |        +--------+-------+
     |                 |
     |                 v
     |        +----------------+
     |        | Failure Tracer |
     |        | collect reason |
     |        +--------+-------+
     |                 |
     +--------<--------+
              |
              v
       +--------------+
       | Slack Message|
       +--------------+
```

---

# ASCII data/event flow diagram

```text
[CLI args]
   |
   v
[Raw Config]
   |
   v
[Validated Config]
   |
   v
[PipelineRequest Model]
   |
   v
[Router selects workflow]
   |
   +--> [Send Slack: pipeline started]
   |
   +--> [Run shell script]
              |
              v
        [tkn/kubectl command]
              |
              v
        [Tekton PipelineRun]
              |
              v
        [Kubernetes Pods]
              |
              v
        [Task status/logs/events]
              |
              v
        [FailureTrace Model]
              |
              v
        [Slack Message Renderer]
              |
              v
        [Send Slack: success/failure]
```

---

# 3. Responsibility of each component

## `main.go`

Responsibility:

* Start the program.
* Call the CLI/root command.
* Exit with the correct status code.

It should **not** contain business logic.

Bad `main.go`:

```go
func main() {
    // parse config
    // validate config
    // call kubectl
    // parse logs
    // format Slack JSON
    // send HTTP request
}
```

Good `main.go`:

```go
func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

Python comparison:

```python
if __name__ == "__main__":
    main()
```

In Python, you also keep the script entrypoint thin and move logic into modules.

---

## CLI layer

Responsibility:

* Parse commands and flags.
* Accept input like pipeline name, namespace, config path, Slack channel.
* Convert CLI input into application-level options.

Example command:

```text
notifier run --config config.yaml --pipeline build-api --namespace dev
```

Python equivalent:

* `argparse`
* `click`
* `typer`

Go equivalent:

* `flag`
* `cobra`
* `urfave/cli`

---

## Config package

Responsibility:

* Load YAML/JSON/env variables.
* Apply defaults.
* Validate required fields.
* Hide configuration source details from the rest of the app.

Example config:

```yaml
slack:
  webhookURL: "${SLACK_WEBHOOK_URL}"
  channel: "#deployments"

tekton:
  namespace: "ci"
  pipelineName: "build-and-test"

kubernetes:
  kubeconfig: "~/.kube/config"

notifier:
  includeLogs: true
  maxLogLines: 80
```

Go concept:

```go
type Config struct {
    Slack      SlackConfig
    Tekton     TektonConfig
    Kubernetes KubernetesConfig
    Notifier   NotifierConfig
}
```

Python comparison:

```python
@dataclass
class Config:
    slack: SlackConfig
    tekton: TektonConfig
```

In Go, structs are commonly used like Python dataclasses, but Go gives you compile-time type checking.

---

## Model package

Responsibility:

* Define core domain objects.
* Avoid leaking Slack, Tekton, or Kubernetes implementation details everywhere.

Useful models:

```go
type PipelineRequest struct {
    PipelineName string
    Namespace    string
    TriggeredBy  string
    CommitSHA    string
}

type PipelineResult struct {
    Name      string
    Namespace string
    Status    PipelineStatus
    Duration  time.Duration
}

type FailureTrace struct {
    PipelineRunName string
    FailedTaskName  string
    PodName         string
    Reason          string
    Message         string
    Logs            []string
    Events          []string
}

type NotificationEvent struct {
    Title        string
    Status       string
    Summary      string
    FailureTrace *FailureTrace
}
```

Python comparison:

```python
@dataclass
class FailureTrace:
    pipeline_run_name: str
    failed_task_name: str
    pod_name: str
    reason: str
    logs: list[str]
```

Important Go idea:

> Keep data structures explicit. Do not pass around loose `map[string]interface{}` unless absolutely needed.

---

## Router package

Responsibility:

* Decide what workflow should happen.
* Connect high-level use cases.
* Avoid putting orchestration logic inside `main.go`.

Example responsibilities:

* `run pipeline and notify`
* `only send test Slack message`
* `trace last failed pipeline`
* `dry-run config validation`

Example:

```go
type Router struct {
    SlackClient   Notifier
    ScriptRunner  Runner
    FailureTracer Tracer
    Renderer      MessageRenderer
}
```

The router is similar to a backend controller or service layer.

Python comparison:

```python
class Router:
    def __init__(self, slack_client, script_runner, failure_tracer):
        ...
```

---

## Slack client package

Responsibility:

* Send Slack messages.
* Hide HTTP request details.
* Handle retryable failures.
* Return meaningful errors.

Interface:

```go
type Notifier interface {
    Send(ctx context.Context, msg SlackMessage) error
}
```

Implementation:

```go
type SlackClient struct {
    WebhookURL string
    HTTPClient *http.Client
}
```

Python comparison:

```python
class Notifier(Protocol):
    def send(self, message: SlackMessage) -> None:
        ...
```

In Go, interfaces are usually small and behavior-based.

---

## Shell script runner package

Responsibility:

* Execute external commands.
* Capture stdout, stderr, and exit code.
* Apply timeout.
* Return structured result.

Example:

```go
type CommandResult struct {
    ExitCode int
    Stdout   string
    Stderr   string
}
```

This is the boundary between your Go app and the operating system.

Python equivalent:

```python
subprocess.run(...)
```

Go equivalent:

```go
exec.CommandContext(ctx, "bash", "scripts/run_pipeline.sh")
```

---

## Tekton package

Responsibility:

* Understand Tekton concepts.
* Work with PipelineRuns, TaskRuns, and statuses.
* Possibly call `tkn` or the Kubernetes API.

Key Tekton resources:

```text
Pipeline
PipelineRun
Task
TaskRun
Step
```

Your notifier does not need to own Tekton. It only needs to:

* trigger a pipeline
* observe status
* collect failure information

---

## Kubernetes package

Responsibility:

* Talk to Kubernetes resources.
* Get pods, logs, events, namespaces, and container statuses.
* Avoid mixing Kubernetes API details into Slack formatting or CLI parsing.

Useful data:

```text
Pod phase
Container exit code
Container reason
Pod events
Container logs
Namespace
Labels
Owner references
```

---

## Failure tracer package

Responsibility:

* Convert low-level infrastructure failure into useful debugging context.

Example:

```text
Tekton PipelineRun failed
 -> find failed TaskRun
 -> find related Pod
 -> read container status
 -> read logs
 -> read Kubernetes events
 -> build FailureTrace
```

This package is extremely important because it transforms raw platform noise into human-readable information.

---

## Message renderer package

Responsibility:

* Convert internal event models into Slack-ready messages.
* Keep formatting separate from business logic.

Example:

```go
type MessageRenderer interface {
    RenderPipelineResult(result PipelineResult, trace *FailureTrace) SlackMessage
}
```

This helps because later you could add:

* Slack
* Teams
* email
* PagerDuty
* GitHub comment

without changing the tracing logic.

---

# 4. Why `main.go` should stay small

`main.go` should stay small because it is only the **entrypoint**, not the application.

A small `main.go` gives you:

* easier testing
* cleaner package design
* reusable logic
* less coupling
* better error handling
* easier interview explanation

A good mental model:

```text
main.go starts the car.
It should not contain the engine, transmission, GPS, fuel system, and dashboard.
```

In production Go services, `main.go` often does only this:

```go
func main() {
    ctx := context.Background()

    app, err := app.New(ctx)
    if err != nil {
        log.Fatal(err)
    }

    if err := app.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

Python equivalent:

```python
def main():
    app = create_app()
    app.run()

if __name__ == "__main__":
    main()
```

---

# 5. How this resembles a production backend/cloud workflow

Even though your project may run as a CLI, it has production-style architecture.

It resembles production because it has:

| Production Concept         | Your Project Equivalent    |
| -------------------------- | -------------------------- |
| API input                  | CLI flags/config           |
| request model              | `PipelineRequest`          |
| service/router layer       | router package             |
| external client            | Slack client               |
| infrastructure integration | Tekton/Kubernetes          |
| observability              | logs/debug traces          |
| failure handling           | `FailureTrace`             |
| config management          | YAML/env                   |
| secrets                    | Slack webhook, kubeconfig  |
| retries/timeouts           | HTTP and command execution |
| testing boundaries         | interfaces and mocks       |

The most important production idea:

> Your application does not do everything itself. It coordinates external systems safely.

That is exactly what backend/cloud software often does.

---

# 6. Where it is CLI-based and where it resembles a service

## CLI-based parts

These are clearly CLI responsibilities:

```text
notifier run --pipeline api-build
notifier trace --pipeline-run api-build-123
notifier test-slack
notifier validate-config
```

CLI traits:

* runs on demand
* receives flags
* exits with status code
* prints logs to terminal
* may be run inside CI/CD

## Service-like parts

These resemble backend service design:

```text
Router
Config loader
Slack client
Tekton client
Kubernetes client
Failure tracer
Message renderer
Validation
Logging
Testing
```

Service traits:

* has clean boundaries
* has dependency injection
* talks to external systems
* has structured errors
* has domain models
* can be tested with mocks

You can explain this in an interview as:

> “The current project is a CLI, but I structured the core logic like a service. The CLI is only one adapter. Later the same core could run behind an HTTP endpoint, cron job, Kubernetes controller, or Tekton task.”

That is a strong cloud-native answer.

---

# 7. How failure information moves through the system

Failure information should become more structured as it moves inward.

```text
Shell exit code / stderr
        |
        v
Tekton PipelineRun condition
        |
        v
Failed TaskRun
        |
        v
Kubernetes Pod status
        |
        v
Container state / exit code
        |
        v
Pod logs and events
        |
        v
FailureTrace model
        |
        v
Slack message
```

Example raw failure:

```text
Error from server (Forbidden): pods is forbidden
```

Better internal model:

```go
FailureTrace{
    FailedTaskName: "deploy",
    PodName: "deploy-app-pod",
    Reason: "RBACDenied",
    Message: "ServiceAccount cannot list pods in namespace ci",
}
```

Better Slack message:

```text
:red_circle: Pipeline failed: build-and-test

Failed task: deploy
Pod: deploy-app-pod
Reason: RBACDenied

Likely cause:
The service account does not have permission to list pods in namespace ci.

Next step:
Check Role/RoleBinding for the Tekton service account.
```

The key is that your project should avoid dumping only raw logs. It should extract enough context for a human to act.

---

# 8. Where logging, validation, testing, and configuration fit

## Logging

Logging belongs at boundaries and important state transitions.

Good logging points:

```text
config loaded
config validation failed
pipeline started
script executed
Tekton status received
failure trace collected
Slack message sent
Slack message failed
```

Avoid logging secrets:

```text
BAD: Slack webhook URL = https://hooks.slack.com/...
GOOD: Slack webhook configured = true
```

Go tools:

```go
log/slog
zap
zerolog
```

Python equivalents:

```python
logging
structlog
loguru
```

---

## Validation

Validation should happen early.

Validate:

* Slack webhook exists
* namespace is not empty
* pipeline name is not empty
* timeout is positive
* max log lines is reasonable
* config file is valid YAML
* kubeconfig or in-cluster config is available

Good rule:

> Fail fast before touching Slack, Tekton, or Kubernetes.

---

## Testing

Test each boundary separately.

| Package  | Test style                                 |
| -------- | ------------------------------------------ |
| config   | table tests for valid/invalid config       |
| model    | simple struct behavior                     |
| router   | mock Slack, mock runner, mock tracer       |
| slack    | test request payload with fake HTTP server |
| script   | fake command runner                        |
| tekton   | fake client or mocked output               |
| kube     | fake Kubernetes client                     |
| trace    | sample PipelineRun/Pod failure fixtures    |
| renderer | golden tests for Slack messages            |

Go testing style:

```go
func TestConfigValidationMissingSlackWebhook(t *testing.T) {
    cfg := Config{}
    err := cfg.Validate()

    if err == nil {
        t.Fatal("expected validation error")
    }
}
```

Python comparison:

```python
def test_config_validation_missing_slack_webhook():
    cfg = Config()
    with pytest.raises(ValueError):
        cfg.validate()
```

Go commonly uses table-driven tests:

```go
tests := []struct {
    name    string
    cfg     Config
    wantErr bool
}{
    {"missing webhook", Config{}, true},
    {"valid config", validConfig(), false},
}
```

---

# Hands-on architecture: at least 7 clean boundaries/packages

A clean package layout could be:

```text
slack-tekton-notifier/
|
+-- cmd/
|   +-- notifier/
|       +-- main.go
|
+-- internal/
|   +-- cli/
|   |   +-- command.go
|   |
|   +-- config/
|   |   +-- config.go
|   |   +-- validate.go
|   |
|   +-- model/
|   |   +-- pipeline.go
|   |   +-- notification.go
|   |   +-- failure.go
|   |
|   +-- router/
|   |   +-- router.go
|   |
|   +-- slack/
|   |   +-- client.go
|   |   +-- message.go
|   |
|   +-- runner/
|   |   +-- shell.go
|   |
|   +-- tekton/
|   |   +-- client.go
|   |   +-- status.go
|   |
|   +-- kube/
|   |   +-- client.go
|   |   +-- logs.go
|   |   +-- events.go
|   |
|   +-- trace/
|   |   +-- failure_tracer.go
|   |
|   +-- render/
|   |   +-- slack_renderer.go
|   |
|   +-- logging/
|   |   +-- logger.go
|   |
|   +-- app/
|       +-- app.go
|
+-- scripts/
|   +-- run_pipeline.sh
|
+-- configs/
|   +-- example.yaml
|
+-- testdata/
|   +-- failed-pipelinerun.yaml
|   +-- failed-pod.json
```

That gives more than 7 boundaries:

1. `cli`
2. `config`
3. `model`
4. `router`
5. `slack`
6. `runner`
7. `tekton`
8. `kube`
9. `trace`
10. `render`
11. `logging`
12. `app`

---

# Service boundary analysis

A service boundary is where one responsibility ends and another begins.

## Boundary 1: CLI boundary

```text
User input -> internal request
```

The CLI should not know Slack JSON, Tekton internals, or Kubernetes pod structure.

---

## Boundary 2: Configuration boundary

```text
YAML/env/flags -> typed Config
```

After config loading, the rest of the app should receive clean Go structs.

---

## Boundary 3: Domain boundary

```text
External data -> internal models
```

Your app should use models like:

```go
PipelineResult
FailureTrace
NotificationEvent
```

instead of passing raw command output everywhere.

---

## Boundary 4: Slack boundary

```text
NotificationEvent -> Slack API request
```

Only the Slack package should know Slack payload structure.

---

## Boundary 5: Shell boundary

```text
Go app -> external process
```

The runner should isolate:

* command execution
* timeout
* stdout
* stderr
* exit code

---

## Boundary 6: Tekton boundary

```text
Tekton concepts -> pipeline status
```

The rest of the app should not need to know every Tekton field.

---

## Boundary 7: Kubernetes boundary

```text
Kubernetes resources -> logs/events/status
```

The tracer can ask the Kubernetes package for pod logs without knowing the raw API details.

---

## Boundary 8: Rendering boundary

```text
Internal result -> human-readable message
```

Formatting should be separate so you can change Slack layout without touching tracing logic.

---

# Reusable abstraction analysis

Good abstractions make this project reusable.

## `Notifier`

```go
type Notifier interface {
    Send(ctx context.Context, event model.NotificationEvent) error
}
```

Current implementation:

```text
SlackNotifier
```

Future implementations:

```text
TeamsNotifier
EmailNotifier
PagerDutyNotifier
GitHubCommentNotifier
```

---

## `Runner`

```go
type Runner interface {
    Run(ctx context.Context, command Command) (CommandResult, error)
}
```

Current implementation:

```text
ShellRunner
```

Future implementations:

```text
TektonAPIRunner
FakeRunner for tests
RemoteRunner
```

---

## `Tracer`

```go
type Tracer interface {
    TraceFailure(ctx context.Context, ref PipelineReference) (*FailureTrace, error)
}
```

Current implementation:

```text
TektonKubernetesTracer
```

Future implementations:

```text
GitHubActionsTracer
ArgoCDTracer
JenkinsTracer
```

---

## `MessageRenderer`

```go
type MessageRenderer interface {
    Render(event NotificationEvent) SlackMessage
}
```

Current implementation:

```text
SlackBlockRenderer
```

Future implementations:

```text
PlainTextRenderer
MarkdownRenderer
JSONRenderer
```

---

## `ConfigLoader`

```go
type ConfigLoader interface {
    Load(path string) (Config, error)
}
```

Current implementation:

```text
YAMLConfigLoader
```

Future implementations:

```text
EnvConfigLoader
VaultConfigLoader
ConfigMapLoader
```

---

# Pseudocode for the full end-to-end flow

```go
func main() {
    err := cli.Execute()
    if err != nil {
        os.Exit(1)
    }
}
```

```go
func ExecuteRunCommand(args Args) error {
    ctx := context.Background()

    cfg, err := config.Load(args.ConfigPath)
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    if err := cfg.Validate(); err != nil {
        return fmt.Errorf("validate config: %w", err)
    }

    app := app.New(cfg)

    request := model.PipelineRequest{
        PipelineName: args.PipelineName,
        Namespace:    cfg.Tekton.Namespace,
        TriggeredBy:  args.User,
        CommitSHA:    args.CommitSHA,
    }

    return app.Router.RunPipelineAndNotify(ctx, request)
}
```

```go
func (r *Router) RunPipelineAndNotify(ctx context.Context, req PipelineRequest) error {
    startEvent := model.NotificationEvent{
        Title:  "Pipeline started",
        Status: "running",
        Summary: req.PipelineName,
    }

    if err := r.Notifier.Send(ctx, startEvent); err != nil {
        r.Logger.Warn("failed to send start notification", "error", err)
    }

    result, err := r.Runner.Run(ctx, Command{
        Name: "bash",
        Args: []string{
            "scripts/run_pipeline.sh",
            req.PipelineName,
            req.Namespace,
        },
    })

    if err == nil && result.ExitCode == 0 {
        successEvent := model.NotificationEvent{
            Title:  "Pipeline succeeded",
            Status: "success",
            Summary: req.PipelineName,
        }

        return r.Notifier.Send(ctx, successEvent)
    }

    trace, traceErr := r.Tracer.TraceFailure(ctx, model.PipelineReference{
        Name:      req.PipelineName,
        Namespace: req.Namespace,
    })

    failureEvent := model.NotificationEvent{
        Title:        "Pipeline failed",
        Status:       "failure",
        Summary:      req.PipelineName,
        FailureTrace: trace,
    }

    notifyErr := r.Notifier.Send(ctx, failureEvent)

    if traceErr != nil {
        return fmt.Errorf("pipeline failed; additionally failed to trace error: %w", traceErr)
    }

    if notifyErr != nil {
        return fmt.Errorf("pipeline failed; additionally failed to notify Slack: %w", notifyErr)
    }

    return fmt.Errorf("pipeline failed: exit code %d", result.ExitCode)
}
```

---

# Simple example

## Scenario

A developer runs:

```text
notifier run --pipeline hello-api --namespace ci
```

The shell script starts a Tekton pipeline.

The pipeline fails because tests fail.

## Failure trace

```text
PipelineRun: hello-api-run-abc123
Failed Task: run-tests
Pod: run-tests-pod
Reason: TestFailure
Message: go test ./... failed
```

## Slack message

```text
:red_circle: Pipeline failed: hello-api

Namespace: ci
Failed task: run-tests
Pod: run-tests-pod
Reason: TestFailure

Last log lines:
--- FAIL: TestCreateUser
expected status 201, got 500

Next step:
Run go test ./... locally and inspect TestCreateUser.
```

This is simple but already useful.

---

# Production-style example

## Scenario

A production CI system runs your notifier inside a Tekton Task.

```text
notifier run \
  --config /workspace/config/notifier.yaml \
  --pipeline payment-service-release \
  --namespace release \
  --commit 8f41a9c \
  --triggered-by github-actions
```

## Production concerns

The notifier should:

* use environment variables for secrets
* not print Slack webhook URL
* use timeouts
* retry Slack delivery
* collect Kubernetes events
* collect only the last N log lines
* avoid leaking sensitive logs
* return non-zero exit code on failure
* emit structured logs
* work in-cluster using service account permissions

## Production Slack message

```text
:red_circle: Release pipeline failed

Service: payment-service
Environment: staging
Namespace: release
PipelineRun: payment-service-release-8f41a9c
Commit: 8f41a9c
Failed task: deploy
Pod: deploy-payment-service-pod
Reason: ImagePullBackOff

Likely cause:
Kubernetes could not pull image registry.example.com/payment-service:8f41a9c

Useful details:
- Check image tag exists
- Check imagePullSecret
- Check registry availability
- Check service account permissions

Last event:
Failed to pull image "registry.example.com/payment-service:8f41a9c"
```

This is much closer to what real platform teams want.

---

# Trade-offs and failure cases

## Trade-off: shell script vs direct API

### Shell script approach

Pros:

* simple
* easy to understand
* reuses `kubectl` and `tkn`
* fast to build

Cons:

* harder to test
* depends on local tools
* output parsing can break
* less portable

### Direct Kubernetes/Tekton API approach

Pros:

* more reliable
* easier to structure data
* better testability with fake clients
* production-grade

Cons:

* more code
* more types
* more Kubernetes knowledge required

Best learning path:

```text
Start with shell script.
Wrap it behind an interface.
Later replace implementation with direct API client.
```

---

## Trade-off: one big package vs many packages

One big package is easier at first.

But many clean packages help you explain architecture:

```text
config != router != slack != trace != kube
```

For interviews, clean package boundaries are more impressive than a large working script.

---

## Trade-off: raw logs vs summarized failure trace

Raw logs are easy.

Summarized traces are better.

Bad Slack message:

```text
Pipeline failed. See logs.
```

Better Slack message:

```text
Pipeline failed at task deploy because image pull failed.
```

Best Slack message:

```text
Pipeline failed at task deploy because image registry.example.com/app:abc123 does not exist.
Check image build step or registry credentials.
```

---

## Common failure cases

| Failure                  | Where detected    | Handling                   |
| ------------------------ | ----------------- | -------------------------- |
| Missing config file      | config loader     | return validation error    |
| Missing Slack webhook    | config validation | fail fast                  |
| Invalid namespace        | Tekton/Kubernetes | include namespace in error |
| `tkn` not installed      | shell runner      | command-not-found error    |
| Pipeline fails           | Tekton            | trace TaskRun/Pod          |
| Pod cannot start         | Kubernetes        | inspect pod events         |
| Container exits non-zero | Kubernetes        | inspect logs               |
| Slack API fails          | Slack client      | retry, log, return error   |
| Logs are huge            | tracer            | limit log lines            |
| Secret appears in logs   | renderer/tracer   | redact sensitive patterns  |
| Timeout                  | context           | cancel command/API call    |

---

# Debugging checklist

Use this when the notifier fails or sends poor messages.

## CLI/debug input

```text
[ ] Did I pass the correct config path?
[ ] Did I pass the correct pipeline name?
[ ] Did I pass the correct namespace?
[ ] Is the command running from the expected working directory?
```

## Config

```text
[ ] Is Slack webhook configured?
[ ] Is namespace configured?
[ ] Is maxLogLines reasonable?
[ ] Are secrets coming from env vars, not committed YAML?
```

## Shell runner

```text
[ ] Is bash available?
[ ] Is the script executable?
[ ] Is tkn installed?
[ ] Is kubectl installed?
[ ] Does the command work manually?
```

## Tekton

```text
[ ] Does the Pipeline exist?
[ ] Was a PipelineRun created?
[ ] Which TaskRun failed?
[ ] What condition reason did Tekton report?
```

## Kubernetes

```text
[ ] Was a Pod created?
[ ] Did the Pod start?
[ ] Did a container fail?
[ ] What is the container exit code?
[ ] Are there pod events?
[ ] Are logs available?
```

## Slack

```text
[ ] Is the webhook valid?
[ ] Is the payload valid JSON?
[ ] Is the message too large?
[ ] Is Slack reachable from the runtime environment?
```

## Logging

```text
[ ] Are logs structured?
[ ] Do logs include pipeline name and namespace?
[ ] Are secrets redacted?
[ ] Is the original error wrapped with context?
```

---

# Go concepts compared with Python

## 1. `struct` vs Python `dataclass`

Go:

```go
type PipelineRequest struct {
    PipelineName string
    Namespace    string
}
```

Python:

```python
@dataclass
class PipelineRequest:
    pipeline_name: str
    namespace: str
```

Go structs are compile-time checked and commonly used for config, models, API payloads, and domain objects.

---

## 2. Interfaces vs duck typing

Go:

```go
type Notifier interface {
    Send(ctx context.Context, event NotificationEvent) error
}
```

Python:

```python
class Notifier(Protocol):
    def send(self, event: NotificationEvent) -> None:
        ...
```

Python often uses duck typing implicitly.

Go makes the expected behavior explicit through interfaces.

---

## 3. Error returns vs exceptions

Go:

```go
cfg, err := config.Load(path)
if err != nil {
    return fmt.Errorf("load config: %w", err)
}
```

Python:

```python
try:
    cfg = load_config(path)
except ConfigError as e:
    raise RuntimeError("load config") from e
```

Go encourages explicit error handling at every boundary.

For this project, that is useful because many things can fail:

```text
files
env vars
shell commands
Slack API
Tekton
Kubernetes
network
```

---

## 4. `context.Context` vs timeout/cancellation patterns

Go:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()
```

Python equivalent:

```python
asyncio.wait_for(...)
subprocess.run(..., timeout=120)
requests.post(..., timeout=10)
```

In Go backend/cloud code, `context.Context` is central. It carries cancellation, deadlines, and request-scoped values.

Use it for:

* Slack HTTP calls
* shell command timeout
* Kubernetes API calls
* Tekton polling

---

## 5. Goroutines vs Python async/threading

Go:

```go
go sendNotification()
```

Python:

```python
asyncio.create_task(send_notification())
```

or:

```python
threading.Thread(target=send_notification).start()
```

For this notifier, you probably do not need much concurrency on Day 1. Later you could use goroutines to:

* poll Tekton status
* stream logs
* send Slack notification asynchronously
* collect pod events and logs in parallel

---

## 6. Packages vs Python modules

Go:

```text
internal/config
internal/slack
internal/trace
```

Python:

```text
notifier/config.py
notifier/slack.py
notifier/trace.py
```

The idea is the same:

> group related code by responsibility.

---

## 7. Compile-time binary vs interpreted script

Python project:

```text
python notifier.py
```

Go project:

```text
go build -o notifier
./notifier
```

For DevOps tools, Go binaries are convenient because they are easy to ship into containers and CI environments.

---

# Interview questions with sample answers

## 1. Why did you keep `main.go` small?

Sample answer:

> I kept `main.go` small because it should only be the entrypoint. The real logic belongs in packages like config, router, Slack, Tekton, and tracing. This makes the project easier to test, easier to maintain, and closer to production Go service structure.

---

## 2. How does failure information move through your notifier?

Sample answer:

> The notifier starts with a command result or Tekton status. If the pipeline fails, it identifies the failed PipelineRun or TaskRun, finds the related Kubernetes pod, reads container status, logs, and events, then converts that into a `FailureTrace` model. The Slack renderer turns that structured trace into a human-readable message.

---

## 3. Why use interfaces in this project?

Sample answer:

> Interfaces allow me to separate behavior from implementation. For example, the router depends on a `Notifier` interface, not directly on Slack. That means I can test the router with a fake notifier and later add Teams, email, or PagerDuty without rewriting the router.

---

## 4. Is this project a CLI or a service?

Sample answer:

> It is currently a CLI because it is triggered by command-line input and exits with a status code. But the internal architecture resembles a service because it has configuration, routing, domain models, clients, logging, validation, and external integrations. The CLI is just one adapter around reusable core logic.

---

## 5. What production improvements would you add?

Sample answer:

> I would add structured logging, retries with backoff for Slack, context-based timeouts, Kubernetes client-go integration instead of parsing shell output, redaction of secrets in logs, stronger config validation, unit tests with fake clients, and possibly metrics for notification success and failure rates.

---

# Final mental model

This project is not just:

```text
Go program sends Slack message
```

It is better described as:

```text
A Go-based cloud workflow coordinator that reads configuration,
executes or observes Tekton pipelines on Kubernetes,
extracts structured failure context,
and sends actionable Slack notifications.
```

That is exactly the kind of architecture you should be comfortable explaining for an IBM Cloud Data Services Software Developer style role.
# DSA Revision in Go: Arrays, Slices, and Big-O

Since you already know Python, think of this lesson as:

```text
Python list  ≈  Go slice
Python tuple ≈  Go array-like fixed collection, but not exactly
```

In Go, **arrays** and **slices** are different. This is one of the first important Go DSA concepts.

---

# 1. Arrays in Go

A Go array has a **fixed size**.

```go
package main

import "fmt"

func main() {
    var nums [3]int

    nums[0] = 10
    nums[1] = 20
    nums[2] = 30

    fmt.Println(nums)
}
```

Output:

```text
[10 20 30]
```

The size is part of the type.

```go
var a [3]int
var b [5]int
```

`[3]int` and `[5]int` are different types.

This is unlike Python, where a list can grow dynamically:

```python
nums = [10, 20, 30]
nums.append(40)
```

In Go, this array cannot grow:

```go
var nums [3]int
```

You cannot append to a fixed array directly.

---

# 2. Slices in Go

A slice is a flexible, dynamic view over an array.

Most of the time in Go DSA, you will use slices, not arrays.

```go
package main

import "fmt"

func main() {
    nums := []int{10, 20, 30}

    nums = append(nums, 40)

    fmt.Println(nums)
}
```

Output:

```text
[10 20 30 40]
```

Python comparison:

```python
nums = [10, 20, 30]
nums.append(40)
```

Go slice:

```go
nums := []int{10, 20, 30}
nums = append(nums, 40)
```

The important difference:

```go
nums = append(nums, 40)
```

In Go, `append` returns a new slice, so you must assign it back.

---

# 3. Array vs Slice

| Concept       | Go Array        | Go Slice       | Python Equivalent      |
| ------------- | --------------- | -------------- | ---------------------- |
| Size          | Fixed           | Dynamic        | Python list is dynamic |
| Syntax        | `[3]int{1,2,3}` | `[]int{1,2,3}` | `[1,2,3]`              |
| Common in DSA | Less common     | Very common    | Python list            |
| Can append?   | No              | Yes            | Yes                    |
| Length        | `len(arr)`      | `len(slice)`   | `len(list)`            |
| Capacity      | Fixed           | Has capacity   | Python hides this      |

Example:

```go
arr := [3]int{1, 2, 3}
slice := []int{1, 2, 3}
```

The only visual difference is the size inside brackets:

```go
[3]int  // array
[]int   // slice
```

---

# 4. Slice Length and Capacity

A slice has:

```text
length   = number of visible elements
capacity = size available before Go needs a new backing array
```

Example:

```go
package main

import "fmt"

func main() {
    nums := make([]int, 3, 5)

    fmt.Println(nums)
    fmt.Println("len:", len(nums))
    fmt.Println("cap:", cap(nums))
}
```

Output:

```text
[0 0 0]
len: 3
cap: 5
```

Meaning:

```text
The slice currently has 3 elements.
It can grow up to 5 elements before needing a new backing array.
```

Python hides this from you. In Python, list capacity exists internally, but you usually do not think about it.

---

# 5. Appending to a Slice

```go
nums := []int{1, 2, 3}

nums = append(nums, 4)
nums = append(nums, 5)

fmt.Println(nums)
```

Output:

```text
[1 2 3 4 5]
```

Important:

```go
append(nums, 4)
```

does not mutate the variable assignment automatically. You need:

```go
nums = append(nums, 4)
```

Python:

```python
nums.append(4)
```

Python modifies the list in place.

Go:

```go
nums = append(nums, 4)
```

Go returns the updated slice.

---

# 6. Iterating Over Slices

Go:

```go
nums := []int{10, 20, 30}

for i, value := range nums {
    fmt.Println(i, value)
}
```

Python:

```python
nums = [10, 20, 30]

for i, value in enumerate(nums):
    print(i, value)
```

If you only need the value in Go:

```go
for _, value := range nums {
    fmt.Println(value)
}
```

The `_` means “ignore this value.”

---

# 7. Indexing

Go:

```go
nums := []int{10, 20, 30}

fmt.Println(nums[0])
fmt.Println(nums[1])
```

Python:

```python
nums = [10, 20, 30]

print(nums[0])
print(nums[1])
```

Both are `O(1)` operations.

That means direct access by index is constant time.

---

# 8. Slicing

Go:

```go
nums := []int{10, 20, 30, 40, 50}

part := nums[1:4]

fmt.Println(part)
```

Output:

```text
[20 30 40]
```

Python:

```python
nums = [10, 20, 30, 40, 50]

part = nums[1:4]
```

Important difference:

In Python, slicing usually creates a new list.

In Go, slicing creates a new slice view over the same backing array.

Example:

```go
package main

import "fmt"

func main() {
    nums := []int{10, 20, 30, 40, 50}

    part := nums[1:4]
    part[0] = 999

    fmt.Println(nums)
    fmt.Println(part)
}
```

Output:

```text
[10 999 30 40 50]
[999 30 40]
```

This surprises many Python developers.

Python comparison:

```python
nums = [10, 20, 30, 40, 50]
part = nums[1:4]

part[0] = 999

print(nums)
print(part)
```

Output:

```text
[10, 20, 30, 40, 50]
[999, 30, 40]
```

So remember:

```text
Go slice shares memory.
Python slice usually copies.
```

---

# 9. Big-O Revision

Big-O describes how runtime or memory grows as input size grows.

Let `n` be the number of elements.

## Common complexities

| Big-O        | Meaning           | Example                       |
| ------------ | ----------------- | ----------------------------- |
| `O(1)`       | constant time     | access `nums[i]`              |
| `O(n)`       | linear time       | loop through slice            |
| `O(n²)`      | quadratic time    | nested loops                  |
| `O(log n)`   | logarithmic time  | binary search                 |
| `O(n log n)` | efficient sorting | merge sort, quicksort average |

---

## O(1)

```go
x := nums[3]
```

Direct index access is constant time.

Python equivalent:

```python
x = nums[3]
```

---

## O(n)

```go
for _, value := range nums {
    fmt.Println(value)
}
```

You visit each element once.

Python equivalent:

```python
for value in nums:
    print(value)
```

---

## O(n²)

```go
for i := 0; i < len(nums); i++ {
    for j := 0; j < len(nums); j++ {
        fmt.Println(nums[i], nums[j])
    }
}
```

For every element, you scan every other element.

If `n = 1000`, this can become roughly 1,000,000 operations.

---

# 10. Common Slice Operations and Big-O

| Operation             | Go Example        | Time Complexity                  |
| --------------------- | ----------------- | -------------------------------- |
| Access by index       | `nums[i]`         | `O(1)`                           |
| Update by index       | `nums[i] = 10`    | `O(1)`                           |
| Append at end         | `append(nums, x)` | usually `O(1)`, sometimes `O(n)` |
| Search unsorted slice | loop              | `O(n)`                           |
| Delete from middle    | copy elements     | `O(n)`                           |
| Insert in middle      | shift elements    | `O(n)`                           |
| Sort                  | `sort.Ints(nums)` | `O(n log n)`                     |

Why append is “usually `O(1)` but sometimes `O(n)`”:

If the slice has enough capacity, Go adds the value directly.

If capacity is full, Go creates a bigger backing array and copies old elements.

That copy costs `O(n)`.

But averaged over many appends, append is considered **amortized O(1)**.

---

# Easy Go DSA Problem

## Problem: Two Sum

Given a slice of integers and a target number, return the indices of two numbers that add up to the target.

Example:

```text
nums = [2, 7, 11, 15]
target = 9
answer = [0, 1]
```

Because:

```text
nums[0] + nums[1] = 2 + 7 = 9
```

---

# Brute-force Thinking

The simplest idea:

```text
Check every pair.
```

For each number, compare it with every number after it.

```text
nums = [2, 7, 11, 15]

Check:
2 + 7
2 + 11
2 + 15
7 + 11
7 + 15
11 + 15
```

## Brute-force Go solution

```go
package main

import "fmt"

func twoSumBrute(nums []int, target int) []int {
    for i := 0; i < len(nums); i++ {
        for j := i + 1; j < len(nums); j++ {
            if nums[i]+nums[j] == target {
                return []int{i, j}
            }
        }
    }

    return []int{}
}

func main() {
    nums := []int{2, 7, 11, 15}
    target := 9

    result := twoSumBrute(nums, target)

    fmt.Println(result)
}
```

## Complexity

```text
Time:  O(n²)
Space: O(1)
```

Why time is `O(n²)`?

Because we use nested loops.

Why space is `O(1)`?

Because we do not create extra data structures that grow with input size.

---

# Optimized Thinking

Instead of checking every pair, ask:

```text
For current number x, what number do I need?
```

Formula:

```text
needed = target - x
```

For example:

```text
target = 9

current = 2
needed = 9 - 2 = 7
```

So while scanning the slice, store previously seen numbers in a map.

In Python, this would be a dictionary.

In Go, this is a map.

Python:

```python
seen = {}
```

Go:

```go
seen := make(map[int]int)
```

This map stores:

```text
number -> index
```

---

## Optimized Go solution

```go
package main

import "fmt"

func twoSum(nums []int, target int) []int {
    seen := make(map[int]int)

    for i, value := range nums {
        needed := target - value

        if previousIndex, ok := seen[needed]; ok {
            return []int{previousIndex, i}
        }

        seen[value] = i
    }

    return []int{}
}

func main() {
    nums := []int{2, 7, 11, 15}
    target := 9

    result := twoSum(nums, target)

    fmt.Println(result)
}
```

Output:

```text
[0 1]
```

---

# Understanding the Go Map Check

This line is very important:

```go
if previousIndex, ok := seen[needed]; ok {
    return []int{previousIndex, i}
}
```

In Go, when reading from a map, you can get two values:

```go
value, ok := myMap[key]
```

Meaning:

```text
value = value stored for that key
ok    = true if key exists, false otherwise
```

Python equivalent:

```python
if needed in seen:
    return [seen[needed], i]
```

Go:

```go
if previousIndex, ok := seen[needed]; ok {
    return []int{previousIndex, i}
}
```

Python:

```python
if needed in seen:
    return [seen[needed], i]
```

---

# Optimized Complexity

```text
Time:  O(n)
Space: O(n)
```

Why time is `O(n)`?

We loop through the slice once.

Map lookup is average `O(1)`.

Why space is `O(n)`?

In the worst case, we may store every number in the map.

---

# Brute Force vs Optimized

| Approach    |    Time |  Space | Idea                            |
| ----------- | ------: | -----: | ------------------------------- |
| Brute force | `O(n²)` | `O(1)` | Check every pair                |
| Optimized   |  `O(n)` | `O(n)` | Use map to remember seen values |

This is a classic DSA trade-off:

```text
Use extra memory to reduce time.
```

---

# Python-to-Go DSA Mapping

| Python                        | Go                        |
| ----------------------------- | ------------------------- |
| `list[int]`                   | `[]int`                   |
| `dict[int, int]`              | `map[int]int`             |
| `len(nums)`                   | `len(nums)`               |
| `nums.append(x)`              | `nums = append(nums, x)`  |
| `for i, x in enumerate(nums)` | `for i, x := range nums`  |
| `if x in seen`                | `if _, ok := seen[x]; ok` |
| `return []`                   | `return []int{}`          |
| `None`                        | `nil`                     |
| exception                     | `error` return            |

---

# Important Go Details for DSA

## 1. Function input as slice

Most DSA functions look like this:

```go
func solve(nums []int) int {
    // logic
}
```

This is similar to Python:

```python
def solve(nums: list[int]) -> int:
    ...
```

---

## 2. Return a slice

```go
return []int{0, 1}
```

Python:

```python
return [0, 1]
```

---

## 3. Empty result

Go:

```go
return []int{}
```

Python:

```python
return []
```

Sometimes Go code returns `nil` instead:

```go
return nil
```

For beginner DSA, `[]int{}` is usually clearer.

---

## 4. Maps must be initialized

This works:

```go
seen := make(map[int]int)
```

This does not work for assignment:

```go
var seen map[int]int
seen[10] = 0 // panic
```

Python comparison:

```python
seen = {}
seen[10] = 0
```

In Go, use `make` before assigning to a map.

---

# Clean Mental Model

For DSA in Go:

```text
Use slices like Python lists.
Use maps like Python dictionaries.
Use range like Python enumerate.
Remember append returns the updated slice.
Remember Go slicing shares the backing array.
Track Big-O the same way as in Python.
```

For today, the most important thing is this:

```text
Arrays are fixed.
Slices are flexible.
Most Go DSA uses slices.
Big-O thinking is language-independent.
```

Your first optimized pattern:

```text
Nested loops O(n²)
        ↓
Hash map lookup O(n)
```

This exact idea appears in many problems beyond Two Sum.
