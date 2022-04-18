# taily

Taily is a configurable system utility that can process messages and act on
them. The goal is to make it able to send alerts on errors.

# Features

- Read log streams from `systemd journald` and `docker`
- Resume reading after a shutdown
- Create custom processing rules
- Send notifications when certain matches are found (e.g. Slack or Telegram)

Work in progress.

# Configuration

See sample config file: [config.yml](config.yml) (subject to change until v2).

# Useful commands

| Action  | Command                                                        |
|---------|----------------------------------------------------------------|
| Build   | `make` or `go build -o bin/taily cmd`                          |
| Test    | `make test` or `go test /...`                                  |
| Run     | `TAILY_CONFIG="$(cat config.yml)" go run ./cmd`                |
| Run (2) | `TAILY_CONFIG="$(cat config.yml)" bin/taily`                   |
| Logs    | `TAILY_LOG='**' TAILY_CONFIG="$(cat config.yml)" bin/taily`    |

The docker client can currently be configured with the default environment
variables:

- `DOCKER_HOST` to set the url to the docker server.
- `DOCKER_API_VERSION` to set the version of the API to reach, leave empty for latest.
- `DOCKER_CERT_PATH` to load the TLS certificates from.
- `DOCKER_TLS_VERIFY` to enable or disable TLS verification, off by default.

# Support

The development of Peer Calls is sponsored by [rondomoon][rondomoon]. If you'd
like enterprise on-site support or become a sponsor, please contact
[hello@rondomoon.com](mailto:hello@rondomoon.com).

[rondomoon]: https://rondomoon.com

# License

[Apache 2.0](LICENSE)
