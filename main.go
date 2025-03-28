package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/cespare/xxhash/v2"
	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/term"
)

type flagsStruct struct {
	config, cache string
	columns, rows int
}

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

func exists(path string) bool {
	var err error

	_, err = os.Stat(path)
	if err == nil {
		return true
	}

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	exit(err)

	return false
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
		`unoconv --stdout -e PageRange=1 -f jpg -- "$TVIEW_FILE" | chafa -s "${TVIEW_COLUMNS}x${TVIEW_ROWS}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels")`,
		`libreoffice --cat "$TVIEW_FILE"`,
	}

	video = []string{
		`ffmpegthumbnailer -i "$TVIEW_FILE" -s 0 -o /dev/stdout | chafa -s "${TVIEW_COLUMNS}x${TVIEW_ROWS}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels")`,
		`mediainfo -- "$TVIEW_FILE"`,
	}

	image = []string{
		`chafa -s "${TVIEW_COLUMNS}x${TVIEW_ROWS}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels") "$TVIEW_FILE"`,
		`magick "$TVIEW_FILE" jpg:- | chafa -s "${TVIEW_COLUMNS}x${TVIEW_ROWS}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels")`,
		`mediainfo -- "$TVIEW_FILE"`,
	}

	jq = []string{
		`jaq -C always . "$TVIEW_FILE"`,
		`jq -C . "$TVIEW_FILE"`,
	}

	text = []string{
		`bat --color always --paging never --terminal-width "$NS" -- "$TVIEW_FILE"`,
		`source-highlight --failsafe -i "$TVIEW_FILE"`,
		`cat -- "$TVIEW_FILE"`,
	}

	html = []string{
		`elinks -dump 1 -no-references -no-numbering -dump-width "$TVIEW_COLUMNS" "$TVIEW_FILE"`,
		`lynx -dump -nonumbers -nolist -width "$TVIEW_COLUMNS" -- "$TVIEW_FILE"`,
		`w3m -dump "$TVIEW_FILE"`,
	}

	diff = []string{
		`delta < "$TVIEW_FILE"`,
		`diff-so-fancy < "$TVIEW_FILE"`,
		`colordiff < "$TVIEW_FILE"`,
	}

	cfg = map[string][]string{
		"application/gzip":                                archive,
		"application/java-archive":                        archive,
		"application/json":                                jq,
		"application/ld+json":                             jq,
		"application/msword":                              office,
		"application/ogg":                                 audio,
		"application/pdf":                                 {`pdftoppm -jpeg -f 1 -singlefile -- "$TVIEW_FILE" | chafa -s "${TVIEW_COLUMNS}x${TVIEW_ROWS}" $([ "${XDG_SESSION_TYPE:-}" = "tty" ] || printf -- "-f sixels")`},
		"application/rtf":                                 text,
		"application/vnd.ms-excel":                        office,
		"application/vnd.ms-powerpoint":                   office,
		"application/vnd.oasis.opendocument.presentation": office,
		"application/vnd.oasis.opendocument.spreadsheet":  office,
		"application/vnd.oasis.opendocument.text":         office,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": office,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         office,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   office,
		"application/vnd.rar":         archive,
		"application/x-7z-compressed": archive,
		"application/x-bittorrent":    {`transmission-show -- "$TVIEW_FILE"`},
		"application/x-bzip":          archive,
		"application/x-bzip2":         archive,
		"application/x-cdf":           audio,
		"application/x-csh":           text,
		"application/x-freearc":       archive,
		"application/x-gzip":          archive,
		"application/x-httpd-php":     text,
		"application/x-sh":            text,
		"application/x-tar":           archive,
		"application/xhtml+xml":       html,
		"application/xml":             text,
		"application/zip":             archive,
		"audio/3gpp":                  audio,
		"audio/3gpp2":                 audio,
		"audio/aac":                   audio,
		"audio/midi":                  audio,
		"audio/mpeg":                  audio,
		"audio/ogg":                   audio,
		"audio/wav":                   audio,
		"audio/webm":                  audio,
		"audio/x-midi":                audio,
		"image/apng":                  image,
		"image/avif":                  image,
		"image/bmp":                   image,
		"image/gif":                   image,
		"image/jpeg":                  image,
		"image/png":                   image,
		"image/svg+xml":               image,
		"image/tiff":                  image,
		"image/vnd.microsoft.icon":    image,
		"image/webp":                  image,
		"inode/directory":             {`ls --color --group-directories-first -w "$TVIEW_COLUMNS" -- "$TVIEW_FILE"`},
		"text/css":                    text,
		"text/csv":                    text,
		"text/html":                   html,
		"text/javascript":             text,
		"text/plain":                  text,
		"text/x-diff":                 diff,
		"text/x-patch":                diff,
		"video/3gpp":                  video,
		"video/3gpp2":                 video,
		"video/mp2t":                  video,
		"video/mp4":                   video,
		"video/mpeg":                  video,
		"video/ogg":                   video,
		"video/webm":                  video,
		"video/x-matroska":            video,
		"video/x-msvideo":             video,

		"application/octet-stream": {
			`exiftool -- "$TVIEW_FILE"`,
			`file -- "$TVIEW_FILE"`,
			`cat -- "$TVIEW_FILE"`,
		},

		"text/markdown": {
			`glow -w "$TVIEW_COLUMNS" -- "$TVIEW_FILE"`,
			`mdcat --columns "$TVIEW_COLUMNS" -- "$TVIEW_FILE"`,
		},
	}

	json.NewDecoder(file).Decode(&cfg)
	exitIf(file.Close())

	return cfg
}

