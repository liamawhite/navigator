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

import { ModeToggle } from './mode-toggle';
import { ClusterSyncStatus } from './ClusterSyncStatus';

export const Navbar: React.FC = () => {
    return (
        <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
            <div className="container mx-auto px-4">
                <div className="flex h-16 items-center justify-between">
                    <div className="flex items-center space-x-3">
                        <div className="w-9 h-9">
                            <img
                                src="/navigator.svg"
                                alt="Navigator"
                                className="w-full h-full"
                            />
                        </div>
                        <div>
                            <h1 className="text-xl font-bold">Navigator</h1>
                            <p className="text-xs text-muted-foreground">
                                Service Registry
                            </p>
                        </div>
                    </div>

                    <div className="flex items-center gap-3">
                        <ClusterSyncStatus />
                        <ModeToggle />
                    </div>
                </div>
            </div>
        </nav>
    );
};
