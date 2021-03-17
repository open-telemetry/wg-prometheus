package configmanagerext

import (
	"context"

	"go.opentelemetry.io/collector/component"

	"github.com/aws-observability/collector-load-balancer/configmanager"
)

// extension.go implements the extension interface

var _ component.Extension = (*extension)(nil)

type extension struct {
	srv *configmanager.Server
}

func (e *extension) Start(_ context.Context, host component.Host) error {
	ctx := context.Background()
	return e.srv.Start(ctx)
}

func (e *extension) Shutdown(ctx context.Context) error {
	return e.srv.Stop(ctx)
}
