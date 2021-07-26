package soratun

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/soracom/soratun/internal"
)

// A SoracomClient represents an API client for SORACOM API. See
// https://developers.soracom.io/en/docs/tools/api-reference/ or
// https://dev.soracom.io/jp/docs/api_guide/
//go:generate mockgen -source client.go -destination internal/mock/client.go
type SoracomClient interface {
	CreateVirtualSim() (*VirtualSim, error)
	CreateArcSession(simId, publicKey string) (*ArcSession, error)
	SetVerbose(v bool)
	Verbose() bool
}

// DefaultSoracomClient is an implementation of the SoracomClient for the general use case.
type DefaultSoracomClient struct {
	apiKey   string       // SORACOM API key.
	token    string       // SORACOM API token.
	endpoint string       // SORACOM API endpoint.
	client   *http.Client // HTTP client.
	verbose  bool
}

// A Profile holds SORACOM API client related information.
type Profile struct {
	// AuthKey is SORACOM API auth key secret.
	AuthKey string `json:"authKey,omitempty"`
	// AuthKeyID is SORACOM API auth key ID.
	AuthKeyID string `json:"authKeyId,omitempty"`
	// Endpoint is SORACOM API endpoint.
	Endpoint string `json:"endpoint,omitempty"`
}

type apiParams struct {
	body   string
	method string
	path   string
}

// VirtualSim represents virtual subscriber.
type VirtualSim struct {
	// OperatorId is operator ID of the subscriber.
	OperatorId string `json:"operatorId"`
	// Status is virtual SIM status, active or terminated as of 2021 first release.
	Status string `json:"status"`
	// SimId is SIM ID of the subscriber.
	SimId string `json:"simId"`
	// ArcSession holds Arc connection information.
	ArcSession ArcSession `json:"arcSessionStatus"`
	// Profiles holds series of SimProfile, (not SORACOM API Profile).
	Profiles map[string]SimProfile `json:"profiles"`
}

// SimProfile is a SIM profile which holds one of profiles in the subscription container.
type SimProfile struct {
	// Iccid is ICCID of the subscriber.
	Iccid string `json:"iccid"`
	// ArcClientPeerPrivateKey is WireGuard private key of the subscriber.
	ArcClientPeerPrivateKey string `json:"arcClientPeerPrivateKey"`
	// ArcClientPeerPublicKey is WireGuard public key of the subscriber.
	ArcClientPeerPublicKey string `json:"arcClientPeerPublicKey"`
	// PrimaryImsi is Imsi of this virtual SIM.
	PrimaryImsi string `json:"primaryImsi"`
}

