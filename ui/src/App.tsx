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

import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from './components/theme-provider';
import { HomePage } from './pages/HomePage';
import { ServiceDetailPage } from './pages/ServiceDetailPage';
import { ServiceInstanceDetailPage } from './pages/ServiceInstanceDetailPage';
import { TopologyPage } from './pages/TopologyPage';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            retry: 1,
            refetchOnWindowFocus: false,
        },
    },
});

function App() {
    return (
        <ThemeProvider defaultTheme="system" storageKey="navigator-ui-theme">
            <QueryClientProvider client={queryClient}>
                <Router>
                    <Routes>
                        <Route path="/" element={<HomePage />} />
                        <Route path="/topology" element={<TopologyPage />} />
                        <Route
                            path="/services/:id"
                            element={<ServiceDetailPage />}
                        />
                        <Route
                            path="/services/:serviceId/instances/:instanceId"
                            element={<ServiceInstanceDetailPage />}
                        />
                    </Routes>
                </Router>
            </QueryClientProvider>
        </ThemeProvider>
    );
}

export default App;
