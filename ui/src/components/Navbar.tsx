import { ModeToggle } from './mode-toggle';

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

                    <ModeToggle />
                </div>
            </div>
        </nav>
    );
};
