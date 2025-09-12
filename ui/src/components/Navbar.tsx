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
import { Button } from './ui/button';
import { Link } from 'react-router-dom';
import { List } from 'lucide-react';

export const Navbar: React.FC = () => {
    return (
        <nav className="bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
            <div className="container mx-auto px-4">
                <div className="flex h-16 items-center justify-between">
                    <div className="flex items-center space-x-2">
                        <div className="w-9 h-9">
                            <img
                                src="/navigator.svg"
                                alt="Navigator"
                                className="w-full h-full"
                            />
                        </div>

                        <div className="flex items-center space-x-2">
                            <Button variant="secondary" size="sm" asChild>
                                <Link
                                    to="/"
                                    className="flex items-center gap-2"
                                >
                                    <List className="h-4 w-4" />
                                    Service Registry
                                </Link>
                            </Button>
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
