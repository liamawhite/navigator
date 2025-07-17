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

import { useState } from 'react';
import { Copy, Circle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { RawConfigDialog } from '@/components/envoy/RawConfigDialog';

interface ConfigActionsProps {
    name: string;
    rawConfig: string;
    configType: string;
    copyId: string;
    showViewButton?: boolean;
}

export const ConfigActions: React.FC<ConfigActionsProps> = ({
    name,
    rawConfig,
    configType,
    copyId,
    showViewButton = true,
}) => {
    const [copiedItem, setCopiedItem] = useState<string | null>(null);

    const copyToClipboard = async (text: string, itemId: string) => {
        try {
            await navigator.clipboard.writeText(text);
            setCopiedItem(itemId);
            setTimeout(() => setCopiedItem(null), 2000);
        } catch (err) {
            console.error('Failed to copy: ', err);
        }
    };

    if (!rawConfig) return null;

    return (
        <div className="flex items-center gap-1">
            {showViewButton && (
                <RawConfigDialog
                    name={name}
                    rawConfig={rawConfig}
                    configType={configType}
                />
            )}
            <Button
                variant="ghost"
                size="sm"
                className="cursor-pointer"
                onClick={() => copyToClipboard(rawConfig, copyId)}
            >
                {copiedItem === copyId ? (
                    <Circle className="w-4 h-4 fill-green-500 text-green-500" />
                ) : (
                    <Copy className="w-4 h-4" />
                )}
            </Button>
        </div>
    );
};
