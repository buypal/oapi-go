package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/buypal/oapi-go/pkg/oapi"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(sigChan)

	cfg, _, err := getConfig()
	if err != nil {
		log.Fatal("config: ", err.Error())
	}

	config, err := cfg.full()
	if err != nil {
		log.Fatal("err config:", err)
	}

	log.SetLevel(log.DebugLevel)

	go func() {
		s := <-sigChan
		log.Infof("received %s signal", s)
		cancel()
	}()

	spec, err := scan(ctx, config)
	if err != nil {
		log.Fatal("err reg:", err)
	}

	data, err := oapi.Format(config.Format, spec)
	if err != nil {
		log.Fatal(err.Error())
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
			log.Fatal(err.Error())
		}
		w, err = os.OpenFile(config.Output, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	_, err = io.Copy(w, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err.Error())
	}
}
