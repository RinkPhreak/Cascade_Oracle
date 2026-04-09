import { useState } from 'react';
import { useAuthStore } from '../features/auth/store';
import { useNavigate } from 'react-router-dom';
import { useAuthLogin } from '../features/auth/hooks/useAuthLogin';

export const LoginPage = () => {
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');
  const [errorText, setErrorText] = useState('');
  
  const loginAction = useAuthStore(state => state.login);
  const navigate = useNavigate();
  const loginMutation = useAuthLogin();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorText('');

    try {
      const data = await loginMutation.mutateAsync({
        body: { login, password }
      });
      
      if (data?.access_token) {
        loginAction(data.access_token);
        navigate('/');
      }
    } catch (err: unknown) {
      setErrorText((err as Error).message || 'An unexpected error occurred');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-base">
      <div className="bg-bg-surface p-8 rounded-lg border border-neutral-800 w-96 shadow-xl">
        <h1 className="text-3xl font-bold mb-6 text-center tracking-wider text-white">CASCADE <span className="text-accent-cyan">PRO</span></h1>
        {errorText && <div className="mb-4 text-danger text-sm text-center">{errorText}</div>}
        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <input 
              type="text" 
              placeholder="Operator Login" 
              className="w-full bg-neutral-900 border border-neutral-800 p-3 rounded text-white focus:border-accent-cyan outline-none disabled:opacity-50"
              value={login}
              onChange={(e) => setLogin(e.target.value)}
              disabled={loginMutation.isPending}
            />
          </div>
          <div>
            <input 
              type="password" 
              placeholder="Password" 
              className="w-full bg-neutral-900 border border-neutral-800 p-3 rounded text-white focus:border-accent-cyan outline-none disabled:opacity-50"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={loginMutation.isPending}
            />
          </div>
          <button 
            type="submit" 
            className="w-full bg-accent-cyan hover:bg-accent-cyan-hover text-white font-bold py-3 px-4 rounded transition-colors disabled:opacity-50"
            disabled={loginMutation.isPending}
          >
            {loginMutation.isPending ? 'A U T H E N T I C A T I N G ...' : 'A U T H E N T I C A T E'}
          </button>
        </form>
      </div>
    </div>
  );
};
