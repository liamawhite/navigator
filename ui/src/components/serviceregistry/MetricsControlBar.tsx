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

import React from 'react';
import { RefreshCw, Clock } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Card, CardContent } from '@/components/ui/card';
import { useMetricsContext, TIME_RANGES } from '../../contexts/MetricsContext';

export const MetricsControlBar: React.FC = () => {
    const {
        timeRange,
        lastUpdated,
        isRefreshing,
        setTimeRange,
        triggerRefresh,
    } = useMetricsContext();

    const formatLastUpdated = (date: Date | null) => {
        if (!date) return 'Never';

        const now = new Date();
        const diff = Math.floor((now.getTime() - date.getTime()) / 1000);

        if (diff < 60) return `${diff}s ago`;
        if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
        return date.toLocaleTimeString();
    };

    return (
        <Card className="mb-6">
            <CardContent className="py-3">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2">
                            <Clock className="w-4 h-4 text-muted-foreground" />
                            <span className="text-sm font-medium">
                                Time Range:
                            </span>
                        </div>
                        <Select
                            value={timeRange.value}
                            onValueChange={(value) => {
                                const selectedRange = TIME_RANGES.find(
                                    (range) => range.value === value
                                );
                                if (selectedRange) {
                                    setTimeRange(selectedRange);
                                }
                            }}
                        >
                            <SelectTrigger className="w-48">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {TIME_RANGES.map((range) => (
                                    <SelectItem
                                        key={range.value}
                                        value={range.value}
                                    >
                                        {range.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={triggerRefresh}
                            disabled={isRefreshing}
                            className="flex items-center gap-2"
                        >
                            <RefreshCw
                                className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`}
                            />
                            Refresh
                        </Button>
                    </div>

                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <span>Last updated:</span>
                        <span className="font-medium">
                            {formatLastUpdated(lastUpdated)}
                        </span>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
};
