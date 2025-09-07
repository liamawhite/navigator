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

package metrics

import "errors"

var (
	// ErrProviderNotFound indicates that a metrics provider was not found
	ErrProviderNotFound = errors.New("metrics provider not found")

	// ErrProviderNotSupported indicates that a metrics provider type is not supported
	ErrProviderNotSupported = errors.New("metrics provider type not supported")

	// ErrMissingEndpoint indicates that the provider endpoint is missing
	ErrMissingEndpoint = errors.New("metrics provider endpoint is required when enabled")

	// ErrProviderUnavailable indicates that the metrics provider is unavailable
	ErrProviderUnavailable = errors.New("metrics provider is unavailable")

	// ErrInvalidQuery indicates that the metrics query is invalid
	ErrInvalidQuery = errors.New("invalid metrics query")

	// ErrNoData indicates that no metric data was found for the query
	ErrNoData = errors.New("no metric data found")

	// ErrTimeout indicates that the metrics query timed out
	ErrTimeout = errors.New("metrics query timed out")
)
