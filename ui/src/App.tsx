import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from './components/theme-provider';
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
        <ThemeProvider defaultTheme="system" storageKey="navigator-ui-theme">
            <QueryClientProvider client={queryClient}>
                <Router>
                    <Routes>
                        <Route path="/" element={<HomePage />} />
                        <Route
                            path="/service/:id"
                            element={<ServiceDetailPage />}
                        />
                    </Routes>
                </Router>
            </QueryClientProvider>
        </ThemeProvider>
    );
}

export default App;
