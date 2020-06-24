package main

import (
	"os"

	cobraprompt "github.com/blukat29/cobra-prompt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

var items = []string{}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.AddCommand(newQuitCommand())
	cmd.AddCommand(newShowCommand())
	cmd.AddCommand(newAddCommand())
	return cmd
}

func newQuitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quit",
		Short: "Quit program",
		RunE:  quitFunc,
	}
	return cmd
}

func quitFunc(cmd *cobra.Command, args []string) error {
	os.Exit(0)
	return nil
}

func completer(hint cobraprompt.CompletionHint) []prompt.Suggest {
	if hint.Flag.Name == "color" {
		return []prompt.Suggest{
			prompt.Suggest{Text: "green", Description: "young apple"},
			prompt.Suggest{Text: "red", Description: "ripen apple"},
		}
	} else {
		return nil
	}
}

func main() {
	rootCmd := newRootCommand()
	cp := cobraprompt.New(rootCmd)

	cp.SetPromptPrefix(">> ")
	cp.SetFlagValueCompleter(completer)
	cp.Run()
}
