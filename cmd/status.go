package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"log"
	"strings"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"s"},
		Short:   "Display SORACOM Arc interface status",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := wgctrl.New()
			if err != nil {
				log.Fatalf("failed to open wgctrl: %v", err)
			}

			defer func() {
				err := c.Close()
				if err != nil {
					log.Printf("failed to close wgctrl: %v ", err)
				}
			}()

			var devices []*wgtypes.Device
			devices, err = c.Devices()
			if err != nil {
				log.Fatalf("failed to get devices: %v", err)
			}

			if len(devices) == 0 {
				fmt.Println("no SORACOM Arc device found")
			}
			for _, d := range devices {
				printDevice(d)

				for _, p := range d.Peers {
					printPeer(p)
				}
			}
		},
	}
}

func printDevice(d *wgtypes.Device) {
	const f = `interface: %s (%s)
  public key: %s
  private key: (hidden)
  listening port: %d

`

	fmt.Printf(
		f,
		d.Name,
		d.Type.String(),
		d.PublicKey.String(),
		//d.PrivateKey.String(),
		d.ListenPort)
}

func printPeer(p wgtypes.Peer) {
	const f = `peer: %s
  endpoint: %s
  allowed ips: %s
  latest handshake: %s
  transfer: %d B received, %d B sent

`

	ips := make([]string, 0, len(p.AllowedIPs))
	for _, ip := range p.AllowedIPs {
		ips = append(ips, ip.String())
	}

	fmt.Printf(
		f,
		p.PublicKey.String(),
		p.Endpoint.String(),
		strings.Join(ips, ", "),
		p.LastHandshakeTime.String(),
		p.ReceiveBytes,
		p.TransmitBytes,
	)
}
