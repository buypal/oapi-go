package main

import (
	"context"

	"github.com/buypal/oapi-go/pkg/oapi"
	"github.com/buypal/oapi-go/pkg/oapi/spec"
	"github.com/buypal/oapi-go/pkg/ocfg"
	"github.com/buypal/oapi-go/pkg/resolver"
	log "github.com/sirupsen/logrus"
)

func scan(ctx context.Context, cfg ocfg.Config) (oapi.OAPI, error) {
	// resolver options
	opts := []resolver.Option{}

	// todo
	opts = append(opts, resolver.WithLog(log.StandardLogger()))

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
