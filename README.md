# soratun [![check](https://github.com/soracom/soratun/actions/workflows/check.yml/badge.svg)](https://github.com/soracom/soratun/actions/workflows/check.yml)

An easy-to-use, userspace [SORACOM Arc](https://soracom.jp/services/arc/) client powered by [wireguard-go](https://git.zx2c4.com/wireguard-go/about/). For deploying and scaling Linux servers/Raspberry Pi devices working with SORACOM platform and SORACOM Arc.

- Quick deployment (copy one or two binary files and done)
- Integration with SORACOM platform

## Tested Platforms

- Linux amd64
  - Ubuntu 20.04.2 LTS
- Linux arm (Raspberry Pi 32-bit)
  - Raspberry Pi OS 2021-05-07
  - Ubuntu 20.04.2 LTS
- macOS Big Sur 11.3 (20E232) -- **For development and testing purpose only**

## Usage

```
soratun -- SORACOM Arc Client

Usage:
  soratun [command]

Available Commands:
  bootstrap   Create virtual SIM and configure soratun
  config      Create initial soratun configuration file without bootstrapping
  help        Help about any command
  status      Display SORACOM Arc interface status
  up          Setup SORACOM Arc interface
  version     Show version
  wg-config   Dump soratun configuration file as WireGuard format

Flags:
      --config string   Specify path to SORACOM Arc client configuration file (default "arc.json")
  -h, --help            help for soratun

Use "soratun [command] --help" for more information about a command.
```

See the schema ([English](./docs/config.en.md) / [Japanese](./docs/config.ja.md)) for configuration file `arc.json` detail.

### Getting Started

1. SORACOM platform setup

   1. Create a new SAM user with following permission, and generate a pair of **Auth Key** and **Auth Key Secret** for SORACOM API, referring following official documents:

      - [Users & Roles](https://developers.soracom.io/en/docs/security/users-and-roles/) / [Soracom API Reference Guide](https://developers.soracom.io/en/docs/tools/api-reference/#generating-an-api-key-and-token)
      - [アクセス管理 (SORACOM Access Management)](https://users.soracom.io/ja-jp/docs/sam/) / [API キーと API トークン](https://users.soracom.io/ja-jp/tools/api/key-and-token/)

      ```json
      {
        "statements": [
          {
            "api": ["Sim:createSim", "Sim:createArcSession"],
            "effect": "allow"
          }
        ]
      }
      ```

2. SORACOM Arc bootstrap -- create a new virtual SIM and `soratun` configuration file

   1. Download the latest binary from the [Releases](https://github.com/soracom/soratun/releases/) section.
   2. Bootstrap with `soratun bootstrap authkey` command:

      ```console
      $ ./soratun bootstrap authkey
      ```

   3. `soratun` will guide your setup through interactive wizard, with asking following questions:

      - **SORACOM API auth key ID (starts with "keyId-")**
      - **SORACOM API auth key (starts with "secret-")**
      - **Coverage to create a new virtual SIM** Global coverage (g.api.soracom.io) / Japan coverage (api.soracom.io)

   4. You will get following output from `soratun`:

      ```
      Created new virtual subscriber: 99999xxxxxxxxxx
      Created/updated configuration file: /path/to/arc.json
      ```

3. Start `soratun` to connect to SORACOM platform:

   ```console
   $ sudo ./soratun up
   $ ping pong.soracom.io
   ```

Tips: you can skip interactive wizard by supplying required parameters via flags as follows.

```console
$ soratun bootstrap authkey --auth-key-id keyId-xxx --auth-key secret-xxx --coverage-type jp
```

For other bootstrapping method detail, please consult SORACOM documentation at:

- English: https://developers.soracom.io/en/
- Japanese: https://users.soracom.io/ja-jp/docs/arc/

### Running as a daemon with `systemd`

Use [`conf/soratun.service.sample`](conf/soratun.service.sample) as a starter, copy file you edited to `/etc/systemd/system/soratun.service` directory, then

```console
$ sudo systemctl enable soratun
$ sudo systemctl start soratun
$ sudo systemctl status soratun
$ journalctl -u soratun -f
$ sudo systemctl stop soratun
```

`soratun` supports systemd watchdog. It'll update the timer every 110 seconds, based on [Protocol & Cryptography - WireGuard](https://www.wireguard.com/protocol/):

> After receiving a packet, if the receiver was the original initiator of the handshake and if the current session key is `REKEY_AFTER_TIME - KEEPALIVE_TIMEOUT - REKEY_TIMEOUT` ms old, we initiate a new handshake.

With the sample unit configuration, `soratun` will be restarted after max. 120 + 110 seconds after Arc session deletion. This timer would be reconsidered in the future.

### Running without `sudo`

You can run `soratun` without `sudo` as follows. See `capabilities(7)` for `CAP_NET_ADMIN` detail.

```console
$ sudo groupadd wg # create a new group for WireGuard users
$ sudo mkdir -p /var/run/wireguard # create a directory where wireguard-go control socket file persists
$ sudo chgrp wg /var/run/wireguard # change group of the directory
$ sudo setcap cap_net_admin+epi soratun # add CAP_NET_ADMIN capability to perform various network related operations
$ sudo usermod -a -G wg ubuntu # update group for WireGuard user
$ # log out to enable group change
$ soratun up soratun0 --log-level verbose
```

Note: Some OSes won't persist `/var/run/wireguard` during OS recycle. We have to find more good way to do this.

## TODOs

More test coverage. At this moment [`cmd/up_test.go`](cmd/up_test.go) won't work.

## License

See [LICENSE](LICENSE) for detail.

## Acknowledgments

"WireGuard" and the "WireGuard" logo are registered trademarks of Jason A. Donenfeld.

Following source codes have been derived from [wireguard-go](https://git.zx2c4.com/wireguard-go/) which is copyrighted by WireGuard LLC under the terms of the [license](https://git.zx2c4.com/wireguard-go/tree/LICENSE).

- [`tunnel.go`](tunnel.go)
- [`cmd/status.go`](cmd/status.go)

## Note

This project is not affiliated with the WireGuard project.

## Contributing

Please see [CONTRIBUTING](docs/CONTRIBUTING.md) for detail.

## FAQ

- **Why I should use this although recent Linux kernel has first class WireGuard support?**
  1. In order to make SORACOM Arc deployment and configuration simple. This client is tightly integrated with SORACOM platform and will do essential steps such as authentication, configuration, etc. on behalf of you.
  2. Of course, you can use Linux kernel-native WireGuard with manual setup.
  3. In the future, SORACOM Arc might introduce new technology other than WireGuard, but (hopefully) this program will make the changes invisible under the hood.
- **Why the operating system XXX is not supported?**
  1. [Platforms supported by wireguard-go](https://git.zx2c4.com/wireguard-go/about/) should work. But please note that this client will set required network configurations up via Netlink. Other platform which don't have similar capability will need manual adjustments.
  2. SORACOM API is platform agnostic so implementation related to that part should work as well, but not tested.
