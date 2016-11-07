package clinic

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/pborman/getopt"
	"gopkg.in/yaml.v2"
)

type App struct {
	Name           string
	Description    string
	Version        string
	Config         interface{}
	Action         interface{}
	Commands       []Command
	DisableHelp    bool
	DisableVersion bool

	fields     []field
	commandMap map[string]Command
	flags      *getopt.Set
	configFile map[interface{}]interface{}
}

var (
	typError       = reflect.TypeOf((*error)(nil)).Elem()
	typStringSlice = reflect.TypeOf([]string{})
)

func (a *App) tryConfig() {
	path := os.ExpandEnv(fmt.Sprintf("$HOME/.config/%s/config.yml", strings.ToLower(a.Name)))
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, &a.configFile)
	if err != nil {
		return
	}

	for _, f := range a.fields {
		if cval, ok := a.configFile[f.long]; ok {
			f.set(cval)
		}
	}
}

func (a *App) usage() {
	usage := fmt.Sprintf("Usage: %s", a.Name)

	if a.Config != nil {
		usage += " [OPTIONS]"
	}

	hasCommands := false
	for _, cmd := range a.Commands {
		if !cmd.Hidden {
			hasCommands = true
			break
		}
	}

	if hasCommands {
		if a.Action == nil {
			usage += " COMMAND"
		} else {
			usage += " [COMMAND]"
		}
	}

	if acceptsArgs(a.Action) {
		usage += " [args...]"
	}

	if !a.DisableHelp {
		usage += fmt.Sprintf("\n       %s [-h | --help]", a.Name)
	}

	if !a.DisableVersion {
		usage += fmt.Sprintf("\n       %s [-v | --version]", a.Name)
	}

	usage += "\n\n"

	if a.Description != "" {
		usage += a.Description + "\n\n"
	}

	fmt.Fprintf(os.Stderr, usage)

	if a.Config != nil {
		fmt.Fprintf(os.Stderr, "Options:\n\n")
		tw := tabwriter.NewWriter(os.Stderr, 0, 1, 4, ' ', 0)
		options := getOptions(a.fields)
		for _, opt := range options {
			fmt.Fprintln(tw, "  "+strings.Join(opt, "\t"))
		}
		fmt.Fprintln(tw)
		tw.Flush()
	}

	if hasCommands {
		fmt.Fprintf(os.Stderr, "Commands:\n\n")
		tw := tabwriter.NewWriter(os.Stderr, 0, 1, 4, ' ', 0)
		for _, cmd := range a.Commands {
			if !cmd.Hidden {
				fmt.Fprintf(tw, "  %s\t%s", cmd.Name, cmd.Description)
			}
		}
		fmt.Fprintln(tw)
		tw.Flush()
	}
}

func (a *App) Run(args []string) {
	if a.Name == "" {
		a.Name = filepath.Base(os.Args[0])
	}

	if a.Version == "" {
		a.DisableVersion = true
	}

	if len(a.Commands) > 0 {
		a.commandMap = make(map[string]Command)
		for _, cmd := range a.Commands {
			if cmd.Name == "" {
				a.Fatal("Command names cannot be empty")
			}

			if cmd.Description == "" {
				a.Fatal("Command descriptions cannot be empty")
			}

			a.commandMap[cmd.Name] = cmd
		}
	}

	a.flags = getopt.New()
	a.flags.SetProgram(a.Name)
	a.flags.SetUsage(a.usage)
	var help bool
	var version bool

	if !a.DisableHelp {
		a.flags.BoolVarLong(&help, "help", 'h', "print usage information")
	}

	if !a.DisableVersion {
		a.flags.BoolVarLong(&version, "version", 'v', "print version information")
	}

	if a.Config != nil {
		if reflect.TypeOf(a.Config).Kind() != reflect.Ptr {
			a.Fatal("Config must be a struct pointer")
		}
	}

	a.fields = parseFields(a.Config)
	addToFlags(a.fields, a.flags)
	a.tryConfig()
	a.flags.Parse(os.Args)

	if version {
		a.Info(a.Version)
		os.Exit(0)
	}

	if a.flags.NArgs() > 0 {
		if cmd, ok := a.commandMap[a.flags.Arg(0)]; ok {
			cmd.app = a
			cmd.Run(a.flags.Args())
			return
		}
	}

	if a.Action == nil || help {
		a.usage()
		os.Exit(0)
	}

	promptForMissing(a.fields, a.configFile, a.flags)

	actionTyp := reflect.TypeOf(a.Action)
	if actionTyp.Kind() != reflect.Func {
		a.Fatal("Action must be a function")
	}

	if actionTyp.NumOut() != 1 || !actionTyp.Out(0).Implements(typError) {
		a.Fatal("Action must return an error")
	}

	var actionArgs []reflect.Value
	for i := 0; i < actionTyp.NumIn(); i++ {
		inTyp := actionTyp.In(i)
		if inTyp == typStringSlice {
			actionArgs = append(actionArgs, reflect.ValueOf(a.flags.Args()))
		} else if inTyp == reflect.TypeOf(a.Config) {
			actionArgs = append(actionArgs, reflect.ValueOf(a.Config))
		} else if inTyp == reflect.TypeOf(a) {
			actionArgs = append(actionArgs, reflect.ValueOf(a))
		} else {
			if a.Config == nil {
				a.Fatal(fmt.Sprintf("Action can only accept parameters of type %s or %s", typStringSlice, reflect.TypeOf(a)))
			}

			a.Fatal(fmt.Sprintf("Action can only accept parameters of type %s, %s or %s", typStringSlice, reflect.TypeOf(a), reflect.TypeOf(a.Config)))
		}
	}

	action := reflect.ValueOf(a.Action)
	ret := action.Call(actionArgs)
	if ret[0].Elem().IsValid() {
		a.Fatal(ret[0].Elem().Interface().(error).Error())
	}
}
