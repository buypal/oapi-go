package main

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/buypal/oapi-go/pkg/ocfg"
)

// Config represents new configuration
type Config struct {
	LogLevel string
	Config   string
	Dir      string
	Format   string
	Output   string

	Usage func()
}

func getConfig() (cfg Config, cmd string, err error) {
	app := kingpin.New("oapi", "openapi v3 tool")
	app.Author("Richard Hutta")

	cfg.Usage = func() { app.Usage(os.Args[1:]) }

	app.Flag("loglevel", "will set loglevel").
		EnumVar(&cfg.LogLevel, "nolog", "errors", "info", "debug", "trace")

	app.Flag("config", "config to be used").
		StringVar(&cfg.Config)

	app.Flag("dir", "execution directory usually dir of main pkg").
		StringVar(&cfg.Dir)

	app.Flag("format", "will set output format").
		EnumVar(&cfg.Format, "json", "yaml", "yml", "json:pretty", "go")

	app.Flag("output", "will set output destination").
		StringVar(&cfg.Output)

	// Parse
	cmd, err = app.Parse(os.Args[1:])
	return
}

func (ff Config) full() (config ocfg.Config, err error) {
	wd, _ := os.Getwd()

	if len(ff.Config) > 0 {
		// path
		path := toAbsPath(ff.Config, wd)
		// read config
		config, err = ocfg.ReadFile(path)
		if err != nil {
			return
		}
	}

	// directory of execution
	dir := wd
	if len(config.Dir) > 0 {
		dir = toAbsPath(config.Dir, filepath.Dir(config.FilePath))
	}
	// if flags are present let them overwrite
	if len(ff.Dir) > 0 {
		dir = toAbsPath(ff.Dir, wd)
	}
	config.Dir = dir

	// format of openapi
	if len(config.Format) == 0 {
		config.Format = "json:pretty"
	}
	if len(ff.Format) > 0 {
		config.Format = ff.Format
	}

	// output destination
	output := "stdout"
	if len(config.Output) > 0 {
		switch config.Output {
		case "stderr", "stdout":
		default:
			output = toAbsPath(config.Output, config.Dir)
		}
	}
	if len(ff.Output) > 0 {
		switch ff.Output {
		case "stderr", "stdout":
		default:
			output = toAbsPath(ff.Output, wd)
		}
	}
	config.Output = output

	// if len(flags.LogLevel) > 0 {
	// 	config.LogLevel = flags.LogLevel
	// }

	return
}

func toAbsPath(path string, fallback string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(fallback, path)
}
