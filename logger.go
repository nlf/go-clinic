package clinic

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

var (
	stdoutColor = isatty.IsTerminal(os.Stdout.Fd())
	stderrColor = isatty.IsTerminal(os.Stderr.Fd())

	blue  = "\033[34m"
	red   = "\033[31m"
	reset = "\033[0m"
)

func (a *App) Info(args ...interface{}) {
	var prefix string
	if stdoutColor {
		prefix = fmt.Sprintf("%s[-]%s ", blue, reset)
	} else {
		prefix = "[-] "
	}

	fmt.Printf(prefix)
	fmt.Println(args...)
}

func (a *App) Error(args ...interface{}) {
	var prefix string
	if stderrColor {
		prefix = fmt.Sprintf("%s[!]%s ", red, reset)
	} else {
		prefix = "[!] "
	}

	fmt.Fprintf(os.Stderr, prefix)
	fmt.Fprintln(os.Stderr, args...)
}

func (a *App) Fatal(args ...interface{}) {
	a.Error(args...)
	os.Exit(1)
}
