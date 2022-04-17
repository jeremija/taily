# taily

Taily is a configurable system utility that can process messages and act on
them. The goal is to make it able to send alerts on errors.

It currently supports reading streams from:

- `systemd journald`
- `docker`

It is smart enough to remember the previous state and resume upon restart.

Work in progress. Current work is being done on implement more `Processor`s.

See sample config file: [config.yml](config.yml) (subject to change until v2).

# Support

The development of Peer Calls is sponsored by [rondomoon][rondomoon]. If you'd
like enterprise on-site support or become a sponsor, please contact
[hello@rondomoon.com](mailto:hello@rondomoon.com).

[rondomoon]: https://rondomoon.com

# License

[Apache 2.0](LICENSE)
