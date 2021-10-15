package soratun

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Up ups new SORACOM Arc tunnel with given ArcSession.
func Up(ctx context.Context, config *Config) {
	iname := config.Interface

	logger := device.NewLogger(
		config.LogLevel,
		fmt.Sprintf("(%s) ", iname),
	)

	// specified interface name and actual interface name may vary
	t, err := tun.CreateTUN(iname, config.Mtu)
	if err != nil {
		logger.Errorf("failed to create new tunnel: %v", err)
		os.Exit(1)
	}

	actualInterfaceName, err := t.Name()
	if err == nil {
		iname = actualInterfaceName
		// renew the prefix with the actual interface name
		logger = device.NewLogger(
			config.LogLevel,
			fmt.Sprintf("(%s) ", iname),
		)
	}

	d := device.NewDevice(t, conn.NewDefaultBind(), logger)
	err = d.Up()
	if err != nil {
		logger.Errorf("Failed to bring up device: %v", err)
		os.Exit(1)
	}

	logger.Verbosef("device started")

	uapi, err := ipc.UAPIListen(iname)
	if err != nil {
		logger.Errorf("Failed to listen on uapi socket: %v", err)
		os.Exit(1)
	}

	defer func() {
		if err := uapi.Close(); err != nil {
			log.Printf("failed to close UAPI listener: %v ", err)
		}
	}()

	errs := make(chan error)
	go func() {
		for {
			c, err := uapi.Accept()
			if err != nil {
				errs <- err
				return
			}
			go d.IpcHandle(c)
		}
	}()

	logger.Verbosef("UAPI listener started")

	client, err := wgctrl.New()
	if err != nil {
		logger.Errorf("failed to open wgctrl: %v", err)
		d.Close()
		os.Exit(1)
	}

	defer func() {
		if err := client.Close(); err != nil {
			logger.Errorf("failed to close wgctrl: %v", err)
		}
	}()

	var allowedIPs []net.IPNet
	for _, v := range config.ArcSession.ArcAllowedIPs {
		allowedIPs = append(allowedIPs, (net.IPNet)(*v))
	}

	err = client.ConfigureDevice(iname, wgtypes.Config{
		PrivateKey:   config.PrivateKey.AsWgKey(),
		FirewallMark: nil,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: *config.ArcSession.ArcServerPeerPublicKey.AsWgKey(),
				Endpoint: &net.UDPAddr{
					IP:   config.ArcSession.ArcServerEndpoint.IP,
					Port: config.ArcSession.ArcServerEndpoint.Port,
				},
				PersistentKeepaliveInterval: duration(time.Duration(config.PersistentKeepalive) * time.Second),
				ReplaceAllowedIPs:           true,
				AllowedIPs:                  allowedIPs,
			},
		},
	})
	if err != nil {
		logger.Errorf("failed to configure new device %s: %v", iname, err)
		d.Close()
		os.Exit(1)
	}

	if err = ConfigureInterface(iname, config); err != nil {
		logger.Errorf("error: %s\n", err)
		d.Close()
		os.Exit(1)
	}

	if len(config.PostUp) > 0 {
		for i, com := range config.PostUp {
			if len(com) == 0 || com[0] == "" {
				continue
			}

			command := replaceInterfaceName(com, iname)
			logger.Verbosef("executing PostUp(%d): %s", i, command)
			result, err := runCommand(command)
			if err != nil {
				logger.Errorf("failed to do PostUp(%d): %s\n", i, err)
				d.Close()
				os.Exit(1)
			}
			logger.Verbosef("PostUp(%d) response: %s", i, result)
		}
	}

	if config.EnableMetrics {
		go func() {
			ticker := time.NewTicker(time.Second * 60)
			defer ticker.Stop()

			for {
				<-ticker.C
				d, err := client.Device(iname)
				if err == nil {
					for _, p := range d.Peers {
						logger.Verbosef("soratun_sent_bytes_total{simId=\"%s\",interface=\"%s\",endpoint=\"%s:%d\"} %d", config.SimId, d.Name, p.Endpoint.IP, p.Endpoint.Port, p.TransmitBytes)
						logger.Verbosef("soratun_received_bytes_total{simId=\"%s\",interface=\"%s\",endpoint=\"%s:%d\"} %d", config.SimId, d.Name, p.Endpoint.IP, p.Endpoint.Port, p.ReceiveBytes)
						logger.Verbosef("soratun_latest_handshake_epoch{simId=\"%s\",interface=\"%s\",endpoint=\"%s:%d\"} %d", config.SimId, d.Name, p.Endpoint.IP, p.Endpoint.Port, p.LastHandshakeTime.Unix())
					}
				}
			}
		}()
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM)
	signal.Notify(term, syscall.SIGABRT) // systemd will restart the process with SIGABRT when watchdog timer expires
	signal.Notify(term, os.Interrupt)

	select {
	case <-term:
	case <-errs:
	case <-d.Wait():
	case <-ctx.Done():
	}

	d.Close()

	if len(config.PostDown) > 0 {
		for i, com := range config.PostDown {
			if len(com) == 0 || com[0] == "" {
				continue
			}

			command := replaceInterfaceName(com, iname)
			logger.Verbosef("executing PostDown(%d): %s", i, command)
			result, err := runCommand(command)
			if err != nil {
				logger.Errorf("failed to do PostDown(%d): %s\n", i, err)
				os.Exit(1)
			}
			logger.Verbosef("PostDown(%d) response: %s", i, result)
		}
	}

	logger.Verbosef("shutting down")
}

func duration(d time.Duration) *time.Duration { return &d }

// DefaultInterfaceName returns a default interface name
func DefaultInterfaceName() string {
	return "soratun0"
}
