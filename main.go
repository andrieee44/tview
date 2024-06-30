package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func exit(err error) {
	fmt.Fprintln(os.Stderr, "tview:", err)
	os.Exit(1)
}

func exitIf(err error) {
	if err != nil {
		exit(err)
	}
}

func panicIf(err error) {
	if err != nil {
		panic(fmt.Errorf("tview: %s", err))
	}
}

func configDir() string {
	const dirName string = "tview"

	var dir string

	dir = os.Getenv("XDG_CONFIG_HOME")
	if dir != "" {
		dir = filepath.Join(dir, dirName)
		panicIf(os.MkdirAll(dir, 0755))

		return dir
	}

	dir = os.Getenv("HOME")
	if dir != "" {
		dir = filepath.Join(dir, ".config", dirName)
		panicIf(os.MkdirAll(dir, 0755))

		return dir
	}

	panic(errors.New("tview: $HOME is empty"))
}

func readConfig(name string) map[string][]string {
	var (
		file *os.File
		cfg  map[string][]string
		err  error
	)

	file, err = os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	exitIf(err)

	cfg = map[string][]string{
		"text/plain": {"cat"},
	}

	json.NewDecoder(file).Decode(&cfg)
	exitIf(file.Close())

	return cfg
}

func getFile(args []string) *os.File {
	var (
		file *os.File
		err  error
	)

	if len(args) == 0 {
		return os.Stdin
	}

	file, err = os.Open(args[0])
	exitIf(err)

	return file
}

func detectMime(file *os.File) string {
	return "text/plain"
}

func programPath(prog string) (string, bool) {
	var (
		path string
		err  error
	)

	path, err = exec.LookPath(prog)
	if err == nil {
		return path, true
	}

	if errors.Is(err, exec.ErrNotFound) {
		return "", false
	}

	panic(fmt.Errorf("tview: %s", err))
}

func execProgram(prog string, file *os.File) {
	var cmd *exec.Cmd

	cmd = exec.Command("/bin/sh", "-c", "--", prog)
	cmd.Stdin = file
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	panicIf(cmd.Err)
	exitIf(cmd.Run())
}

func viewFile(file *os.File, cfg map[string][]string) {
	var (
		mime, bin, path string
		bins            []string
		ok              bool
	)

	mime = detectMime(file)
	bins, ok = cfg[mime]
	if !ok {
		exit(fmt.Errorf("%s: empty program list", mime))
	}

	for _, bin = range bins {
		path, ok = programPath(bin)
		if ok {
			execProgram(path, file)
			panicIf(file.Close())

			return
		}
	}

	exit(fmt.Errorf("%s: no valid programs", mime))
}

func main() {
	var cfgFlag string

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), `usage: tview FILE

tview displays the FILE based on mimetype.
mimetype programs list is in config file.
if no FILE, read from STDIN.

example: tview file.html

program receives the FILE through STDIN.`)

		flag.PrintDefaults()
	}

	flag.StringVar(&cfgFlag, "c", filepath.Join(configDir(), "config.json"), "config file path")
	flag.Parse()

	viewFile(getFile(flag.Args()), readConfig(cfgFlag))
}
