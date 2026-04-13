import { Link } from 'react-router-dom';
import { Languages, Home, FileText, LogOut } from 'lucide-react';
import { Button } from './ui/button';
import { useAuth } from '../contexts/AuthContext';

export function Header() {
  const { isAuthenticated, logout } = useAuth();

  return (
    <header className="border-b">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2">
            <Languages className="h-8 w-8 text-primary" />
            <div>
              <h1 className="text-2xl font-bold">DeepLX Web</h1>
              <p className="text-xs text-muted-foreground">Free Translation Service</p>
            </div>
          </Link>

          <nav className="flex items-center gap-2">
            <Link to="/">
              <Button variant="ghost">
                <Home className="mr-2 h-4 w-4" />
                Text
              </Button>
            </Link>
            <Link to="/document">
              <Button variant="ghost">
                <FileText className="mr-2 h-4 w-4" />
                Document
              </Button>
            </Link>
            {isAuthenticated && (
              <Button variant="ghost" onClick={logout}>
                <LogOut className="mr-2 h-4 w-4" />
                退出
              </Button>
            )}
          </nav>
        </div>
      </div>
    </header>
  );
}
