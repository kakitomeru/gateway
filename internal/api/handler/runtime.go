package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func IsHeaderAllowed(allowedHeaders map[string]struct{}) func(string) (string, bool) {
	return func(s string) (string, bool) {
		if _, ok := allowedHeaders[s]; ok {
			return s, true
		}

		return runtime.DefaultHeaderMatcher(s)
	}
}

func MetadataHandler(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}

	accessToken, ok := ctx.Value("accessToken").(string)
	if ok {
		md.Set("x-access-token", accessToken)
	}

	userID, ok := ctx.Value("userID").(string)
	fmt.Println(userID)
	if ok {
		md.Set("x-user-id", userID)
	}

	return md
}

func ErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	m runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	log.Printf("GRPC Gateway Error: %v, Path: %s", err, r.URL.Path)

	s := status.Convert(err)

	statusCode := runtime.HTTPStatusFromCode(s.Code())
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	if s.Code() == codes.InvalidArgument {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid body request",
		})
	} else if statusCode == http.StatusServiceUnavailable {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "service unavailable",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]string{
			"error": s.Message(),
		})
	}
}

func RoutingErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	m runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	httpStatus int,
) {
	message := "unexpected routing error"
	switch httpStatus {
	case http.StatusBadRequest:
		message = strings.ToLower(http.StatusText(httpStatus))
	case http.StatusMethodNotAllowed:
		message = strings.ToLower(http.StatusText(httpStatus))
	case http.StatusNotFound:
		message = strings.ToLower("unknown endpoint")
	}

	w.WriteHeader(httpStatus)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