// NewDefaultSoracomClient returns new SoracomClient for caller.
func NewDefaultSoracomClient(p Profile) (SoracomClient, error) {
	authKeyId := p.AuthKeyID
	if authKeyId == "" || !strings.HasPrefix(authKeyId, "keyId-") {
		return nil, fmt.Errorf("invalid AuthKeyId \"%s\". It must starts with \"keyId-\"", authKeyId)
	}

	authKey := p.AuthKey
	if authKey == "" || !strings.HasPrefix(authKey, "secret-") {
		return nil, fmt.Errorf("invalid AuthKey \"%s\". It must starts with \"secret-\"", authKey)
	}

	endpoint := p.Endpoint
	if endpoint == "" {
		endpoint = "https://api.soracom.io"
	}

	c := DefaultSoracomClient{
		apiKey:   "",
		token:    "",
		endpoint: endpoint,
		client:   http.DefaultClient,
		verbose:  false,
	}

	body, err := json.Marshal(struct {
		AuthKeyID           string `json:"authKeyId"`
		AuthKey             string `json:"authKey"`
		TokenTimeoutSeconds int    `json:"tokenTimeoutSeconds"`
	}{
		AuthKeyID:           authKeyId,
		AuthKey:             authKey,
		TokenTimeoutSeconds: 5 * 60,
	})
	if err != nil {
		return nil, err
	}

	res, err := c.callAPI(&apiParams{
		method: "POST",
		path:   "/auth",
		body:   string(body),
	})
	if err != nil {
		return nil, err
	}

	ar := struct {
		APIKey string `json:"apiKey"`
		Token  string `json:"token"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(&ar); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.apiKey = ar.APIKey
	c.token = ar.Token
	return &c, nil
}

// SetVerbose sets if verbose output is enabled or not.
func (c *DefaultSoracomClient) SetVerbose(v bool) {
	c.verbose = v
}

// Verbose returns if verbose output is enabled or not.
func (c *DefaultSoracomClient) Verbose() bool {
	return c.verbose
}

// CreateVirtualSim creates new virtual SIM.
func (c *DefaultSoracomClient) CreateVirtualSim() (*VirtualSim, error) {
	body, err := json.Marshal(struct {
		Type         string `json:"type"`
		Subscription string `json:"subscription"`
	}{
		Type:         "virtual",
		Subscription: "planArc01",
	})
	if err != nil {
		return nil, err
	}

	res, err := c.callAPI(&apiParams{
		method: "POST",
		path:   "/sims",
		body:   string(body),
	})
	if err != nil {
		return nil, err
	}

	var subscriber VirtualSim
	err = json.NewDecoder(res.Body).Decode(&subscriber)
	return &subscriber, err
}

// CreateArcSession creates new Arc session.
func (c *DefaultSoracomClient) CreateArcSession(simId, publicKey string) (*ArcSession, error) {
	// bootstrapped SIM will have attached credential. So we can just sent empty object.
	res, err := c.callAPI(&apiParams{
		method: "POST",
		path:   "/sims/" + simId + "/sessions/arc",
		body:   "{}",
	})
	if err != nil {
		return nil, err
	}

	var session ArcSession
	err = json.NewDecoder(res.Body).Decode(&session)
	return &session, err
}

func (c *DefaultSoracomClient) callAPI(params *apiParams) (*http.Response, error) {
	req, err := c.makeRequest(params)
	if err != nil {
		return nil, err
	}

	if c.Verbose() {
		fmt.Fprintln(os.Stderr, "--- Request dump ---------------------------------")
		r, _ := httputil.DumpRequest(req, true)
		fmt.Fprintf(os.Stderr, "%s\n", r)
		fmt.Fprintln(os.Stderr, "--- End of request dump --------------------------")
	}
	res, err := c.doRequest(req)
	return res, err
}

func (c *DefaultSoracomClient) makeRequest(params *apiParams) (*http.Request, error) {
	var body io.Reader
	if params.body != "" {
		body = strings.NewReader(params.body)
	}

	req, err := http.NewRequest(params.method,
		fmt.Sprintf("%s/v1%s", c.endpoint, params.path),
		body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Soracom-Lang", "en")
	req.Header.Set("User-Agent", internal.UserAgent)
	if c.apiKey != "" {
		req.Header.Set("X-Soracom-Api-Key", c.apiKey)
	}
	if c.token != "" {
		req.Header.Set("X-Soracom-Token", c.token)
	}
	return req, nil
}

func (c *DefaultSoracomClient) doRequest(req *http.Request) (*http.Response, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if c.Verbose() && res != nil {
		fmt.Fprintln(os.Stderr, "--- Response dump --------------------------------")
		r, _ := httputil.DumpResponse(res, true)
		fmt.Fprintf(os.Stderr, "%s\n", r)
		fmt.Fprintln(os.Stderr, "--- End of response dump -------------------------")
	}

	if res.StatusCode >= http.StatusBadRequest {
		defer func() {
			err := res.Body.Close()
			if err != nil {
				fmt.Println("failed to close response", err)
			}
		}()
		r, _ := ioutil.ReadAll(res.Body)
		return res, fmt.Errorf("%s: %s %s: %s", res.Status, req.Method, req.URL, r)
	}
	return res, nil
}
