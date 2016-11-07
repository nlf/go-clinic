### clinic

this package is a golang cli framework that aims to be as simple to consume as possible.

it contains the things that i personally find useful, and that's about it. if you have an idea for something to add feel free to open an issue.

### features

clinic automatically generates flags based on structs. in addition it will attempt to populate those structs based on a yaml formatted config file located at `$HOME/.config/APPNAME/config.yml`.

if a struct field is tagged `prompt:"yes"` and that field is not set in the config file, and not passed as a flag, clinic will prompt the user for input at runtime.

### usage

first, the smallest viable app

```golang
package main

import (
  "fmt"
  "os"

  "github.com/nlf/go-clinic"
)

func main() {
  app := clinic.App{
    Action: func() error {
      fmt.Println("hello, world")
      return nil
    },
  }

  app.Run(os.Args)
}
```

want to pass some flags?

```golang
package main

import (
  "fmt"
  "os"

  "github.com/nlf/go-clinic"
)

type config struct {
  Name string
}

func main() {
  app := clinic.App{
    Config: &config{
      Name: "world",
    },
    Action: func(cfg *config) error {
      fmt.Printf("hello, %s\n", cfg.Name)
      return nil
    },
  }

  app.Run(os.Args)
}
```

clinic will use reflection to automatically create flags for you. by default the flag will be the struct field's name, lowercase (in this example `--name`).
defaults are defined by setting an initial value to the `Config` field. in the example above if the app is run with `app` the result will be `hello, world`.
however, if it's run `app --name r2d2` the result will be `hello, r2d2`. the config struct is injected into the action function automatically.

want some prettier output?

```golang
    Action: func(app *clinic.App, cfg *config) error {
      app.Info(fmt.Sprintf("hello, %s\n", cfg.Name))
      return nil
    },
```

`Info`, `Error` and `Fatal` methods are available on the `App` object, they're very simple methods to put a prefix on your output to draw some attention to them.

also available on the `App` is a `Spin` method:

```golang
err := app.Spin("Doing some work", func() error {
  return doSomeLongRunningThing()
})
```

this will print the string passed as the first parameter with a spinner that will run while the function passed as the second parameter runs. once the function
completes, if it returns `nil` the spinner will be replaced with the text `OK`, if it returns an `error` the spinner will be replaced with `FAIL`. in either
case the return value of the function will be passed as the result of the `Spin` function.

clinic also supports commands

```golang
package main

import (
  "fmt"
  "os"

  "github.com/nlf/go-clinic"
)

type config struct {
  Name string
}

type cookConfig struct {
  Food string
}

func main() {
  app := clinic.App{
    Config: &config{
      Name: "world",
    },
    Action: func(cfg *config) error {
      fmt.Printf("hello, %s\n", cfg.Name)
      return nil
    },
    Commands: []clinic.Command{
      {
        Name: "cook",
        Description: "cook something",
        Config: &cookConfig{
          Food: "bacon",
        },
        Action: func(cfg *config, cookCfg *cookConfig) error {
          fmt.Printf("cooking some %s for %s\n", cookCfg.Food, cfg.Name)
          return nil
        },
      },
    },
  }

  app.Run(os.Args)
}
```

the root config and the command's config are both available to be injected into a command's action, as well as the `App` object itself.

clinic generates help output for you, the output of `app --help` for the above example looks like the following

```
Usage: app [OPTIONS] [COMMAND]
       app [-h | --help]

Options:

  --name ["world"]    Name

Commands:

  cook    cook something
```

as you can see, the `--name` flag is present and reflects the default value that was set. the `Name` on the right is the usage which can be overridden with a struct tag.
for example

```golang
type config struct {
  Name string `usage:"name of the person to greet"`
}
```

would change the output to

```
Usage: app [OPTIONS] [COMMAND]
       app [-h | --help]

Options:

  --name ["world"]    name of the person to greet

Commands:

  cook    cook something
```

at the bottom, commands and their descriptions are also listed. likewise help for a command is available by running `app --help cook` giving this output

```
Usage: app [OPTIONS] cook

cook something

Options:

  --name ["world"]      name of the person to greet
  --food ["bacon"]      Food
```

options for both the root config, as well as the command are displayed and the command list is removed.

if a `Version` field is provided to the `App` object a `--version` flag is also available that simply prints the contents of `Version` and exits.
