# soratun 設定ファイルスキーマ

SORACOM Arc クライアント soratun の設定ファイルのスキーマです。手動で編集できますが一貫性のない変更を行った場合接続できなくなる可能性があります。可能な限り `soratun bootstrap` コマンドを使用してください。

## Properties

| Property               | Type                        | Required | Description                                                                                                                    |
|------------------------|-----------------------------|----------|--------------------------------------------------------------------------------------------------------------------------------|
| `enableMetrics`        | boolean                     | **Yes**  | 有効にした場合、ログレベルが `error` または `verbose` の際に標準出力にメトリックスを約 60 秒毎に出力します。                   |
| `interface`            | string                      | **Yes**  | soratun が作成するインターフェース名。macOS でテストする場合、OS の制限のため `utun` で始まる文字列を指定してください。        |
| `logLevel`             | integer                     | **Yes**  | ログレベル (0: 出力無し / 1: エラーのみ出力 / 2: デバッグ情報も出力)                                                           |
| `privateKey`           | string                      | **Yes**  | WireGuard 秘密鍵。通常は編集しないでください。                                                                                 |
| `publicKey`            | string                      | **Yes**  | WireGuard 公開鍵。通常は編集しないでください。                                                                                 |
| `additionalAllowedIPs` | string[]                    | No       | soratun 作成時に WireGuard の AllowedIPs に追加する CIDR の配列。このネットワーク宛の通信も `soratun` 経由になります。         |
| `arcSessionStatus`     | [object](#arcsessionstatus) | No       | SORACOM Arc 接続情報。自動的に生成または更新されますので通常は編集しないでください。                                           |
| `profile`              | [object](#profile)          | No       | SORACOM API 接続情報。`soratun bootstrap authkey` を実行した際に保存されます。その他のブートストラップ方法では使用されません。 |
| `simId`                | string                      | No       | バーチャル SIM の SIM ID                                                                                                       |

## arcSessionStatus

SORACOM Arc 接続情報。自動的に生成または更新されますので通常は編集しないでください。

### Properties

| Property                 | Type     | Required | Description                                                                        |
|--------------------------|----------|----------|------------------------------------------------------------------------------------|
| `arcAllowedIPs`          | string[] | **Yes**  | SORACOM Arc サーバーから受信した WireGuard AllowedIPs の配列                       |
| `arcClientPeerIpAddress` | string   | **Yes**  | クライアントの IP アドレス                                                         |
| `arcServerEndpoint`      | string   | **Yes**  | SORACOM Arc サーバーの UDP エンドポイント (`IP アドレスまたはホスト名:ポート番号`) |
| `arcServerPeerPublicKey` | string   | **Yes**  | SORACOM Arc サーバーの WireGuard 公開鍵                                            |

## profile

SORACOM API 接続情報。`soratun bootstrap authkey` を実行した際に保存されます。その他のブートストラップ方法では使用されません。

### Properties

| Property    | Type   | Required | Description                                                                                                          |
|-------------|--------|----------|----------------------------------------------------------------------------------------------------------------------|
| `authKeyId` | string | **Yes**  | SORACOM API 認証キー ID                                                                                              |
| `authKey`   | string | **Yes**  | SORACOM API 認証キーシークレット                                                                                     |
| `endpoint`  | string | **Yes**  | SORACOM API のエンドポイント。Global カバレッジ: https://g.api.soracom.io / Japan カバレッジ: https://api.soracom.io |

