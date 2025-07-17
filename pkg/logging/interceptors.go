// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that logs requests and responses
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		requestID := GenerateRequestID()

		// Create request-scoped logger
		requestLogger := For(ComponentGRPC).With(
			"request_id", requestID,
			"method", info.FullMethod,
		)

		// Add client information if available
		if p, ok := peer.FromContext(ctx); ok {
			requestLogger = requestLogger.With("client_addr", p.Addr.String())
		}

		// Add logger and request ID to context
		ctx = WithRequestID(ctx, requestID)
		ctx = WithLogger(ctx, requestLogger)

		requestLogger.Debug("grpc request started")

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Log the result
		if err != nil {
			code := status.Code(err)
			requestLogger.Error("grpc request failed",
				"duration_ms", duration.Milliseconds(),
				"error", err.Error(),
				"grpc_code", code.String(),
			)
		} else {
			requestLogger.Info("grpc request completed",
				"duration_ms", duration.Milliseconds(),
			)
		}

		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC streaming server interceptor that logs stream operations
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		requestID := GenerateRequestID()

		// Create request-scoped logger
		requestLogger := For(ComponentGRPC).With(
			"request_id", requestID,
			"method", info.FullMethod,
		)

		// Add client information if available
		if p, ok := peer.FromContext(stream.Context()); ok {
			requestLogger = requestLogger.With("client_addr", p.Addr.String())
		}

		// Create a wrapped stream with context
		ctx := WithRequestID(stream.Context(), requestID)
		ctx = WithLogger(ctx, requestLogger)
		wrappedStream := &wrappedServerStream{stream, ctx}

		requestLogger.Debug("grpc stream started")

		// Call the handler
		err := handler(srv, wrappedStream)

		// Calculate duration
		duration := time.Since(start)

		// Log the result
		if err != nil {
			code := status.Code(err)
			requestLogger.Error("grpc stream failed",
				"duration_ms", duration.Milliseconds(),
				"error", err.Error(),
				"grpc_code", code.String(),
			)
		} else {
			requestLogger.Info("grpc stream completed",
				"duration_ms", duration.Milliseconds(),
			)
		}

		return err
	}
}

// wrappedServerStream wraps a grpc.ServerStream to provide context with logging
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
