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

	prompt             *prompt.Prompt
	promptPrefix       string
	writer             prompt.ConsoleWriter
	flagValueCompleter FlagValueCompleter
}

type CompletionHint struct {
	Cmd  *cobra.Command // Currently selected command
	Flag *pflag.Flag    // Flag whose value is currently being typed
	Args []string       // Whole command line
	Prev string         // Previous word
	Curr string         // Current word under the cursor
}

type FlagValueCompleter func(hint CompletionHint) []prompt.Suggest

// New returns a new CobraPrompt based on given root command.
func New(rootCmd *cobra.Command, opts ...prompt.Option) *CobraPrompt {
	cp := &CobraPrompt{
		RootCmd:            rootCmd,
		promptPrefix:       "> ",
		writer:             &RawWriter{},
		flagValueCompleter: defaultFlagValueCompleter,
	}

	opts = append(opts, prompt.OptionCompletionOnDown()) // github.com/c-bata/go-prompt@b6bf267 or later
	opts = append(opts, prompt.OptionLivePrefix(cp.getPrefix))
	opts = append(opts, prompt.OptionShowCompletionAtStart())
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
	items := []prompt.Suggest{}
	words, prev, curr := splitDocument(d)

	// 1. Find current command
	cmd := cp.RootCmd
	found, _, err := cmd.Find(words)
	if err == nil {
		cmd = found
	}

	hint := CompletionHint{cmd, nil, words, prev, curr}

	// 2. Add subcommands to suggestions
	commandItems := cp.suggestCommand(hint)
	items = append(items, commandItems...)

	// 3. Add flag values to suggestions
	flagValueItems := cp.suggestFlagValue(hint)
	items = append(items, flagValueItems...)

	// 4. Add flags to suggestions, only if flag value suggestions are empty.
	if len(flagValueItems) == 0 {
		flagItems := cp.suggestFlag(hint)
		items = append(items, flagItems...)
	}

	return items
}

// getPrefix returns the prompt string (something like "shell> ").
func (cp *CobraPrompt) getPrefix() (prefix string, useLivePrefix bool) {
	return cp.promptPrefix, true
}

func (cp *CobraPrompt) suggestCommand(hint CompletionHint) []prompt.Suggest {
	items := []prompt.Suggest{}
	for _, c := range hint.Cmd.Commands() {
		if !c.IsAvailableCommand() {
			continue
		}
		if prefixMatches(c.Name(), hint.Curr) {
			items = append(items, prompt.Suggest{
				Text:        c.Name(),
				Description: c.Short})
		}
	}
	return items
}

func (cp *CobraPrompt) suggestFlagValue(hint CompletionHint) []prompt.Suggest {
	items := []prompt.Suggest{}

	visit := func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		name := "--" + f.Name
		shorthand := "-" + f.Shorthand
		if name == hint.Prev || (f.Shorthand != "" && shorthand == hint.Prev) {
			hint.Flag = f
			valueItems := cp.flagValueCompleter(hint)
			items = append(items, valueItems...)
		}
	}

	hint.Cmd.LocalFlags().VisitAll(visit)
	hint.Cmd.InheritedFlags().VisitAll(visit)
	return items
}

func (cp *CobraPrompt) suggestFlag(hint CompletionHint) []prompt.Suggest {
	items := []prompt.Suggest{}

	visit := func(f *pflag.Flag) {
		if f.Hidden {
			return
		}

		name := "--" + f.Name
		if prefixMatches(name, hint.Curr) {
			items = append(items, prompt.Suggest{
				Text:        name,
				Description: f.Usage})
		}

		if f.Shorthand != "" {
			shorthand := "-" + f.Shorthand
			if prefixMatches(shorthand, hint.Curr) {
				items = append(items, prompt.Suggest{
					Text:        shorthand,
					Description: f.Usage})
			}
		}
	}

	hint.Cmd.LocalFlags().VisitAll(visit)
	hint.Cmd.InheritedFlags().VisitAll(visit)
	return items
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

func splitDocument(d prompt.Document) (words []string, prev string, curr string) {
	words = splitString(d.Text)

	if d.GetWordBeforeCursor() == "" {
		// 1) cursor is at whitespace.
		//        [info --name ]
		//                    ^
		//  => prev = "--name", curr = ""
		if len(words) >= 1 {
			prev = words[len(words)-1]
		}
		curr = ""
	} else {
		// 2) cursor is in the middle of a word.
		//        [info --name abc]
		//                       ^
		//  => prev = "--name", curr = "abc"
		if len(words) >= 2 {
			prev = words[len(words)-2]
		}
		if len(words) >= 1 {
			curr = words[len(words)-1]
		}
	}
	return
}

func prefixMatches(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
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

func defaultFlagValueCompleter(hint CompletionHint) []prompt.Suggest {
	if hint.Flag.Value.Type() == "bool" {
		return nil
	} else {
		return []prompt.Suggest{
			prompt.Suggest{
				Text:        hint.Flag.DefValue,
				Description: "default value",
			},
		}
	}
}
