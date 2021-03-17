package configmanager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServer_Start(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.TargetsBaseDir = "./testdata"
	srv, err := NewServer(cfg, ServerOptions{Logger: zap.NewExample()})
	require.Nil(t, err)
	require.Nil(t, srv.Start(context.Background()))
	// http://localhost:8520/v1/prometheus/targets/
	// FIXME: use an http client to valid these api works
	time.Sleep(30 * time.Second)
	require.Nil(t, srv.Stop(context.Background()))
}
