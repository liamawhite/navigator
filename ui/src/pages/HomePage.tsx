import { ServiceList } from '../components/ServiceList';
import { useNavigate } from 'react-router-dom';

export const HomePage: React.FC = () => {
    const navigate = useNavigate();

    const handleServiceSelect = (serviceId: string) => {
        navigate(`/service/${serviceId}`);
    };

    return (
        <div className="container mx-auto px-4 py-8">
            <div className="mb-8">
                <h1 className="text-3xl font-bold text-gray-900 mb-2">
                    Navigator Service Registry
                </h1>
                <p className="text-gray-600">
                    Discover and manage Kubernetes services in your cluster
                </p>
            </div>

            <ServiceList onServiceSelect={handleServiceSelect} />
        </div>
    );
};
