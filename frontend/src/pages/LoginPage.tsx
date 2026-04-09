import { useState } from 'react';
import { useAuthStore } from '../features/auth/store';
import { useNavigate } from 'react-router-dom';
import { client } from '../api/client';
import { postApiV1AuthLogin } from '../api/generated';

export const LoginPage = () => {
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  
  const loginAction = useAuthStore(state => state.login);
  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const { data, error } = await postApiV1AuthLogin({
        body: { login, password }
      });
      
      if (error) {
        setError(error.message || 'Invalid credentials');
        return;
      }
      
      if (data?.access_token) {
        loginAction(data.access_token);
        navigate('/');
      }
    } catch (err) {
      setError('An unexpected error occurred');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-base">
      <div className="bg-bg-surface p-8 rounded-lg border border-neutral-800 w-96 shadow-xl">
        <h1 className="text-3xl font-bold mb-6 text-center tracking-wider text-white">CASCADE <span className="text-accent-cyan">PRO</span></h1>
        {error && <div className="mb-4 text-danger text-sm text-center">{error}</div>}
        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <input 
              type="text" 
              placeholder="Operator Login" 
              className="w-full bg-neutral-900 border border-neutral-800 p-3 rounded text-white focus:border-accent-cyan outline-none"
              value={login}
              onChange={(e) => setLogin(e.target.value)}
            />
          </div>
          <div>
            <input 
              type="password" 
              placeholder="Password" 
              className="w-full bg-neutral-900 border border-neutral-800 p-3 rounded text-white focus:border-accent-cyan outline-none"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          <button 
            type="submit" 
            className="w-full bg-accent-cyan hover:bg-accent-cyan-hover text-white font-bold py-3 px-4 rounded transition-colors"
          >
            A U T H E N T I C A T E
          </button>
        </form>
      </div>
    </div>
  );
};
