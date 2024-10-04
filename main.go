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
	"golang.org/x/term"
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
		file                                                     *os.File
		audioVideo, image, archive, office, text, jq, html, diff []string
		cfg                                                      map[string][]string
		err                                                      error
	)

	file, err = os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	exitIf(err)

	audioVideo = []string{`mediainfo -- "$TVIEW_FILE"`}
	archive = []string{`atool -l -- "$TVIEW_FILE"`}
	office = []string{`libreoffice --cat "$TVIEW_FILE"`}

	image = []string{
		`chafa -f sixel -s "${TVIEW_WIDTH}x${TVIEW_HEIGHT}" -- "$TVIEW_FILE"`,
		`mediainfo -- "$TVIEW_FILE"`,
	}

	jq = []string{
		`jaq --color always . "$TVIEW_FILE"`,
		`jq -C . "$TVIEW_FILE"`,
	}

	text = []string{
		`bat --color always --paging never --terminal-width "$TVIEW_WIDTH" -- "$TVIEW_FILE"`,
		`source-highlight --failsafe -i "$TVIEW_FILE"`,
		`cat -- "$TVIEW_FILE"`,
	}

	html = []string{
		`elinks -dump 1 -no-references -no-numbering -dump-width "$TVIEW_WIDTH" "$TVIEW_FILE"`,
		`lynx -dump -nonumbers -nolist -width "$TVIEW_WIDTH" -- "$TVIEW_FILE"`,
		`w3m -dump "$TVIEW_FILE"`,
	}

	diff = []string{
		`delta < "$TVIEW_FILE"`,
		`diff-so-fancy < "$TVIEW_FILE"`,
		`colordiff < "$TVIEW_FILE"`,
	}

	cfg = map[string][]string{
		"audio/aac":             audioVideo,
		"application/x-abiword": office,
		"image/apng":            image,
		"application/x-freearc": archive,
		"image/avif":            image,
		"video/x-msvideo":       audioVideo,
		"image/bmp":             image,
		"application/x-bzip":    archive,
		"application/x-bzip2":   archive,
		"application/x-cdf":     audioVideo,
		"application/x-csh":     text,
		"text/css":              text,
		"text/csv":              text,
		"application/msword":    office,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": office,
		"application/gzip":         archive,
		"application/x-gzip":       archive,
		"image/gif":                image,
		"text/html":                html,
		"image/vnd.microsoft.icon": image,
		"application/java-archive": archive,
		"image/jpeg":               image,
		"text/javascript":          text,
		"application/json":         jq,
		"application/ld+json":      jq,
		"audio/midi":               audioVideo,
		"audio/x-midi":             audioVideo,
		"audio/mpeg":               audioVideo,
		"video/mp4":                audioVideo,
		"video/mpeg":               audioVideo,
		"application/vnd.oasis.opendocument.presentation": office,
		"application/vnd.oasis.opendocument.spreadsheet":  office,
		"application/vnd.oasis.opendocument.text":         office,
		"audio/ogg":                     audioVideo,
		"video/ogg":                     audioVideo,
		"application/ogg":               audioVideo,
		"image/png":                     image,
		"application/pdf":               office,
		"application/x-httpd-php":       text,
		"application/vnd.ms-powerpoint": office,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": office,
		"application/vnd.rar":      archive,
		"application/rtf":          text,
		"application/x-sh":         text,
		"image/svg+xml":            image,
		"application/x-tar":        archive,
		"image/tiff":               image,
		"video/mp2t":               audioVideo,
		"text/plain":               text,
		"audio/wav":                audioVideo,
		"audio/webm":               audioVideo,
		"video/webm":               audioVideo,
		"image/webp":               image,
		"application/xhtml+xml":    html,
		"application/vnd.ms-excel": office,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": office,
		"application/xml":             text,
		"application/zip":             archive,
		"video/3gpp":                  audioVideo,
		"audio/3gpp":                  audioVideo,
		"video/3gpp2":                 audioVideo,
		"audio/3gpp2":                 audioVideo,
		"application/x-7z-compressed": archive,
		"text/x-diff":                 diff,
		"text/x-patch":                diff,
		"application/x-bittorrent":    {`transmission-show -- "$TVIEW_FILE"`},
		"inode/directory":             {`ls --color --group-directories-first -w "$TVIEW_WIDTH" -- "$TVIEW_FILE"`},

		"application/octet-stream": {
			`exiftool -- "$TVIEW_FILE"`,
			`file -- "$TVIEW_FILE"`,
			`cat -- "$TVIEW_FILE"`,
		},

		"text/markdown": {
			`glow -w "$TVIEW_WIDTH" -- "$TVIEW_FILE"`,
			`mdcat --columns "$TVIEW_WIDTH" -- "$TVIEW_FILE"`,
		},
	}

	json.NewDecoder(file).Decode(&cfg)
	exitIf(file.Close())

	return cfg
}

