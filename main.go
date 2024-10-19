package main

import (
	"encoding/json"
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
	
	panicIf(os.MkdirAll(filepath.Join(dir, "." + dirName), 0755))
}

func readConfig(name string) map[string][]string {
	var (
		file                                                       *os.File
		audio, archive, office, video, image, jq, text, html, diff []string
		cfg                                                        map[string][]string
		err                                                        error
	)

	file, err = os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	exitIf(err)

	audio = []string{`mediainfo -- "$TVIEW_FILE"`}
	archive = []string{`atool -l -- "$TVIEW_FILE"`}

	office = []string{
		`unoconv --stdout -e PageRange=1 -f jpg -- "$TVIEW_FILE" | chafa -s "${TVIEW_WIDTH}x${TVIEW_HEIGHT}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels")`,
		`libreoffice --cat "$TVIEW_FILE"`,
	}

	video = []string{
		`ffmpegthumbnailer -i "$TVIEW_FILE" -s 0 -o /dev/stdout | chafa -s "${TVIEW_WIDTH}x${TVIEW_HEIGHT}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels")`,
		`mediainfo -- "$TVIEW_FILE"`,
	}

	image = []string{
		`chafa -s "${TVIEW_WIDTH}x${TVIEW_HEIGHT}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels") "$TVIEW_FILE"`,
		`magick "$TVIEW_FILE" jpg:- | chafa -s "${TVIEW_WIDTH}x${TVIEW_HEIGHT}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels")`,
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
		"audio/aac":             audio,
		"image/apng":            image,
		"application/x-freearc": archive,
		"image/avif":            image,
		"video/x-msvideo":       video,
		"image/bmp":             image,
		"application/x-bzip":    archive,
		"application/x-bzip2":   archive,
		"application/x-cdf":     audio,
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
		"audio/midi":               audio,
		"audio/x-midi":             audio,
		"audio/mpeg":               audio,
		"video/mp4":                video,
		"video/mpeg":               video,
		"application/vnd.oasis.opendocument.presentation": office,
		"application/vnd.oasis.opendocument.spreadsheet":  office,
		"application/vnd.oasis.opendocument.text":         office,
		"audio/ogg":                     audio,
		"video/ogg":                     video,
		"application/ogg":               audio,
		"image/png":                     image,
		"application/x-httpd-php":       text,
		"application/vnd.ms-powerpoint": office,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": office,
		"application/vnd.rar":      archive,
		"application/rtf":          text,
		"application/x-sh":         text,
		"image/svg+xml":            image,
		"application/x-tar":        archive,
		"image/tiff":               image,
		"video/mp2t":               video,
		"text/plain":               text,
		"audio/wav":                audio,
		"audio/webm":               audio,
		"video/webm":               video,
		"image/webp":               image,
		"application/xhtml+xml":    html,
		"application/vnd.ms-excel": office,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": office,
		"application/xml":             text,
		"application/zip":             archive,
		"video/3gpp":                  video,
		"audio/3gpp":                  audio,
		"video/3gpp2":                 video,
		"audio/3gpp2":                 audio,
		"application/x-7z-compressed": archive,
		"text/x-diff":                 diff,
		"text/x-patch":                diff,
		"application/pdf":             {`pdftoppm -jpeg -f 1 -singlefile -- "$TVIEW_FILE" | chafa -s "${TVIEW_WIDTH}x${TVIEW_HEIGHT}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels")`},
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

func execProgram(bin, path string, width, height, x, y int) bool {
	var cmd *exec.Cmd

	cmd = exec.Command("/bin/sh", "-c", "--", bin)
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
	var mime, parentMime, bin string

	mime, parentMime = detectMime(path)
	for _, bin = range append(append(cfg[mime], cfg[parentMime]...), cfg["application/octet-stream"]...) {
		if execProgram(bin, path, width, height, x, y) {
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
		fmt.Fprintln(flag.CommandLine.Output(), `usage: tview [OPTION]... FILE

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
