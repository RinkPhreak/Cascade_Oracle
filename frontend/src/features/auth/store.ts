import { create } from 'zustand';

interface AuthState {
  token: string | null;
  login: (token: string) => void;
  logout: () => void;
  isAuthenticated: () => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem('cascade_token'),
  login: (token: string) => {
    localStorage.setItem('cascade_token', token);
    set({ token });
  },
  logout: () => {
    localStorage.removeItem('cascade_token');
    set({ token: null });
  },
  isAuthenticated: () => !!get().token,
}));
