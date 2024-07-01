package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
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
		"text/plain":               {"cat"},
		"application/octet-stream": {"file", "cat"},
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

func detectMime(file *os.File) (io.Reader, string) {
	var (
		header *bytes.Buffer
		mime   *mimetype.MIME
		err    error
	)

	header = bytes.NewBuffer(nil)

	mime, err = mimetype.DetectReader(io.TeeReader(file, header))
	exitIf(err)

	return io.MultiReader(header, file), mime.String()
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

func execProgram(prog string, stdin io.Reader) {
	var cmd *exec.Cmd

	cmd = exec.Command("/bin/sh", "-c", "--", prog)
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	panicIf(cmd.Err)
	exitIf(cmd.Run())
}

func viewFile(file *os.File, cfg map[string][]string) {
	var (
		mime, bin, path string
		stdin           io.Reader
		bins            []string
		ok              bool
	)

	stdin, mime = detectMime(file)

	bins, ok = cfg[mime]
	if !ok {
		bins = cfg["application/octet-stream"]
	}

	for _, bin = range bins {
		path, ok = programPath(bin)
		if ok {
			execProgram(path, stdin)
			panicIf(file.Close())

			return
		}

		fmt.Fprintf(os.Stderr, "tview: %s: cannot locate program\n", bin)
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
