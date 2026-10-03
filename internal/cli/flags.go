package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrHelp = errors.New("help requested")

type Config struct {
	Root    string
	Port    int
	Host    string
	Depth   int
	Open    bool
	Upload  bool
	Quiet   bool
	Https   bool
	Once    bool
	Expires string
}

func Usage() string {
	return `Usage:
  serve [path] [flags]

Flags:
  --host <host>      bind address (default: 0.0.0.0)
  --port <port>      port to listen on (default: 8080)
  --open             open in browser after startup
  --quiet            suppress banner output
  --upload           allow file uploads
  --depth <n>        folder tree depth (default: 2)
  --https            serve with TLS on :8443
  --once             serve a single request then exit
  --expires <dur>    auto-shutdown after duration like 30m
  --help             show help
`
}

func ParseArgs(args []string) (Config, error) {
	cfg := Config{
		Root:    ".",
		Port:    8080,
		Host:    "0.0.0.0",
		Depth:   2,
		Expires: "",
	}

	positionals := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		a := args[i]

		switch {
		case a == "--":
			positionals = append(positionals, args[i+1:]...)
			goto done
		case a == "-h" || a == "--help":
			return cfg, ErrHelp
		case a == "--open":
			cfg.Open = true
		case a == "--upload":
			cfg.Upload = true
		case a == "--quiet":
			cfg.Quiet = true
		case a == "--https":
			cfg.Https = true
		case a == "--once":
			cfg.Once = true
		case a == "--port":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing value for --port")
			}
			p, err := strconv.Atoi(args[i+1])
			if err != nil || p <= 0 {
				return cfg, fmt.Errorf("invalid port %q", args[i+1])
			}
			cfg.Port = p
			i++
		case strings.HasPrefix(a, "--port="):
			p, err := strconv.Atoi(strings.TrimPrefix(a, "--port="))
			if err != nil || p <= 0 {
				return cfg, fmt.Errorf("invalid port %q", strings.TrimPrefix(a, "--port="))
			}
			cfg.Port = p
		case a == "--host":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing value for --host")
			}
			cfg.Host = args[i+1]
			i++
		case strings.HasPrefix(a, "--host="):
			cfg.Host = strings.TrimPrefix(a, "--host=")
		case a == "--depth":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing value for --depth")
			}
			d, err := strconv.Atoi(args[i+1])
			if err != nil || d <= 0 {
				return cfg, fmt.Errorf("invalid depth %q", args[i+1])
			}
			cfg.Depth = d
			i++
		case strings.HasPrefix(a, "--depth="):
			d, err := strconv.Atoi(strings.TrimPrefix(a, "--depth="))
			if err != nil || d <= 0 {
				return cfg, fmt.Errorf("invalid depth %q", strings.TrimPrefix(a, "--depth="))
			}
			cfg.Depth = d
		case a == "--expires":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing value for --expires")
			}
			cfg.Expires = args[i+1]
			i++
		case strings.HasPrefix(a, "--expires="):
			cfg.Expires = strings.TrimPrefix(a, "--expires=")
		case strings.HasPrefix(a, "-"):
			return cfg, fmt.Errorf("unknown flag: %s", a)
		default:
			positionals = append(positionals, a)
		}
	}

done:
	if len(positionals) > 1 {
		return cfg, fmt.Errorf("too many arguments: %s", strings.Join(positionals, " "))
	}
	if len(positionals) == 1 {
		cfg.Root = positionals[0]
	}
	return cfg, nil
}
