package shell

import (
	"encoding/csv"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/google/shlex"
	"github.com/influxdata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/sagan/ptool/cmd"
)

const persistFlagValuesFlag = "persist-flag-values"

func executeShellLine(root *cobra.Command, line string) error {
	persist, _ := root.PersistentFlags().GetBool(persistFlagValuesFlag)
	if !persist {
		resetShellFlags(root)
	}
	args, err := shlex.Split(line)
	if err != nil {
		args = nil
	}
	if len(args) > 0 && strings.HasPrefix(args[0], "!") && len(args[0]) > 1 {
		args = append([]string{"!", args[0][1:]}, args[1:]...)
	}
	// Alias and run also read os.Args when invoking Cobra from inside a command.
	previousArgs := os.Args
	os.Args = append([]string{previousArgs[0]}, args...)
	defer func() { os.Args = previousArgs }()
	return root.Execute()
}

func resetShellFlags(command *cobra.Command) {
	command.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed && flag.Value.String() == flag.DefValue {
			return
		}
		if value, ok := flag.Value.(pflag.SliceValue); ok {
			// pflag represents slice defaults as CSV enclosed in brackets.
			defaultCSV := strings.TrimSuffix(strings.TrimPrefix(flag.DefValue, "["), "]")
			values, _ := csv.NewReader(strings.NewReader(defaultCSV)).Read()
			_ = value.Replace(values)
		} else {
			_ = flag.Value.Set(flag.DefValue)
		}
		flag.Changed = false
	})
	for _, child := range command.Commands() {
		resetShellFlags(child)
	}
}

// The command and flag completion behavior is adapted from stromland/cobra-prompt.
func shellSuggestions(root *cobra.Command, document *prompt.Document) []prompt.Suggest {
	line := document.CurrentLineBeforeCursor()
	args, err := shlex.Split(line)
	if err != nil {
		args = strings.Fields(line)
	}
	command, commandArgs, _ := root.Find(args)
	prefix := document.GetWordBeforeCursor()
	if prefix == "--" && len(commandArgs) > 0 {
		commandArgs = commandArgs[:len(commandArgs)-1]
	}
	var suggestions []prompt.Suggest
	if strings.HasPrefix(prefix, "-") {
		command.InitDefaultHelpFlag()
		if !shellFlagsTerminated(command.Flags(), commandArgs) {
			addFlag := func(flag *pflag.Flag) {
				if flag.Hidden {
					return
				}
				if strings.HasPrefix(prefix, "--") {
					suggestions = append(suggestions, prompt.Suggest{Text: "--" + flag.Name, Description: flag.Usage})
				} else if flag.Shorthand != "" {
					suggestions = append(suggestions, prompt.Suggest{Text: "-" + flag.Shorthand, Description: flag.Usage})
				}
			}
			command.LocalFlags().VisitAll(addFlag)
			command.InheritedFlags().VisitAll(addFlag)
		}
	}
	for _, subcommand := range command.Commands() {
		if !subcommand.Hidden {
			suggestions = append(suggestions, prompt.Suggest{Text: subcommand.Name(), Description: subcommand.Short})
		}
	}
	annotation := command.Annotations["cobra-prompt-dynamic-suggestions"]
	if annotation != "" {
		suggestions = append(suggestions, cmd.ShellDynamicSuggestionsFunc(annotation, document)...)
	}
	return prompt.FilterHasPrefix(suggestions, prefix, true)
}

func shellFlagsTerminated(flags *pflag.FlagSet, args []string) bool {
	if !slices.Contains(args, "--") {
		return false
	}
	parser := pflag.NewFlagSet("", pflag.ContinueOnError)
	parser.SetOutput(io.Discard)
	parser.SetNormalizeFunc(flags.GetNormalizeFunc())
	parser.ParseErrorsAllowlist.UnknownFlagsHandling = pflag.IgnoreUnknownFlag
	flags.VisitAll(func(flag *pflag.Flag) {
		flagCopy := *flag
		parser.AddFlag(&flagCopy)
	})
	// Parse the flag positions without assigning any values on the command.
	_ = parser.ParseAll(args, func(*pflag.Flag, string) error { return nil })
	return parser.ArgsLenAtDash() >= 0
}
