package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/soracom/soratun"
	"github.com/stretchr/testify/assert"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func Test_upCmd(t *testing.T) {
	assertSkip(t)
	setNoDynamicClientSetupEnvVar(t)

	wgClientPrivateKey, err := wgtypes.GeneratePrivateKey()
	assert.NoError(t, err)
	imsi := "999999XXXXXXXXX"

	serverEndpoint, _ := net.ResolveUDPAddr("udp", "192.0.2.2:22212")
	clientPeerIPAddress := net.ParseIP("198.51.100.2")
	_, allowedIPNet, _ := net.ParseCIDR("203.0.113.0/24")
	wgServerPrivateKey, _ := wgtypes.GeneratePrivateKey()
	wgServerPublicKeyBytes, _ := base64.StdEncoding.DecodeString(wgServerPrivateKey.PublicKey().String())
	var wgServerPublicKey soratun.Key
	copy(wgServerPublicKey[:], wgServerPublicKeyBytes[:wgtypes.KeyLen])

	arcConf, err := os.CreateTemp("", "arc.conf")
	assert.NoError(t, err)

	confJSON, err := json.Marshal(map[string]interface{}{
		"privateKey": wgClientPrivateKey.String(),
		"publicKey":  wgClientPrivateKey.PublicKey().String(),
		"imsi":       imsi,
		"interface":  "soratun0",
		"logLevel":   2,
		"profile": map[string]string{
			"authKey":   "secret-xxx",
			"authKeyId": "keyId-xxx",
			"endpoint":  "https://api.soracom.io",
		},
		"arcSessionStatus": map[string]interface{}{
			"ArcServerPeerPublicKey": wgServerPrivateKey.PublicKey().String(),
			"ArcServerEndpoint":      serverEndpoint.String(),
			"ArcAllowedIPs":          []string{allowedIPNet.String()},
			"ArcClientPeerIpAddress": clientPeerIPAddress.String(),
		},
	})
	assert.NoError(t, err)

	_, err = arcConf.Write(confJSON)
	assert.NoError(t, err)

	cancellableCtx, cancel := context.WithCancel(context.Background())
	ctx = cancellableCtx

	os.Args = []string{"__soratun__", "up", "--config", arcConf.Name()}
	go func() {
		err = RootCmd.Execute()
		assert.NoError(t, err)
	}()

	wgCtrl, err := wgctrl.New()
	assert.NoError(t, err)

	isDeviceAcquired := false
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(1000 * time.Millisecond) // XXX: heuristic!!!!!!
		device, err := wgCtrl.Device("soratun0")
		if err != nil || device == nil {
			time.Sleep(2000 * time.Millisecond) // XXX: heuristic!!
			continue
		}
		isDeviceAcquired = true

		assert.NoError(t, err)
		assert.Equal(t, wgClientPrivateKey.String(), base64.StdEncoding.EncodeToString(device.PrivateKey[:]))
		assert.Equal(t, wgClientPrivateKey.PublicKey().String(), base64.StdEncoding.EncodeToString(device.PublicKey[:]))
		assert.NotZero(t, device.ListenPort)

		assert.Len(t, device.Peers, 1)
		peer := device.Peers[0]
		assert.Equal(t, wgServerPublicKey.String(), peer.PublicKey.String())
		assert.Equal(t, serverEndpoint, peer.Endpoint)
		assert.Equal(t, []net.IPNet{*allowedIPNet}, peer.AllowedIPs)
		break
	}

	assert.True(t, isDeviceAcquired)
	cancel()

	isDeviceRemoved := false
	for i := 0; i < maxAttempts; i++ {
		device, _ := wgCtrl.Device("soratun0")
		if device != nil {
			time.Sleep(2000 * time.Millisecond) // XXX: heuristic!!
			continue
		}
		isDeviceRemoved = true
		break
	}
	assert.True(t, isDeviceRemoved)
}

func assertSkip(t *testing.T) {
	if os.Getenv("WG_INTEG_TEST") == "" {
		fmt.Println("INFO: skips the WireGuard integration testing. if you'd like to enable the integration testing, please set the environment value `WG_INTEG_TEST` with non-empty value")
		t.Skip()
	}
}

func setNoDynamicClientSetupEnvVar(t *testing.T) {
	const noDynamicClientSetupEnvVarName = "__SORACOM_NO_DYNAMIC_CLIENT_SETUP_FOR_TEST"

	err := os.Setenv(noDynamicClientSetupEnvVarName, "enabled")
	assert.NoError(t, err)

	t.Cleanup(func() {
		// drop the envvar after a test
		_ = os.Setenv(noDynamicClientSetupEnvVarName, "")
	})
}
