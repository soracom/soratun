package soratun

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const arcServerEndpointDefaultPort string = "11010"

// UDPAddr represents the UDP address with keeping original endpoint.
type UDPAddr struct {
	IP          net.IP
	Port        int
	RawEndpoint []byte
}

// aliases to add custom un/marshaler to each types.
type (
	// Key is an alias for wgtypes.Key.
	Key wgtypes.Key
	// IPNet is an alias for net.IPNet.
	IPNet net.IPNet
)

// Config holds SORACOM Arc client configurations.
type Config struct {
	// PrivateKey is WireGuard private key.
	PrivateKey Key `json:"privateKey"`
	// PublicKey is WireGuard public key.
	PublicKey Key `json:"publicKey"`
	// SimId is virtual SIM's SimId for the connection.
	SimId string `json:"simId"`
	// LogLevel specifies logging level, verbose, error, or silent.
	LogLevel int `json:"logLevel"`
	// If EnableMetrics is true, metrics will be logged when log-level is verbose or error.
	EnableMetrics bool `json:"enableMetrics"`
	// Interface is name for the tunnel interface.
	Interface string `json:"interface"`
	// AdditionalAllowedIPs holds a set of WireGuard allowed IPs in addition to the list which will get while creating Arc session.
	AdditionalAllowedIPs []*IPNet `json:"additionalAllowedIPs,omitempty"`
	// Mtu of the interface.
	Mtu int `json:"mtu,omitempty"`
	// WireGuard PersistentKeepalive parameter.
	PersistentKeepalive int `json:"persistentKeepalive,omitempty"`
	// PostUp is array of commands which will be executed after the interface is up successfully.
	PostUp [][]string `json:"postUp,omitempty"`
	// PostDown is array of commands which will be executed after the interface is removed successfully.
	PostDown [][]string `json:"postDown,omitempty"`
	// Profile is for SORACOM API access.
	Profile *Profile `json:"profile,omitempty"`
	// ArcSession holds connection information provided from SORACOM Arc server.
	ArcSession *ArcSession `json:"arcSessionStatus,omitempty"`
}

// ArcSession holds SORACOM Arc configurations received from the server.
type ArcSession struct {
	// ArcServerPeerPublicKey is WireGuard public key of the SORACOM Arc server.
	ArcServerPeerPublicKey Key `json:"arcServerPeerPublicKey"`
	// ArcServerEndpoint is a UDP endpoint of the SORACOM Arc server.
	ArcServerEndpoint *UDPAddr `json:"arcServerEndpoint"`
	// ArcAllowedIPs holds IP addresses allowed for routing from the SORACOM Arc server.
	ArcAllowedIPs []*IPNet `json:"arcAllowedIPs"`
	// ArcClientPeerPrivateKey holds private key from SORACOM Arc server.
	ArcClientPeerPrivateKey Key `json:"arcClientPeerPrivateKey,omitempty"`
	// ArcClientPeerIpAddress is an IP address for this client.
	ArcClientPeerIpAddress net.IP `json:"arcClientPeerIpAddress,omitempty"`
}

// NewKey returns a Key from a base64-encoded string.
func NewKey(s string) (Key, error) {
	key, err := wgtypes.ParseKey(s)
	if err != nil {
		return Key{}, err
	}
	return Key(key), nil
}

// UnmarshalText decodes a byte array of private key to the Key. If text is invalid WireGuard key, UnmarshalText returns an error.
func (k *Key) UnmarshalText(text []byte) error {
	key, err := wgtypes.ParseKey(string(text))
	if err != nil {
		return err
	}
	copy(k[:], key[:])
	return nil
}

// MarshalText encodes Key to an array of bytes.
func (k *Key) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

// AsWgKey converts Key back to wgtypes.Key.
func (k *Key) AsWgKey() *wgtypes.Key {
	key, _ := wgtypes.NewKey(k[:])
	return &key
}

// AsHexString returns hexadecimal encoding of Key.
func (k *Key) AsHexString() string {
	return hex.EncodeToString(k[:])
}

// String returns string representation of Key.
func (k Key) String() string {
	return k.AsWgKey().String()
}

// UnmarshalText converts a byte array into UDPAddr. UnmarshalText returns error if the format is invalid (not "ip" or "ip:port"), IP address specified is invalid, or the port is not a 16-bit unsigned integer.
func (a *UDPAddr) UnmarshalText(text []byte) error {
	h, p, err := net.SplitHostPort(string(text))
	if err != nil {
		h = string(text)
		p = arcServerEndpointDefaultPort
	}

	var ip net.IP
	ip = net.ParseIP(h)
	if ip == nil {
		ips, err := net.LookupIP(h)
		if err != nil || len(ips) < 1 {
			return fmt.Errorf("invalid endpoint \"%s\": %s", h, err)
		}
		ip = ips[0]
	}

	port, err := strconv.Atoi(p)
	if err != nil || port < 0 || port > 65535 {
		return fmt.Errorf("invalid serverEndpoint port number: %s, it should be a 16-bit unsigned integer", p)
	}

	a.IP, a.Port = ip, port
	a.RawEndpoint = text
	return nil
}

// MarshalText converts struct to a string.
func (a *UDPAddr) MarshalText() ([]byte, error) {
	if len(a.RawEndpoint) <= 0 {
		return []byte(fmt.Sprintf("%s:%d", a.IP, a.Port)), nil
	}
	return a.RawEndpoint, nil
}

// UnmarshalText converts a byte array into IPNet. UnmarshalText returns error if invalid CIDR is provided.
func (n *IPNet) UnmarshalText(text []byte) error {
	_, ipnet, err := net.ParseCIDR(string(text))
	if err != nil {
		return err
	}

	n.IP, n.Mask = ipnet.IP, ipnet.Mask
	return nil
}

// MarshalText converts struct to a string.
func (n *IPNet) MarshalText() ([]byte, error) {
	prefix, _ := n.Mask.Size()
	return []byte(fmt.Sprintf("%s/%d", n.IP, prefix)), nil
}

// MarshalJSON converts struct to JSON, omitting ArcClientPeerPrivateKey field which is redundant for configuration file.
func (a *ArcSession) MarshalJSON() ([]byte, error) {
	var tmp struct {
		ArcServerPeerPublicKey Key      `json:"arcServerPeerPublicKey"`
		ArcServerEndpoint      *UDPAddr `json:"arcServerEndpoint"`
		ArcAllowedIPs          []*IPNet `json:"arcAllowedIPs"`
		ArcClientPeerIpAddress net.IP   `json:"arcClientPeerIpAddress"`
	}
	tmp.ArcServerPeerPublicKey = a.ArcServerPeerPublicKey
	tmp.ArcServerEndpoint = a.ArcServerEndpoint
	tmp.ArcClientPeerIpAddress = a.ArcClientPeerIpAddress
	tmp.ArcAllowedIPs = a.ArcAllowedIPs
	return json.Marshal(&tmp)
}
