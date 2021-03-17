package configmanager

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/dyweb/gommon/errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	cmpb "github.com/aws-observability/collector-load-balancer/proto/generated/clb/configmanager"
)

// Server runs within/next to collector as extension/sidecar to update scrape targets
// based on request from CollectorBalancer server.
type Server struct {
	cmpb.UnimplementedConfigManagerServiceServer

	cfg    ServerConfig
	logger *zap.Logger

	grpcSrv *grpc.Server
	httpSrv *http.Server
}

type ServerOptions struct {
	Logger *zap.Logger
}

// NewServer creates the base folder for saving targets files from CollectorLoadBalancer server.
func NewServer(cfg ServerConfig, opts ServerOptions) (*Server, error) {
	if err := os.MkdirAll(cfg.TargetsBaseDir, 0664); err != nil {
		return nil, errors.Wrapf(err, "create base dir for targets failed %s", cfg.TargetsBaseDir)
	}
	return &Server{
		cfg:    cfg,
		logger: opts.Logger,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.startGRPC(ctx); err != nil {
		return err
	}
	return s.startHTTP(ctx)
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpSrv != nil {
		s.logger.Info("Stopping http server", zap.String("Endpoint", s.cfg.HTTP.Endpoint))
		if err := s.httpSrv.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "stop http server failed")
		}
	}

	if s.grpcSrv != nil {
		s.logger.Info("Stopping grpc server", zap.String("Endpoint", s.cfg.GRPC.Endpoint))
		s.grpcSrv.GracefulStop()
	}
	return nil
}

func (s *Server) startGRPC(_ context.Context) error {
	addr := s.cfg.GRPC.Endpoint
	s.logger.Info("Starting grpc server", zap.String("Endpoint", addr))

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, "grpc listen failed on addr %s", addr)
	}
	srv := grpc.NewServer()
	cmpb.RegisterConfigManagerServiceServer(srv, s)
	s.grpcSrv = srv
	go func() {
		if err := srv.Serve(lis); err != nil {
			s.logger.Error("grpc server stopped", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) startHTTP(ctx context.Context) error {
	addr := s.cfg.HTTP.Endpoint
	s.logger.Info("Starting http server", zap.String("Endpoint", addr))

	mux := runtime.NewServeMux()
	if err := cmpb.RegisterConfigManagerServiceHandlerServer(ctx, mux, s); err != nil {
		return errors.Wrap(err, "register http gateway failed")
	}
	httpServ := &http.Server{Addr: addr, Handler: mux}
	s.httpSrv = httpServ
	go func() {
		if err := httpServ.ListenAndServe(); err != nil {
			s.logger.Error("http server stopped", zap.Error(err))
		}
	}()
	return nil
}

// UpdateTargets writes prometheus targets from sever to specified file locations.
// TODO: we could use the context to do something tracing, e.g. time spent from new targets appear until it got distributed to all the collectors.
func (s *Server) UpdatePrometheusFileSD(_ context.Context, req *cmpb.UpdatePrometheusFileSDReq) (*cmpb.UpdatePrometheusFileSDRRes, error) {
	s.logger.Info("Updating files", zap.Int("Count", len(req.Files)))

	merr := errors.NewMultiErr()
	prefix := s.cfg.TargetsBaseDir
	updated := 0
	for _, f := range req.Files {
		// Make sure we don't write file to arbitrary locations.
		if !strings.HasPrefix(f.Location, prefix) {
			// TODO(error): typed error
			merr.Append(errors.Errorf("invalid targets location %s, must have prefix %s", f.Location, prefix))
			continue
		}
		b, err := json.Marshal(f.Targets)
		if merr.Append(errors.Wrapf(err, "encode json failed for file %s", f.Location)) {
			continue
		}
		// TODO: might need to create folder if we allow child folders, e.g. /etc/clb/instance-name/k8spods.json
		err = ioutil.WriteFile(f.Location, b, 0644)
		if merr.Append(errors.Wrapf(err, "write targets failed for file %s", f.Location)) {
			continue
		}
		updated++
	}

	s.logger.Info("Updated files", zap.Int("Count", updated))
	return &cmpb.UpdatePrometheusFileSDRRes{
		UpdatedFiles: int32(updated),
	}, merr.ErrorOrNil()
}

func (s *Server) GetPrometheusFileSD(_ context.Context, req *cmpb.GetPrometheusFileSDReq) (*cmpb.GetPrometheusFileSDRes, error) {
	// TODO: loop the base dir to find all the targets we have
	return nil, errors.Errorf("not implemented")
}
