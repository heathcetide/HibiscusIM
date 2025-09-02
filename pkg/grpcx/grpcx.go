package grpcx

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// ServerConfig gRPC 服务器配置
type ServerConfig struct {
	Addr             string
	UnaryTimeout     time.Duration
	EnableReflection bool
}

// ClientConfig gRPC 客户端配置
type ClientConfig struct {
	Target         string
	UnaryTimeout   time.Duration
	WithInsecure   bool
	DefaultHeaders map[string]string
}

// NewServer 创建 gRPC Server，已内置日志/恢复/超时拦截器
func NewServer(cfg ServerConfig, extra ...grpc.UnaryServerInterceptor) *grpc.Server {
	interceptors := []grpc.UnaryServerInterceptor{
		serverTimeoutInterceptor(cfg.UnaryTimeout),
		recoveryInterceptor(),
	}
	interceptors = append(extra, interceptors...)
	gs := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	if cfg.EnableReflection {
		reflection.Register(gs)
	}
	return gs
}

// Dial 创建客户端连接，内置超时与默认Header注入拦截器
func Dial(cfg ClientConfig, extra ...grpc.UnaryClientInterceptor) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}
	if cfg.WithInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cis := []grpc.UnaryClientInterceptor{
		clientTimeoutInterceptor(cfg.UnaryTimeout),
		clientHeaderInterceptor(cfg.DefaultHeaders),
	}
	cis = append(cis, extra...)
	opts = append(opts, grpc.WithChainUnaryInterceptor(cis...))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return grpc.DialContext(ctx, cfg.Target, opts...)
}

// ---------- Interceptors ----------

func serverTimeoutInterceptor(d time.Duration) grpc.UnaryServerInterceptor {
	if d <= 0 {
		d = 30 * time.Second
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		c, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		return handler(c, req)
	}
}

func recoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				// 返回通用错误，避免进程崩溃
				err = grpc.Errorf(13, "internal error") // codes.Internal
			}
		}()
		return handler(ctx, req)
	}
}

func clientTimeoutInterceptor(d time.Duration) grpc.UnaryClientInterceptor {
	if d <= 0 {
		d = 30 * time.Second
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		c, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		return invoker(c, method, req, reply, cc, opts...)
	}
}

func clientHeaderInterceptor(headers map[string]string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if len(headers) > 0 {
			md := metadata.New(headers)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
