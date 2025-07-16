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