func detectMime(path string) (string, string) {
	var (
		mime, parentMime *mimetype.MIME
		err              error
	)

	mime, err = mimetype.DetectFile(path)
	exitIf(err)

	parentMime = mime.Parent()
	if parentMime == nil {
		return strings.Split(mime.String(), ";")[0], "application/octet-stream"
	}

	return strings.Split(mime.String(), ";")[0], strings.Split(parentMime.String(), ";")[0]
}

func binaryPath(bin string) (string, bool) {
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

		return strings.Join(argv, " "), true
	}

	if errors.Is(err, exec.ErrNotFound) {
		return "", false
	}

	panic(fmt.Errorf("tview: %s", err))
}

func execProgram(binPath, path string, width, height, x, y int) bool {
	var cmd *exec.Cmd

	cmd = exec.Command("/bin/sh", "-c", "--", binPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = append(cmd.Environ(), []string{
		fmt.Sprintf("TVIEW_FILE=%s", path),
		fmt.Sprintf("TVIEW_WIDTH=%d", width),
		fmt.Sprintf("TVIEW_HEIGHT=%d", height),
		fmt.Sprintf("TVIEW_X=%d", x),
		fmt.Sprintf("TVIEW_Y=%d", y),
	}...)

	panicIf(cmd.Err)

	return cmd.Run() == nil
}

func viewFile(path string, cfg map[string][]string, width, height, x, y int) {
	var (
		mime, parentMime, bin, binPath string
		ok                             bool
	)

	mime, parentMime = detectMime(path)
	for _, bin = range append(append(cfg[mime], cfg[parentMime]...), cfg["application/octet-stream"]...) {
		binPath, ok = binaryPath(bin)
		if !ok {
			continue
		}

		if execProgram(binPath, path, width, height, x, y) {
			return
		}
	}

	exit(fmt.Errorf("%s: no valid programs", mime))
}

func main() {
	var (
		cfgFlag                             string
		widthFlag, heightFlag, xFlag, yFlag int
		width, height                       int
		argv                                []string
	)

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), `usage: tview FILE

tview displays the FILE based on mimetype.
mimetype programs list is in config file.

example: tview file.html`)

		flag.PrintDefaults()
	}

	width, height, _ = term.GetSize(int(os.Stdin.Fd()))
	flag.StringVar(&cfgFlag, "c", filepath.Join(configDir(), "config.json"), "config file path")
	flag.IntVar(&widthFlag, "w", width, "terminal width")
	flag.IntVar(&heightFlag, "h", height, "terminal height")
	flag.IntVar(&xFlag, "x", 0, "x coordinates of pane")
	flag.IntVar(&yFlag, "y", 0, "y coordinates of pane")
	flag.Parse()

	argv = flag.Args()
	if len(argv) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	viewFile(argv[0], readConfig(cfgFlag), widthFlag, heightFlag, xFlag, yFlag)
}
