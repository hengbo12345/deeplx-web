import { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { KeyRound } from 'lucide-react';

export function AuthModal() {
  const { login } = useAuth();
  const [tokenInput, setTokenInput] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!tokenInput.trim()) {
      setError('请输入令牌');
      return;
    }
    setLoading(true);
    setError('');
    const ok = await login(tokenInput.trim());
    setLoading(false);
    if (!ok) {
      setError('令牌无效，请重新输入');
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-card rounded-lg shadow-lg p-6 w-full max-w-sm mx-4">
        <div className="flex items-center gap-2 mb-4">
          <KeyRound className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">请输入访问令牌</h2>
        </div>
        <form onSubmit={handleSubmit}>
          <input
            type="password"
            value={tokenInput}
            onChange={(e) => setTokenInput(e.target.value)}
            placeholder="输入令牌"
            className="w-full border rounded-md px-3 py-2 mb-3 bg-background text-foreground"
            autoFocus
            disabled={loading}
          />
          {error && (
            <p className="text-sm text-red-500 mb-3">{error}</p>
          )}
          <button
            type="submit"
            className="w-full bg-primary text-primary-foreground rounded-md py-2 hover:bg-primary/90 disabled:opacity-50"
            disabled={loading}
          >
            {loading ? '验证中...' : '确认'}
          </button>
        </form>
      </div>
    </div>
  );
}
