# soratun configuration schema

Configuration schema for `soratun`, SORACOM Arc client (default file name `arc.json`). You can manually edit any properties, but inconsistent modification might be resulted in connection failure. Use `soratun bootstrap` command as possible as you can to update the configuration.

## Properties

| Property                      | Type                        | Required | Description                                                                                                                                                                                |
|-------------------------------|-----------------------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `enableMetrics`               | boolean                     | **Yes**  | Enable metrics logging every 60 seconds, if logLevel is error (1) or verbose (2)                                                                                                           |
| `interface`                   | string                      | **Yes**  | Interface name. if you are testing on macOS, the interface name must be "utun[0-9]+" for an explicit interface name, or just "utun" to have the kernel select the lowest available number. |
| `logLevel`                    | integer                     | **Yes**  | Logging level (0: silent / 1: error / 2: verbose)                                                                                                                                          |
| `privateKey`                  | string                      | **Yes**  | WireGuard private key. Do not modify this unless you know what you are doing                                                                                                               |
| `publicKey`                   | string                      | **Yes**  | WireGuard public key. Do not modify this unless you know what you are doing                                                                                                                |
| `additionalAllowedIPs`        | string[]                    | No       | Array of additional WireGuard allowed CIDRs                                                                                                                                                |
| `arcSessionStatus`            | [object](#arcsessionstatus) | No       | SORACOM Arc connection information. Usually you should not edit this property manually.                                                                                                    |
| `mtu`                         | number                      | No       | MTU for the interface                                                                                                                                                                      |
| `persistentKeepaliveInterval` | number                      | No       | WireGuard `PersistentKeepaliveInterval` for the SORACOM Arc server                                                                                                                         |
| `profile`                     | [object](#profile)          | No       | SORACOM API client information. Saved if you use `soratun bootstrap authkey` command. Other bootstrap methods don't use this.                                                              |
| `simId`                       | string                      | No       | SIM ID of your virtual SIM                                                                                                                                                                 |

## arcSessionStatus

SORACOM Arc connection information. Usually you should not edit this property manually.

### Properties

| Property                 | Type     | Required | Description                                                              |
|--------------------------|----------|----------|--------------------------------------------------------------------------|
| `arcAllowedIPs`          | string[] | **Yes**  | An array of CIDRs allowed for routing from the SORACOM Arc server        |
| `arcClientPeerIpAddress` | string   | **Yes**  | An IP address for this client                                            |
| `arcServerEndpoint`      | string   | **Yes**  | A UDP endpoint of the SORACOM Arc server in `ip or hostname:port` format |
| `arcServerPeerPublicKey` | string   | **Yes**  | WireGuard public key of the SORACOM Arc server                           |

## profile

SORACOM API client information. Saved if you use `soratun bootstrap authkey` command. Other bootstrap methods don't use this.

### Properties

| Property    | Type   | Required | Description                                                                                              |
|-------------|--------|----------|----------------------------------------------------------------------------------------------------------|
| `authKeyId` | string | **Yes**  | SORACOM API auth key                                                                                     |
| `authKey`   | string | **Yes**  | SORACOM API auth key secret                                                                              |
| `endpoint`  | string | **Yes**  | SORACOM API endpoint. Global coverage: https://g.api.soracom.io / Japan coverage: https://api.soracom.io |

