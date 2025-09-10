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

import React, { Component, ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

interface Props {
    children: ReactNode;
    fallbackClassName?: string;
}

interface State {
    hasError: boolean;
    error?: Error;
    errorInfo?: string;
}

/**
 * Error boundary specifically designed for graph visualization errors
 * Provides graceful fallback and recovery options
 */
export class GraphErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = { hasError: false };
    }

    static getDerivedStateFromError(error: Error): State {
        return {
            hasError: true,
            error,
        };
    }

    componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
        console.error('Graph rendering error:', error, errorInfo);

        // Log additional context for debugging
        console.error('Component stack:', errorInfo.componentStack);

        this.setState({
            hasError: true,
            error,
            errorInfo: errorInfo.componentStack,
        });

        // Report to monitoring service if available
        // TODO: Add error reporting integration
    }

    handleRetry = () => {
        this.setState({
            hasError: false,
            error: undefined,
            errorInfo: undefined,
        });
    };

    render() {
        if (this.state.hasError) {
            const isDevelopment = process.env.NODE_ENV === 'development';

            return (
                <Card
                    className={`border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950 ${this.props.fallbackClassName || ''}`}
                >
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-red-700 dark:text-red-400">
                            <AlertTriangle className="w-5 h-5" />
                            Graph Rendering Error
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="text-red-600 dark:text-red-300">
                            <p className="font-medium mb-2">
                                The service graph failed to render due to an
                                unexpected error.
                            </p>

                            {isDevelopment && this.state.error && (
                                <div className="bg-red-100 dark:bg-red-900/50 p-3 rounded-md text-sm font-mono">
                                    <p className="font-semibold">Error:</p>
                                    <p className="text-red-800 dark:text-red-200">
                                        {this.state.error.message}
                                    </p>

                                    {this.state.error.stack && (
                                        <details className="mt-2">
                                            <summary className="cursor-pointer font-semibold">
                                                Stack Trace
                                            </summary>
                                            <pre className="mt-1 text-xs whitespace-pre-wrap overflow-auto max-h-32">
                                                {this.state.error.stack}
                                            </pre>
                                        </details>
                                    )}
                                </div>
                            )}
                        </div>

                        <div className="flex items-center gap-3">
                            <Button
                                onClick={this.handleRetry}
                                variant="outline"
                                size="sm"
                                className="flex items-center gap-2"
                            >
                                <RefreshCw className="w-4 h-4" />
                                Retry
                            </Button>

                            <Button
                                onClick={() => window.location.reload()}
                                variant="outline"
                                size="sm"
                            >
                                Reload Page
                            </Button>
                        </div>

                        <div className="text-sm text-muted-foreground">
                            <p>
                                This error has been logged for investigation. If
                                the problem persists, try refreshing the page or
                                contact support.
                            </p>
                        </div>
                    </CardContent>
                </Card>
            );
        }

        return this.props.children;
    }
}
