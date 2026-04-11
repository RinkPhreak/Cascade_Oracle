import { client } from './generated/client.gen';
import { useAuthStore } from '../features/auth/store';

// We set API client options to point to the vite proxy
client.setConfig({
  baseUrl: '',
});

// Add interceptor to append JWT to all requests
client.interceptors.request.use((request) => {
  const token = useAuthStore.getState().token;

  // If the body is FormData, we must strip the Content-Type header
  // to let the browser set it automatically with the correct boundary.
  // In @hey-api/client-fetch, the request being intercepted might still have the original body.
  const isFormData = (request as any).body instanceof FormData;

  if (isFormData) {
    request.headers.delete('Content-Type');
  }

  if (token) {
    const headers = Object.fromEntries(request.headers.entries());
    
    // Safety: ensure it's not re-added during header merging if it was deleted above
    if (isFormData) {
      delete headers['content-type'];
      delete headers['Content-Type'];
    }

    return new Request(request, {
      headers: {
        ...headers,
        'Authorization': `Bearer ${token}`
      }
    });
  }
  return request;
});
export { client };
