package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/buypal/oapi-go"
	"github.com/buypal/oapi-go/internal/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(sigChan)

	cfg, _, err := getConfig()
	if err != nil {
		panic(err)
	}

	var log logging.Printer

	if cfg.LogLevel != "nolog" {
		// Instantiate logger
		logger := cfg.logrus()
		log = logging.NewLogger(func(lvl logging.Level, msg string, args ...interface{}) {
			logger.Logf(logrus.Level(lvl), msg, args...)
		})
	} else {
		log = logging.Void()
	}

	config, err := cfg.full()
	if err != nil {
		logging.Fatal(log, "err config: %s", err.Error())
	}

	go func() {
		s := <-sigChan
		logging.Info(log, "received %s signal", s)
		cancel()
	}()

	spec, err := scan(ctx, log, config)
	if err != nil {
		logging.Fatal(log, "err during scan: %s", err.Error())
	}

	data, err := oapi.Format(config.Format, spec)
	if err != nil {
		logging.Fatal(log, "err during format: %s", err.Error())
	}

	var w io.Writer

	switch config.Output {
	case "stdout", "":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	default:
		err = os.Remove(config.Output)
		if err != nil {
			logging.Fatal(log, err.Error())
		}
		w, err = os.OpenFile(config.Output, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			logging.Fatal(log, err.Error())
		}
	}

	_, err = io.Copy(w, bytes.NewBuffer(data))
	if err != nil {
		logging.Fatal(log, err.Error())
	}
}