func detectMime(data []byte) (string, string) {
	var (
		mime, parentMime *mimetype.MIME
		err              error
	)

	mime = mimetype.Detect(data)
	exitIf(err)

	parentMime = mime.Parent()
	if parentMime == nil {
		return strings.Split(mime.String(), ";")[0], "application/octet-stream"
	}

	return strings.Split(mime.String(), ";")[0], strings.Split(parentMime.String(), ";")[0]
}

func execProgram(bin string, flags *flagsStruct, path string, cache *os.File) bool {
	var (
		cmd     *exec.Cmd
		success bool
		err     error
	)

	cmd = exec.Command("/bin/sh", "-c", "--", bin)
	cmd.Stdout = cache
	cmd.Stderr = os.Stderr

	cmd.Env = append(cmd.Environ(), []string{
		fmt.Sprintf("TVIEW_FILE=%s", path),
		fmt.Sprintf("TVIEW_COLUMNS=%d", flags.columns),
		fmt.Sprintf("TVIEW_ROWS=%d", flags.rows),
	}...)

	panicIf(cmd.Err)
	success = cmd.Run() == nil

	_, err = cache.Seek(0, io.SeekStart)
	panicIf(err)

	return success
}

func viewFile(flags *flagsStruct, path string) {
	var (
		data                             []byte
		mime, parentMime, bin, cachePath string
		cfg                              map[string][]string
		cacheHit                         bool
		cache                            *os.File
		err                              error
	)

	data, err = os.ReadFile(path)
	exitIf(err)

	mime, parentMime = detectMime(data)
	cfg = readConfig(flags.config)
	cachePath = filepath.Join(flags.cache, fmt.Sprintf("%x", xxhash.Sum64(append(data, fmt.Sprintf("%d%d", flags.columns, flags.rows)...))))
	cacheHit = exists(cachePath)

	cache, err = os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE, 0600)
	panicIf(err)

	for _, bin = range append(append(cfg[mime], cfg[parentMime]...), cfg["application/octet-stream"]...) {
		if cacheHit || execProgram(bin, flags, path, cache) {
			_, err = io.Copy(os.Stdout, cache)
			panicIf(err)

			exitIf(cache.Close())

			return
		}
	}

	exit(fmt.Errorf("%s: no valid programs", mime))
}

func main() {
	var (
		configPath    string
		flags         *flagsStruct
		columns, rows int
		argv          []string
		err           error
	)

	configPath, err = xdg.ConfigFile("tview/config.json")
	panicIf(err)

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: tview [--cache <directory>] [--columns <columns>] [--config <path>] [--rows <rows>] <FILE>")
		flag.PrintDefaults()
	}

	columns, rows, _ = term.GetSize(int(os.Stdin.Fd()))
	flags = &flagsStruct{}

	flag.StringVar(&flags.cache, "cache", filepath.Join(xdg.CacheHome, "tview"), `The directory used to store the cached file previews. Defaults to "${XDG_CACHE_HOME}/tview".`)
	flag.IntVar(&flags.columns, "columns", columns, `The amount of terminal columns passed to the external programs with the environment variable "$TVIEW_COLUMNS". Defaults to the column count of the terminal tview is running in.`)
	flag.StringVar(&flags.config, "config", configPath, `The path to configuration file. Defaults to "${XDG_CONFIG_HOME}/tview/config.json".`)
	flag.IntVar(&flags.rows, "rows", rows, `The amount of terminal rows passed to the external programs with the environment variable "$TVIEW_ROWS". Defaults to the row count of the terminal tview is running in.`)
	flag.Parse()

	argv = flag.Args()
	if len(argv) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	viewFile(flags, argv[0])
}
