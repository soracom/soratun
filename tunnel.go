// +build !windows

// Package soratun implements userspace SORACOM Arc client.
package soratun

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// https://www.wireguard.com/protocol/
// > After receiving a packet, if the receiver was the original initiator of the handshake
// > and if the current session key is `REKEY_AFTER_TIME - KEEPALIVE_TIMEOUT - REKEY_TIMEOUT` ms old,
// > we initiate a new handshake.
// As of golang.zx2c4.com/wireguard v0.0.0-20210604143328-f9b48a961cd2, it will be 120 - 10 - 5 = 105 seconds.
// Added 5 seconds without any concrete theory
var watchdogTimeout = device.RekeyAfterTime - device.KeepaliveTimeout - device.RekeyTimeout + 5*time.Second

const (
	// LogLevelVerbose is an alias for WireGuard device logger equivalent.
	LogLevelVerbose = device.LogLevelVerbose
	// LogLevelError is an alias for WireGuard device logger equivalent.
	LogLevelError = device.LogLevelError
	// LogLevelSilent is an alias for WireGuard device logger equivalent.
	LogLevelSilent = device.LogLevelSilent
	// DefaultPersistentKeepaliveInterval defines WireGuard persistent keepalive interval to SORACOM Arc.
	DefaultPersistentKeepaliveInterval = 60
	// DefaultMTU is MTU for the configured interface.
	DefaultMTU = device.DefaultMTU
)

// Up ups new SORACOM Arc tunnel with given ArcSession.
func Up(ctx context.Context, config *Config) {
	iname := config.Interface

	logger := device.NewLogger(
		config.LogLevel,
		fmt.Sprintf("(%s) ", iname),
	)

	if isWatchdogEnabled() {
		logger.Verbosef("systemd watchdog is available. Will update watchdog timer every %s seconds", watchdogTimeout)
		_, err := daemon.SdNotify(false, daemon.SdNotifyReloading)
		if err != nil {
			logger.Errorf("failed to notify reloading to systemd")
		}
	}

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

	logger.Verbosef("device started")

	fileUAPI, err := ipc.UAPIOpen(iname)
	if err != nil {
		logger.Errorf("UAPI listen error: %v", err)
		os.Exit(1)
	}

	uapi, err := ipc.UAPIListen(iname, fileUAPI)
	if err != nil {
		logger.Errorf("failed to listen on UAPI socket: %v", err)
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
				PublicKey:                   *config.ArcSession.ArcServerPeerPublicKey.AsWgKey(),
				Endpoint:                    (*net.UDPAddr)(config.ArcSession.ArcServerEndpoint),
				PersistentKeepaliveInterval: duration(time.Duration(config.PersistentKeepalive) * time.Second),
				ReplaceAllowedIPs:           true,
				AllowedIPs:                  allowedIPs,
			},
		},
	})
	if err != nil {
		logger.Errorf("failed to configure new device %s: %v", iname, err)
		os.Exit(1)
	}

	if err = ConfigureInterface(iname, config); err != nil {
		logger.Errorf("error: %s\n", err)
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
			}
			logger.Verbosef("PostUp(%d) response: %s", i, result)
		}
	}

	if isWatchdogEnabled() {
		_, err = daemon.SdNotify(false, daemon.SdNotifyReady)
		if err != nil {
			logger.Errorf("failed to notify ready to systemd")
		}
		go func() {
			ticker := time.NewTicker(watchdogTimeout)
			defer ticker.Stop()

			for {
				<-ticker.C
				d, err := client.Device(iname)
				if err != nil {
					logger.Errorf("failed to update watchdog timer to systemd")
				} else {
					for _, p := range d.Peers {
						if time.Since(p.LastHandshakeTime) < watchdogTimeout {
							_, err := daemon.SdNotify(false, daemon.SdNotifyWatchdog)
							if err != nil {
								logger.Errorf("failed to update watchdog timer to systemd")
							} else {
								logger.Verbosef("update watchdog timer")
							}
						}
					}
				}
			}
		}()
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
						logger.Errorf("soratun_sent_bytes_total{simId=\"%s\",interface=\"%s\",endpoint=\"%s:%d\"} %d", config.SimId, d.Name, p.Endpoint.IP, p.Endpoint.Port, p.TransmitBytes)
						logger.Errorf("soratun_received_bytes_total{simId=\"%s\",interface=\"%s\",endpoint=\"%s:%d\"} %d", config.SimId, d.Name, p.Endpoint.IP, p.Endpoint.Port, p.ReceiveBytes)
						logger.Errorf("soratun_latest_handshake_epoch{simId=\"%s\",interface=\"%s\",endpoint=\"%s:%d\"} %d", config.SimId, d.Name, p.Endpoint.IP, p.Endpoint.Port, p.LastHandshakeTime.Unix())
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
			}
			logger.Verbosef("PostDown(%d) response: %s", i, result)
		}
	}

	logger.Verbosef("shutting down")
}

func duration(d time.Duration) *time.Duration { return &d }

func isWatchdogEnabled() bool {
	enabled, _ := daemon.SdWatchdogEnabled(false)
	return enabled != 0
}

// DefaultInterfaceName returns a default interface name
func DefaultInterfaceName() string {
	iname := "soratun0"
	if runtime.GOOS == "darwin" {
		iname = "utun"
	}
	return iname
}

func runCommand(c []string) (string, error) {
	result, err := exec.Command(c[0], c[1:]...).CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("error while running \"%s\"", strings.Join(c, " "))
	}

	return fmt.Sprintf("'%s'\n", strings.TrimSpace(string(result))), nil
}

func replaceInterfaceName(command []string, iname string) []string {
	var replaced []string
	for _, s := range command {
		replaced = append(replaced, strings.Replace(s, "%i", iname, -1))
	}
	return replaced
}
