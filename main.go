package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Kai-Animator/kanji-boxes/model"
	"github.com/Kai-Animator/kanji-boxes/storage"
)

// make build 時に -ldflags で上書きされる
var version = "dev"

func main() {
	name := filepath.Base(os.Args[0])
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	showVersion := flags.Bool("version", false, "print version and exit")
	flags.Usage = func() { printUsage(flags.Output(), name) }
	_ = flags.Parse(os.Args[1:])

	if *showVersion {
		fmt.Printf("%s %s\n", name, resolveVersion())
		return
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected argument: %s\n\n", flags.Arg(0))
		printUsage(os.Stderr, name)
		os.Exit(2)
	}

	program := tea.NewProgram(model.New())
	if _, err := program.Run(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}
}

// go install で入れた場合は ldflags が無いため、モジュールのバージョン情報を使う
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func printUsage(w io.Writer, name string) {
	dataDir, err := storage.ResolveDataDir("")
	if err != nil {
		dataDir = "unavailable (" + err.Error() + ")"
	}
	fmt.Fprintf(w, `%[1]s - terminal kanji trainer using the Leitner box system

Usage:
  %[1]s              start the app
  %[1]s --help       show this help
  %[1]s --version    print version

Environment:
  %[2]s    data directory (current: %[3]s)
  OPENAI_API_KEY     enables AI autofill on the Add Card screen

Key hints are shown at the bottom of every screen.
More: https://github.com/Kai-Animator/kanji-boxes
`, name, storage.EnvDataDir, dataDir)
}
