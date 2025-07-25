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

import { useState, useMemo, useRef } from 'react';
import { Eye, X } from 'lucide-react';
import * as yaml from 'js-yaml';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog';
import CodeMirror from '@uiw/react-codemirror';
import { langs } from '@uiw/codemirror-extensions-langs';
import { catppuccinMocha, catppuccinLatte } from '@catppuccin/codemirror';

interface RawConfigDialogProps {
    name: string;
    rawConfig: string;
    configType?: string; // e.g., "Listener", "Cluster", "Route", etc.
}

export const RawConfigDialog: React.FC<RawConfigDialogProps> = ({
    name,
    rawConfig,
    configType = 'Configuration',
}) => {
    const [format, setFormat] = useState<'yaml' | 'json'>('json');
    const editorRef = useRef<HTMLElement | null>(null);

    // Check if dark mode is enabled
    const isDarkMode = document.documentElement.classList.contains('dark');

    const formattedConfig = useMemo(() => {
        if (!rawConfig) return 'No raw configuration available';

        try {
            if (format === 'yaml') {
                const jsonObj = JSON.parse(rawConfig);
                return yaml.dump(jsonObj, {
                    indent: 4,
                    lineWidth: -1,
                    noRefs: true,
                    sortKeys: false,
                    skipInvalid: true,
                });
            } else {
                // Pretty print JSON
                const jsonObj = JSON.parse(rawConfig);
                return JSON.stringify(jsonObj, null, 4);
            }
        } catch {
            return 'Error parsing configuration';
        }
    }, [rawConfig, format]);

    const extensions = useMemo(() => {
        return format === 'json' ? [langs.json()] : [langs.yaml()];
    }, [format]);

    return (
        <Dialog>
            <DialogTrigger asChild>
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 w-8 p-0 cursor-pointer"
                >
                    <Eye className="h-4 w-4" />
                </Button>
            </DialogTrigger>
            <DialogContent
                className="!max-w-[95vw] !w-[95vw] max-h-[90vh] overflow-hidden flex flex-col"
                showCloseButton={false}
            >
                <DialogHeader>
                    <div className="flex items-center justify-between">
                        <DialogTitle>
                            {configType} Configuration: {name}
                        </DialogTitle>
                        <div className="flex items-center gap-2">
                            <div className="flex bg-muted rounded-lg p-1">
                                <Button
                                    variant={
                                        format === 'yaml' ? 'default' : 'ghost'
                                    }
                                    size="sm"
                                    onClick={() => setFormat('yaml')}
                                    className={`h-7 px-3 rounded-md text-xs font-medium transition-colors cursor-pointer ${
                                        format === 'yaml'
                                            ? 'bg-background text-foreground shadow-sm'
                                            : 'text-muted-foreground hover:text-foreground'
                                    }`}
                                >
                                    YAML
                                </Button>
                                <Button
                                    variant={
                                        format === 'json' ? 'default' : 'ghost'
                                    }
                                    size="sm"
                                    onClick={() => setFormat('json')}
                                    className={`h-7 px-3 rounded-md text-xs font-medium transition-colors cursor-pointer ${
                                        format === 'json'
                                            ? 'bg-background text-foreground shadow-sm'
                                            : 'text-muted-foreground hover:text-foreground'
                                    }`}
                                >
                                    JSON
                                </Button>
                            </div>
                            <DialogClose asChild>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-7 w-7 p-0 cursor-pointer"
                                >
                                    <X className="h-4 w-4" />
                                </Button>
                            </DialogClose>
                        </div>
                    </div>
                    <DialogDescription className="sr-only">
                        View and search through the raw configuration data for{' '}
                        {configType.toLowerCase()}: {name}
                    </DialogDescription>
                </DialogHeader>
                <div className="flex-1 overflow-auto">
                    <CodeMirror
                        ref={editorRef}
                        value={formattedConfig}
                        extensions={extensions}
                        theme={isDarkMode ? catppuccinMocha : catppuccinLatte}
                        readOnly={true}
                        basicSetup={{
                            lineNumbers: true,
                            foldGutter: true,
                            searchKeymap: true,
                            highlightSelectionMatches: true,
                        }}
                        height="100%"
                    />
                </div>
            </DialogContent>
        </Dialog>
    );
};
