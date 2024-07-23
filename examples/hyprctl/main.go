// Limited reimplementation of hyprctl using hyprland-go to show an example
// on how it can be used.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/thiagokokada/hyprland-go"
)

var c *hyprland.RequestClient

// https://stackoverflow.com/a/28323276
type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(v string) error {
	*i = append(*i, v)
	return nil
}

func must1[T any](v T, err error) T {
	must(err)
	return v
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustMarshalIndent(v any) []byte {
	return must1(json.MarshalIndent(v, "", "   "))
}

func main() {
	dispatchFS := flag.NewFlagSet("dispatch", flag.ExitOnError)
	var dispatch arrayFlags
	dispatchFS.Var(&dispatch, "c", "Command to dispatch. Please quote commands with arguments (e.g.: 'exec kitty')")

	setcursorFS := flag.NewFlagSet("setcursor", flag.ExitOnError)
	theme := setcursorFS.String("theme", "Adwaita", "Cursor theme")
	size := setcursorFS.Int("size", 32, "Cursor size")

	flag.Parse()

	m := map[string]func(){
		"activewindow": func() {
			v := must1(c.ActiveWindow())
			fmt.Printf("%s\n", mustMarshalIndent(v))
		},
		"activeworkspace": func() {
			v := must1(c.ActiveWorkspace())
			fmt.Printf("%s\n", mustMarshalIndent(v))
		},
		"dispatch": func() {
			dispatchFS.Parse(os.Args[2:])
			if len(dispatch) == 0 {
				fmt.Println("-c is required for dispatch")
				os.Exit(1)
			} else {
				v := must1(c.Dispatch(dispatch...))
				fmt.Printf("%s\n", v)
			}
		},
		"kill": func() {
			v := must1(c.Kill())
			fmt.Printf("%s\n", v)
		},
		"reload": func() {
			v := must1(c.Reload())
			fmt.Printf("%s\n", v)
		},
		"setcursor": func() {
			setcursorFS.Parse(os.Args[2:])
			v := must1(c.SetCursor(*theme, *size))
			fmt.Printf("%s\n", v)
		},
		"version": func() {
			v := must1(c.Version())
			fmt.Printf("%s\n", mustMarshalIndent(v))
		},
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Printf("  %s [subcommand] <options>\n\n", os.Args[0])
		fmt.Println("Available subcommands:")
		for k := range m {
			fmt.Printf("  - %s\n", k)
		}
		os.Exit(1)
	}

	subcmd := os.Args[1]
	if run, ok := m[subcmd]; ok {
		c = hyprland.MustClient()
		run()
	} else {
		fmt.Printf("Unknown subcommand: %s\n", subcmd)
		os.Exit(1)
	}
}
