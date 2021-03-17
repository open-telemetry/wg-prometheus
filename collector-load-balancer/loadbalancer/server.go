package loadbalancer

import (
	"context"
	"sync"

	"github.com/dyweb/gommon/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	otelv1 "github.com/aws-observability/collector-load-balancer/api/v1"
)

// Server contains multiple instances
type Server struct {
	ctx         context.Context
	logger      *zap.Logger
	childLogger *zap.Logger // TODO: zap does not dedupe field, so we keep one w/o Component: clb/Server

	// K8S only
	crdScheme *runtime.Scheme
	k8sClient client.Client

	mu        sync.Mutex
	instances map[InstanceName]*Instance
}

type ServerOptions struct {
	Logger    *zap.Logger
	CRDScheme *runtime.Scheme
	K8sClient client.Client
}

// ctx should only get cancelled when the process is going to exit.
func NewServer(opts ServerOptions) (*Server, error) {
	logger := opts.Logger.With(zap.String("Component", "clb/Server"))
	// log something so we panic early on nil logger.
	logger.Info("Creating load balancer server")
	return &Server{
		ctx:         context.Background(),
		logger:      logger,
		childLogger: opts.Logger,
		crdScheme:   opts.CRDScheme,
		k8sClient:   opts.K8sClient,
		instances:   make(map[InstanceName]*Instance),
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting load balancer server")
	s.ctx = ctx

	// TODO: should starts its own grpc server to provide API when running outside k8s
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}

func (s *Server) AddInstanceIfNotExists(clb otelv1.CollectorLoadBalancer) (*Instance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := NewInstanceName(clb)
	if ins, ok := s.instances[name]; ok {
		return ins, nil
	}
	ins, err := NewInstance(InstanceOptions{
		Logger:       s.childLogger,
		CRDScheme:    s.crdScheme,
		K8sClient:    s.k8sClient,
		Name:         name,
		InitialState: clb,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create instance failed")
	}
	s.instances[name] = ins

	logger := s.logger.With(zap.String("InstanceName", name.String()))
	go func() {
		if err := ins.Run(s.ctx); err != nil {
			logger.Error("Instance exits with error", zap.Error(err))
			// TODO: send the error out
		} else {
			logger.Info("Instance stopped")
		}
	}()
	return ins, nil
}

func (s *Server) RemoveInstance(ctx context.Context, name InstanceName) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	instance, ok := s.instances[name]
	if !ok {
		s.logger.Warn("Removing instance that is already gone", zap.String("InstanceName", name.String()))
		return nil
	}
	s.logger.Info("Removing instance", zap.String("InstanceName", name.String()))
	if err := instance.Stop(ctx); err != nil {
		s.logger.Error("Remove instance failed", zap.Error(err), zap.String("InstanceName", name.String()))
		return err
	}
	delete(s.instances, name)
	s.logger.Info("Removed instance", zap.String("InstanceName", name.String()))
	return nil
}
