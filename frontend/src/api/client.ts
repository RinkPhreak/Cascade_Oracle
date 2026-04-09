import { client } from './generated/client.gen';
import { useAuthStore } from '../features/auth/store';

// We set API client options to point to the vite proxy
client.setConfig({
  baseUrl: '',
});

// Add interceptor to append JWT to all requests
client.interceptors.request.use((request) => {
  const token = useAuthStore.getState().token;
  if (token) {
    request.headers.set('Authorization', `Bearer ${token}`);
  }
  return request;
});
export { client };
