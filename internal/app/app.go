package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kakitomeru/gateway/internal/api/handler"
	"github.com/kakitomeru/gateway/internal/api/middleware"
	"github.com/kakitomeru/gateway/internal/config"
	"github.com/kakitomeru/shared/env"
	"github.com/kakitomeru/shared/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var allowedHeaders = map[string]struct{}{
	"x-request-id": {},
}

type App struct {
	cfg    *config.Config
	host   string
	port   string
	gwmux  *runtime.ServeMux
	router *gin.Engine
}

func NewApp(cfg *config.Config) *App {
	router := gin.New()
	router.Use(gin.Logger())

	router.Use(otelgin.Middleware(
		cfg.Name,
		otelgin.WithSpanNameFormatter(func(c *gin.Context) string {
			return fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		}),
	))

	router.Use(middleware.AuthMiddleware(cfg.ProtectedRoutes))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown endpoint"})
	})

	return &App{
		cfg:    cfg,
		host:   env.GetGatewayHost(),
		port:   env.GetGatewayPort(),
		router: router,
		gwmux: runtime.NewServeMux(
			runtime.WithOutgoingHeaderMatcher(handler.IsHeaderAllowed(allowedHeaders)),
			runtime.WithMetadata(handler.MetadataHandler),
			runtime.WithErrorHandler(handler.ErrorHandler),
			runtime.WithRoutingErrorHandler(handler.RoutingErrorHandler),
		),
	}
}

func (a *App) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shutdownTracer, err := telemetry.NewTracerProvider(ctx, a.cfg.Name, env.GetOtelCollector())
	if err != nil {
		log.Fatalf("failed to create tracer provider: %v", err)
	}
	defer func() {
		if err := shutdownTracer(ctx); err != nil {
			log.Printf("failed to shutdown tracer provider: %v", err)
		}
	}()

	shutdownMeter, err := telemetry.NewMeterProvider(ctx, a.cfg.Name, env.GetOtelCollector())
	if err != nil {
		log.Fatalf("failed to create meter provider: %v", err)
	}
	defer func() {
		if err := shutdownMeter(ctx); err != nil {
			log.Printf("failed to shutdown meter provider: %v", err)
		}
	}()

	var dialOpts = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	if err := handler.SetupHandlers(ctx, a.gwmux, dialOpts); err != nil {
		return err
	}

	a.router.Group("/api/v1/*{grpc_gateway}").Any("", gin.WrapH(a.gwmux))

	srv := &http.Server{
		Addr:    ":" + a.port,
		Handler: a.router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	return a.gracefullyShutdown(ctx, srv)
}

func (a *App) gracefullyShutdown(ctx context.Context, srv *http.Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Received interrupt signal, shutting down server...")
	case <-ctx.Done():
		log.Println("Parent context cancelled, shutting down server...")
	}

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server gracefully stopped")
	return nil
}
