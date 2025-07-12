import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { HomePage } from './pages/HomePage';
import { ServiceDetailPage } from './pages/ServiceDetailPage';

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
        <QueryClientProvider client={queryClient}>
            <Router>
                <div className="min-h-screen bg-gray-50">
                    <Routes>
                        <Route path="/" element={<HomePage />} />
                        <Route
                            path="/service/:id"
                            element={<ServiceDetailPage />}
                        />
                    </Routes>
                </div>
            </Router>
        </QueryClientProvider>
    );
}

export default App;
