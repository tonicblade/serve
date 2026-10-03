package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"serve/internal/cli"
	"serve/internal/server"
)

func main() {
	cfg, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrHelp) {
			fmt.Print(cli.Usage())
			return
		}
		fmt.Fprintf(os.Stderr, "serve: %v\n\n%s", err, cli.Usage())
		os.Exit(1)
	}

	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: unable to resolve root path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: unable to access %s: %v\n", root, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "serve: root must be a directory: %s\n", root)
		os.Exit(1)
	}

	if cfg.Expires != "" {
		d, err := time.ParseDuration(cfg.Expires)
		if err != nil {
			fmt.Fprintf(os.Stderr, "serve: invalid expires duration %q\n", cfg.Expires)
			os.Exit(1)
		}
		go func() {
			time.Sleep(d)
			os.Exit(0)
		}()
	}

	if err := server.Start(root, cfg.Host, cfg.Port, cfg.Depth, cfg.Quiet, cfg.Open, cfg.Once); err != nil {
		fmt.Fprintf(os.Stderr, "serve: %v\n", err)
		os.Exit(1)
	}
}
