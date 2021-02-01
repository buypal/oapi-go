package main

import (
	"context"

	"github.com/buypal/oapi-go"
	"github.com/buypal/oapi-go/internal/logging"
	"github.com/buypal/oapi-go/internal/oapi/config"
	"github.com/buypal/oapi-go/internal/oapi/spec"
)

func scan(ctx context.Context, log logging.Printer, cfg config.Config) (oapi.OAPI, error) {
	// resolver options
	opts := []oapi.Option{
		oapi.WithLog(log),
	}

	// execution directory
	if len(cfg.Dir) > 0 {
		opts = append(opts, oapi.WithDir(cfg.Dir))
	}

	if len(cfg.Overrides) > 0 {
		opts = append(opts, oapi.WithOverride(cfg.Overrides))
	}

	if len(cfg.Operations) > 0 {
		opts = append(opts, oapi.WithDefOps(cfg.Operations))
	}

	opts = append(opts, oapi.WithRootSchema(spec.OpenAPI{
		Info:         cfg.Info,
		Servers:      cfg.Servers,
		Components:   cfg.Components,
		Security:     cfg.Security,
		Tags:         cfg.Tags,
		ExternalDocs: cfg.ExternalDocs,
	}))

	return oapi.Scan(ctx, opts...)
}
