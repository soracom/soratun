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

// A SoracomKryptonClient represents a maybe-over-complicated API client for SORACOM Krypton Provisioning API. See
// https://developers.soracom.io/en/api/krypton/
// https://users.soracom.io/ja-jp/tools/krypton-api/
type SoracomKryptonClient interface {
	Bootstrap() (*ArcSession, error)
	SetVerbose(v bool)
	Verbose() bool
}

// A KryptonClientConfig holds SORACOM Krypton provisioning API client related information.
type KryptonClientConfig struct {
	Endpoint string
}

// DefaultSoracomKryptonClient is an implementation of the SoracomKryptonClient for the general use case.
type DefaultSoracomKryptonClient struct {
	endpoint string       // SORACOM Krypton provisioning API endpoint
	client   *http.Client // HTTP client
	verbose  bool
}

// NewDefaultSoracomKryptonClient returns new SoracomClient for caller.
func NewDefaultSoracomKryptonClient(config *KryptonClientConfig) SoracomKryptonClient {
	c := DefaultSoracomKryptonClient{
		endpoint: config.Endpoint,
		client:   http.DefaultClient,
		verbose:  false,
	}

	return &c
}

// SetVerbose sets if verbose output is enabled or not.
func (c *DefaultSoracomKryptonClient) SetVerbose(v bool) {
	c.verbose = v
}

// Verbose returns if verbose output is enabled or not.
func (c *DefaultSoracomKryptonClient) Verbose() bool {
	return c.verbose
}

// Bootstrap bootstraps Arc virtual SIM.
func (c *DefaultSoracomKryptonClient) Bootstrap() (*ArcSession, error) {
	res, err := c.callAPI(&apiParams{
		method: "POST",
		path:   "/provisioning/soracom/arc/bootstrap",
		body:   "{}",
	})
	if err != nil {
		return nil, err
	}

	var arcSession ArcSession
	err = json.NewDecoder(res.Body).Decode(&arcSession)
	return &arcSession, err
}

// BootstrapWithKeyID bootstraps Arc virtual SIM with SIM authentication.
func (c *DefaultSoracomKryptonClient) BootstrapWithKeyID() (*ArcSession, error) {
	res, err := c.callAPI(&apiParams{
		method: "POST",
		path:   "/provisioning/soracom/arc/bootstrap",
		body:   "{}",
	})
	if err != nil {
		return nil, err
	}

	var config ArcSession
	err = json.NewDecoder(res.Body).Decode(&config)
	return &config, err
}

func (c *DefaultSoracomKryptonClient) callAPI(params *apiParams) (*http.Response, error) {
	req, err := c.makeRequest(params)
	if err != nil {
		return nil, err
	}

	if c.Verbose() {
		fmt.Fprintln(os.Stderr, "--- Request dump ---------------------------------")
		r, _ := httputil.DumpRequest(req, true)
		fmt.Fprintln(os.Stderr, r)
		fmt.Fprintln(os.Stderr, "--- End of request dump --------------------------")
	}
	res, err := c.doRequest(req)
	return res, err
}

func (c *DefaultSoracomKryptonClient) makeRequest(params *apiParams) (*http.Request, error) {
	var body io.Reader
	if params.body != "" {
		body = strings.NewReader(params.body)
	}

	req, err := http.NewRequest(params.method,
		fmt.Sprintf("%s/v1/%s", strings.TrimSuffix(c.endpoint, "/"), strings.TrimPrefix(params.path, "/")),
		body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Soracom-Lang", "en")
	req.Header.Set("User-Agent", internal.UserAgent)
	return req, nil
}

func (c *DefaultSoracomKryptonClient) doRequest(req *http.Request) (*http.Response, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if c.Verbose() && res != nil {
		fmt.Fprintln(os.Stderr, "--- Response dump --------------------------------")
		r, _ := httputil.DumpResponse(res, true)
		fmt.Fprintln(os.Stderr, r)
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
