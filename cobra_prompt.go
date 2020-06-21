package cobraprompt

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/google/shlex"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CobraPrompt struct {
	RootCmd *cobra.Command

	prompt       *prompt.Prompt
	promptPrefix string
	writer       prompt.ConsoleWriter
}

// New returns a new CobraPrompt based on given root command.
func New(rootCmd *cobra.Command, opts ...prompt.Option) *CobraPrompt {
	cp := &CobraPrompt{
		RootCmd:      rootCmd,
		promptPrefix: "> ",
		writer:       &RawWriter{},
	}

	opts = append(opts, prompt.OptionLivePrefix(cp.getPrefix))
	opts = append(opts, prompt.OptionWriter(cp.writer))

	p := prompt.New(
		func(in string) {
			cp.executor(in)
		},
		func(d prompt.Document) []prompt.Suggest {
			return cp.completer(d)
		},
		opts...,
	)
	cp.prompt = p

	return cp
}

// Run starts an interactive shell.
func (cp *CobraPrompt) Run() {
	cp.prompt.Run()
}

// SetPromptPrefix sets the command prompt string (a.k.a. prefix).
// This function can be called any time.
func (cp *CobraPrompt) SetPromptPrefix(prefix string) {
	cp.promptPrefix = prefix
}

// executor executes a command from given input line.
func (cp *CobraPrompt) executor(in string) {
	words := splitString(in)

	resetFlagValues(cp.RootCmd)

	cp.RootCmd.SetArgs(words)
	err := cp.RootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

// completer returns a list of suggestions that can fit in the current cursor
// location.
func (cp *CobraPrompt) completer(d prompt.Document) []prompt.Suggest {
	return nil
}

// getPrefix returns the prompt string (something like "shell> ").
func (cp *CobraPrompt) getPrefix() (prefix string, useLivePrefix bool) {
	return cp.promptPrefix, true
}

func splitString(in string) []string {
	// Split by whitespaces, but respects single- and double-quotes
	// e.g. 'info --name "John Doe"' -> "info", "--name", "John Doe".
	words, err := shlex.Split(in)
	// Falls back to whitespace split when error occurs (e.g. unmatched quotes)
	if err != nil {
		words = strings.Fields(in)
	}
	return words
}

// resetFlagValues resets values of all flags in a command to their default
// values. It is recommended to call this function before executing any command.
func resetFlagValues(c *cobra.Command) {
	c.Flags().VisitAll(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue)
	})
	for _, subcommand := range c.Commands() {
		resetFlagValues(subcommand)
	}
}
