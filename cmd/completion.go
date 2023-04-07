package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func completionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "completion [bash|zsh]",
		Aliases: []string{"v"},
		Hidden:  true,
		Short:   "Generate the autocompletion script for the specified shell",
		Long: `Generate the autocompletion script for soratun for the specified shell. You will need to start a new shell for this setup to take effect.

# Bash
This script depends on the "bash-completion" package. If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:
    $ source <(soratun completion bash)

To load completions for every new session, execute once:
Linux:
    $ soratun completion bash > /etc/bash_completion.d/soratun
macOS:
    $ soratun completion bash > /usr/local/etc/bash_completion.d/soratun

# Zsh
If shell completion is not already enabled in your environment you will need to enable it. You can execute the following once:
    $ echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for every new session, execute once:
Linux:
    $ soratun completion zsh > "${fpath[1]}/_soratun"
macOS:
    $ soratun completion zsh > /usr/local/share/zsh/site-functions/_soratun
`,
		ValidArgs: []string{"bash", "zsh"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			switch args[0] {
			case "bash":
				err = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				err = cmd.Root().GenZshCompletion(os.Stdout)
			}
			if err != nil {
				log.Fatalf("Error while creating a completion script: %s", err)
			}
		},
	}
}
