# driftwatch

A daemon that monitors infrastructure config files for unexpected drift and alerts via webhook or Slack.

---

## Installation

```bash
go install github.com/yourorg/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/driftwatch.git && cd driftwatch && go build -o driftwatch .
```

---

## Usage

Create a config file (`driftwatch.yaml`) and start the daemon:

```yaml
interval: 60s
paths:
  - /etc/nginx/nginx.conf
  - /etc/app/config.toml
alerts:
  slack:
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  webhook:
    url: "https://your-alerting-endpoint.example.com/notify"
```

```bash
driftwatch --config driftwatch.yaml
```

driftwatch will compute checksums of the specified files on startup, then re-check them at the configured interval. If a file changes unexpectedly, an alert is sent to your configured Slack channel or webhook endpoint.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `driftwatch.yaml` | Path to config file |
| `--once` | `false` | Run a single check and exit |
| `--log-level` | `info` | Log verbosity (`debug`, `info`, `warn`, `error`) |

---

## Contributing

Pull requests and issues are welcome. Please open an issue before submitting large changes.

---

## License

MIT © yourorg