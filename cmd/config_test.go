package cmd

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"testing"

	"github.com/soracom/soratun"
	"github.com/stretchr/testify/assert"
)

func Test_configCmd(t *testing.T) {
	// preparing to capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()

	func() {
		os.Stdout = w
		defer func() {
			_ = w.Close()
			os.Stdout = originalStdout
		}()

		os.Args = []string{"__soratun__", "config"}
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}()

	captured, err := io.ReadAll(r)
	assert.NoError(t, err)

	var conf soratun.Config
	err = json.Unmarshal(captured, &conf)
	assert.NoError(t, err)

	localhost, err := net.LookupIP("localhost")
	assert.NoError(t, err)

	assert.EqualValues(t, localhost[0], conf.ArcSession.ArcServerEndpoint.IP)
	assert.Equal(t, 11010, conf.ArcSession.ArcServerEndpoint.Port)
	assert.EqualValues(t, "localhost:11010", conf.ArcSession.ArcServerEndpoint.RawEndpoint)
}
