import { Moon, Sun } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useTheme } from '@/components/theme-provider';

export function ModeToggle() {
    const { theme, setTheme } = useTheme();

    const toggleTheme = () => {
        if (theme === 'light') {
            setTheme('dark');
        } else {
            setTheme('light');
        }
    };

    return (
        <Button
            variant="outline"
            size="icon"
            onClick={toggleTheme}
            className="cursor-pointer"
        >
            <Moon className="h-[1.2rem] w-[1.2rem] transition-all dark:hidden" />
            <Sun className="h-[1.2rem] w-[1.2rem] transition-all hidden dark:block" />
            <span className="sr-only">Toggle theme</span>
        </Button>
    );
}
