package main

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/buypal/oapi-go/pkg/oapi/config"
	"github.com/sirupsen/logrus"
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
		Default("nolog").
		EnumVar(&cfg.LogLevel, "nolog", "panic", "fatal", "error", "warn", "info", "debug", "trace")

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

func (ff Config) logrus() *logrus.Logger {
	// Instantiate logger
	logrs := logrus.New()
	logrs.SetLevel(logrus.FatalLevel)

	// parse level
	lx, err := logrus.ParseLevel(ff.LogLevel)
	if err != nil {
		logrs.Fatal(err.Error())
	}
	logrs.SetLevel(lx)
	return logrs
}

func (ff Config) full() (cfg config.Config, err error) {
	wd, _ := os.Getwd()

	if len(ff.Config) > 0 {
		// path
		path := toAbsPath(ff.Config, wd)
		// read config
		cfg, err = config.ReadFile(path)
		if err != nil {
			return
		}
	}

	// directory of execution
	dir := wd
	if len(cfg.Dir) > 0 {
		dir = toAbsPath(cfg.Dir, filepath.Dir(cfg.FilePath))
	}
	// if flags are present let them overwrite
	if len(ff.Dir) > 0 {
		dir = toAbsPath(ff.Dir, wd)
	}
	cfg.Dir = dir

	// format of openapi
	if len(cfg.Format) == 0 {
		cfg.Format = "json:pretty"
	}
	if len(ff.Format) > 0 {
		cfg.Format = ff.Format
	}

	// output destination
	output := "stdout"
	if len(cfg.Output) > 0 {
		switch cfg.Output {
		case "stderr", "stdout":
		default:
			output = toAbsPath(cfg.Output, cfg.Dir)
		}
	}
	if len(ff.Output) > 0 {
		switch ff.Output {
		case "stderr", "stdout":
		default:
			output = toAbsPath(ff.Output, wd)
		}
	}
	cfg.Output = output

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
