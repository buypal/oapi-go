package main

import (
	"context"

	"github.com/buypal/oapi-go/internal/logging"
	"github.com/buypal/oapi-go/internal/oapi"
	"github.com/buypal/oapi-go/internal/oapi/config"
	"github.com/buypal/oapi-go/internal/oapi/resolver"
	"github.com/buypal/oapi-go/internal/oapi/spec"
)

func scan(ctx context.Context, log logging.Printer, cfg config.Config) (oapi.OAPI, error) {
	// resolver options
	opts := []resolver.Option{
		resolver.WithLog(log),
	}

	// execution directory
	if len(cfg.Dir) > 0 {
		opts = append(opts, resolver.WithDir(cfg.Dir))
	}

	if len(cfg.Overrides) > 0 {
		opts = append(opts, resolver.WithOverride(cfg.Overrides))
	}

	if len(cfg.Operations) > 0 {
		opts = append(opts, resolver.WithDefOps(cfg.Operations))
	}

	opts = append(opts, resolver.WithRootSchema(spec.OpenAPI{
		Info:         cfg.Info,
		Servers:      cfg.Servers,
		Components:   cfg.Components,
		Security:     cfg.Security,
		Tags:         cfg.Tags,
		ExternalDocs: cfg.ExternalDocs,
	}))

	return resolver.Scan(ctx, opts...)
}
