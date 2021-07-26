package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/soracom/soratun/internal"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	versionInfo, _ := json.Marshal(map[string]string{
		"version":  internal.Version,
		"revision": internal.Revision,
	})

	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Show version",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", versionInfo)
		},
	}
}
