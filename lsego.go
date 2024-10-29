package lsego

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/helin0815/gowebmonitor/pkg/log"
	"github.com/helin0815/gowebmonitor/pkg/otelutils"
	"google.golang.org/grpc"
)

type middleware func(http.Handler) http.Handler

type Lsego struct {
	grpcSrv *grpc.Server
	grpcMux *runtime.ServeMux

	onShutdown []func()
	root       string
	handlers   []middleware
}

func Default() *Lsego {
	lg := New()
	lg.Use(DefaultLogger)
	return lg
}

func New() *Lsego {
	return &Lsego{
		root:     "/",
		handlers: make([]middleware, 0),
	}
}

func (lg *Lsego) SetRootPath(root string) {
	lg.root = root
}

func (lg *Lsego) RootHandle(handler http.Handler) {
	http.Handle(lg.root, http.StripPrefix(lg.root[:len(lg.root)-1], handler))
}

func (lg *Lsego) Use(h func(handler http.Handler) http.Handler) {
	lg.handlers = append(lg.handlers, h)
}

func (lg *Lsego) UseGrpc(opt ...grpc.ServerOption) *Lsego {
	lg.grpcSrv = grpc.NewServer(opt...)
	return lg
}

func (lg *Lsego) UseGrpcGw(opts ...runtime.ServeMuxOption) *Lsego {
	lg.grpcMux = runtime.NewServeMux(opts...)
	lg.RootHandle(lg.grpcMux)
	return lg
}

func (lg *Lsego) RegisterOnShutdown(f func()) {
	lg.onShutdown = append(lg.onShutdown, f)
}

func (lg *Lsego) HTTPServe(addr string) {
	lg.setupMetricsHandler()
	lg.setupLifecycleHandler()
	srv := &http.Server{
		Addr:    addr,
		Handler: http.DefaultServeMux,
	}

	for i := range lg.handlers {
		srv.Handler = lg.handlers[len(lg.handlers)-1-i](srv.Handler)
	}

	ctx := context.Background()
	shutdown := otelutils.InitProvider(ctx)
	if shutdown != nil {
		defer shutdown(ctx)
	}

	srv.Handler = otelutils.GetHandlerWithOTEL(srv.Handler, "HTTPServe", otelutils.UrlsToIgnore(internalPaths...))

	go func() {
		log.Printf("httpServer listening at %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

	setupSigHandler(srv.Shutdown)
}

func (lg *Lsego) GrpcRegister(f func(s *grpc.Server, gwMux *runtime.ServeMux)) {
	if lg.grpcSrv == nil {
		lg.UseGrpc()
	}
	if lg.grpcMux == nil {
		lg.UseGrpcGw()
	}

	f(lg.grpcSrv, lg.grpcMux)
}

func (lg *Lsego) GRPCServe(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("grpcServer listening at %v", lis.Addr())
		if err := lg.grpcSrv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	setupSigHandler(func(ctx context.Context) error {
		lg.grpcSrv.GracefulStop()
		return nil
	})
}

func (lg *Lsego) GRPCServeWithGateway(grpcAddr string, gwAddr string) {
	go lg.GRPCServe(grpcAddr)
	lg.HTTPServe(gwAddr)
}

func setupSigHandler(shutdown func(ctx context.Context) error) {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown:", err)
	}

	log.Printf("Server exiting")
}
