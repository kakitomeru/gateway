package handler

import (
	"context"
	"log"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	authpb "github.com/kakitomeru/auth/pkg/pb/v1"
	"github.com/kakitomeru/shared/env"
	snippetpb "github.com/kakitomeru/snippet/pkg/pb/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

func SetupHandlers(ctx context.Context, mux *runtime.ServeMux, dialOpts []grpc.DialOption) error {
	// Add retry backoff configuration
	backoffConfig := backoff.Config{
		BaseDelay:  1.0 * time.Second,
		Multiplier: 1.6,
		Jitter:     0.2,
		MaxDelay:   120 * time.Second,
	}

	// Add connection backoff to dial options
	dialOpts = append(dialOpts,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoffConfig,
			MinConnectTimeout: 5 * time.Second,
		}),
	)

	// Setup auth service with retry
	authEndpoint := env.GetAuthHost() + ":" + env.GetAuthPort()
	log.Printf("Connecting to auth service at %s", authEndpoint)
	if err := authpb.RegisterAuthServiceHandlerFromEndpoint(
		ctx, mux, authEndpoint, dialOpts,
	); err != nil {
		return err
	}

	// Setup snippet service with retry
	snippetEndpoint := env.GetSnippetHost() + ":" + env.GetSnippetPort()
	log.Printf("Connecting to snippet service at %s", snippetEndpoint)
	if err := snippetpb.RegisterSnippetServiceHandlerFromEndpoint(
		ctx, mux, snippetEndpoint, dialOpts,
	); err != nil {
		return err
	}

	return nil
}
