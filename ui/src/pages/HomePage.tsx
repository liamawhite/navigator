import { ServiceList } from '../components/ServiceList';
import { Navbar } from '../components/Navbar';
import { useNavigate } from 'react-router-dom';

export const HomePage: React.FC = () => {
    const navigate = useNavigate();

    const handleServiceSelect = (serviceId: string) => {
        navigate(`/service/${serviceId}`);
    };

    return (
        <div className="min-h-screen bg-background">
            <Navbar />
            <div className="container mx-auto px-4 py-8">
                <ServiceList onServiceSelect={handleServiceSelect} />
            </div>
        </div>
    );
};
