package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		"text/plain":               {"file --", "cat --"},
		"application/octet-stream": {"file --", "cat --"},
	}

	json.NewDecoder(file).Decode(&cfg)
	exitIf(file.Close())

	return cfg
}

func detectMime(path string) string {
	var (
		mime *mimetype.MIME
		err  error
	)

	mime, err = mimetype.DetectFile(path)
	exitIf(err)

	return strings.Split(mime.String(), ";")[0]
}

func binaryPath(bin, path string) (string, bool) {
	var (
		argv    []string
		binPath string
		err     error
	)

	argv = strings.Fields(bin)
	if len(argv) == 0 {
		exit(errors.New("unexpected empty binary path"))
	}

	binPath, err = exec.LookPath(argv[0])
	if err == nil {
		argv[0] = binPath
		argv = append(argv, path)

		return strings.Join(argv, " "), true
	}

	if errors.Is(err, exec.ErrNotFound) {
		return "", false
	}

	panic(fmt.Errorf("tview: %s", err))
}

func execProgram(binPath string) {
	var cmd *exec.Cmd

	cmd = exec.Command("/bin/sh", "-c", "--", binPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	panicIf(cmd.Err)
	exitIf(cmd.Run())
}

func viewFile(path string, cfg map[string][]string) {
	var (
		mime, bin, binPath string
		bins               []string
		ok                 bool
	)

	mime = detectMime(path)

	bins, ok = cfg[mime]
	if !ok {
		bins = cfg["application/octet-stream"]
	}

	for _, bin = range bins {
		binPath, ok = binaryPath(bin, path)
		if !ok {
			continue
		}

		execProgram(binPath)

		return
	}

	exit(fmt.Errorf("%s: no valid programs", mime))
}

func main() {
	var (
		cfgFlag string
		argv    []string
	)

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), `usage: tview FILE

tview displays the FILE based on mimetype.
mimetype programs list is in config file.

example: tview file.html`)

		flag.PrintDefaults()
	}

	flag.StringVar(&cfgFlag, "c", filepath.Join(configDir(), "config.json"), "config file path")
	flag.Parse()

	argv = flag.Args()
	if len(argv) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	viewFile(argv[0], readConfig(cfgFlag))
}
