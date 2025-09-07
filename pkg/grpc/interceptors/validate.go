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

package interceptors

import (
	"context"
	"log/slog"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ValidationInterceptor creates a gRPC unary interceptor that validates requests using protovalidate
func ValidationInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	validator, err := protovalidate.New()
	if err != nil {
		logger.Error("failed to create protovalidate validator", "error", err)
		// Return a no-op interceptor if validation setup fails
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Only validate proto messages
		if msg, ok := req.(proto.Message); ok {
			if err := validator.Validate(msg); err != nil {
				logger.Warn("validation failed", "method", info.FullMethod, "error", err)
				return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
			}
		}

		return handler(ctx, req)
	}
}

// StreamValidationInterceptor creates a gRPC stream interceptor that validates requests using protovalidate
func StreamValidationInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	validator, err := protovalidate.New()
	if err != nil {
		logger.Error("failed to create protovalidate validator", "error", err)
		// Return a no-op interceptor if validation setup fails
		return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return handler(srv, stream)
		}
	}

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrappedStream := &validatingStream{
			ServerStream: stream,
			validator:    validator,
			logger:       logger,
			method:       info.FullMethod,
		}
		return handler(srv, wrappedStream)
	}
}

// validatingStream wraps a grpc.ServerStream to validate incoming messages
type validatingStream struct {
	grpc.ServerStream
	validator protovalidate.Validator
	logger    *slog.Logger
	method    string
}

// RecvMsg validates incoming messages before passing them to the handler
func (s *validatingStream) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	// Only validate proto messages
	if msg, ok := m.(proto.Message); ok {
		if err := s.validator.Validate(msg); err != nil {
			s.logger.Warn("stream validation failed", "method", s.method, "error", err)
			return status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
		}
	}

	return nil
}
