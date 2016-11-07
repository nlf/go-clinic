package clinic

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
)

type Command struct {
	Name        string
	Description string
	Config      interface{}
	Action      interface{}
	Hidden      bool

	app    *App
	fields []field
}

func (c *Command) tryConfig() {
	if rawSegment, ok := c.app.configFile[strings.ToLower(c.Name)]; ok {
		if segment, ok := rawSegment.(map[interface{}]interface{}); ok {
			for _, f := range c.fields {
				if val, ok := segment[f.long]; ok {
					f.set(val)
				}
			}
		}
	}
}

func (c *Command) usage() {
	usage := fmt.Sprintf("Usage: %s", c.app.Name)

	if c.app.Config != nil || c.Config != nil {
		usage += " [OPTIONS]"
	}

	usage += " " + c.Name

	if acceptsArgs(c.Action) {
		usage += " [args...]"
	}

	usage += "\n\n" + c.Description + "\n\n"

	fmt.Fprintf(os.Stderr, usage)

	if c.app.Config != nil || c.Config != nil {
		fmt.Fprintf(os.Stderr, "Options:\n\n")
		tw := tabwriter.NewWriter(os.Stderr, 0, 4, 0, '\t', 0)
		if c.app.Config != nil {
			options := getOptions(c.app.fields)
			for _, opt := range options {
				fmt.Fprintln(tw, "  "+strings.Join(opt, "\t"))
			}
		}

		if c.Config != nil {
			options := getOptions(c.fields)
			for _, opt := range options {
				fmt.Fprintln(tw, "  "+strings.Join(opt, "\t"))
			}
		}

		fmt.Fprintln(tw)
		tw.Flush()
	}
}

func (c *Command) Run(args []string) {
	if c.Config != nil {
		if reflect.TypeOf(c.Config).Kind() != reflect.Ptr {
			c.app.Fatal("Config must be a struct pointer")
		}
	}

	c.fields = parseFields(c.Config)
	addToFlags(c.fields, c.app.flags)
	c.tryConfig()
	c.app.flags.Parse(c.app.flags.Args())

	if c.app.flags.Lookup("help").Value().String() == "true" {
		c.usage()
		os.Exit(0)
	}

	promptForMissing(c.app.fields, c.app.configFile, c.app.flags)
	var segment map[interface{}]interface{}
	if _, ok := c.app.configFile[strings.ToLower(c.Name)]; ok {
		segment = c.app.configFile[strings.ToLower(c.Name)].(map[interface{}]interface{})
	}
	promptForMissing(c.fields, segment, c.app.flags)

	actionTyp := reflect.TypeOf(c.Action)
	if actionTyp.Kind() != reflect.Func {
		c.app.Fatal("Action must be a function")
	}

	if actionTyp.NumOut() != 1 || !actionTyp.Out(0).Implements(typError) {
		c.app.Fatal("Action must return an error")
	}

	var actionArgs []reflect.Value
	for i := 0; i < actionTyp.NumIn(); i++ {
		inTyp := actionTyp.In(i)
		if inTyp == typStringSlice {
			actionArgs = append(actionArgs, reflect.ValueOf(c.app.flags.Args()))
		} else if inTyp == reflect.TypeOf(c.Config) {
			actionArgs = append(actionArgs, reflect.ValueOf(c.Config))
		} else if inTyp == reflect.TypeOf(c.app.Config) {
			actionArgs = append(actionArgs, reflect.ValueOf(c.app.Config))
		} else if inTyp == reflect.TypeOf(c.app) {
			actionArgs = append(actionArgs, reflect.ValueOf(c.app))
		} else {
			msg := fmt.Sprintf("Action can only accept parameters of type %s", typStringSlice)
			if c.Config != nil {
				if c.app.Config != nil {
					msg += fmt.Sprintf(", %s, %s or %s", reflect.TypeOf(c.app), reflect.TypeOf(c.Config), reflect.TypeOf(c.app.Config))
				} else {
					msg += fmt.Sprintf(", %s or %s", reflect.TypeOf(c.app), reflect.TypeOf(c.Config))
				}
			} else if c.app.Config != nil {
				msg += fmt.Sprintf(", %s or %s", reflect.TypeOf(c.app), reflect.TypeOf(c.app.Config))
			} else {
				msg += fmt.Sprintf(" or %s", reflect.TypeOf(c.app))
			}

			c.app.Fatal(msg)
		}
	}

	action := reflect.ValueOf(c.Action)
	ret := action.Call(actionArgs)
	if ret[0].Elem().IsValid() {
		c.app.Fatal(ret[0].Elem().Interface().(error).Error())
	}
}
