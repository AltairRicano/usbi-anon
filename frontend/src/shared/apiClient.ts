import axios from 'axios';
import { useAuthStore } from '../features/auth/useAuthStore';
import { AuthResponseSchema } from './schemas';

const apiBaseURL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

export const apiClient = axios.create({
  baseURL: apiBaseURL,
  timeout: 10_000,
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json',
  },
});

// ── Interceptor de request: inyecta el JWT ───────────────────────────────────
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// ── Interceptor de response: refresca el access token en un 401, una sola
// vez por request (RFC 7807 Problem Details, ver internal/httpproblem en Go).
apiClient.interceptors.response.use(
  (response) => response,
  async (error: unknown) => {
    if (axios.isAxiosError(error)) {
      const status = error.response?.status;
      const originalRequest = error.config;

      if (status === 401 && originalRequest && !(originalRequest as { _retry?: boolean })._retry) {
        const refreshToken = useAuthStore.getState().refreshToken;
        if (refreshToken) {
          try {
            (originalRequest as { _retry?: boolean })._retry = true;
            const refreshResponse = await axios.post(`${apiBaseURL}/auth/refresh`, { refresh_token: refreshToken });
            // Validado en runtime (no solo tipado): una respuesta de refresh
            // malformada nunca debe poblar la sesión.
            const data = AuthResponseSchema.parse(refreshResponse.data);
            useAuthStore.getState().login(data.user, data.access_token, data.refresh_token);
            originalRequest.headers.Authorization = `Bearer ${data.access_token}`;
            return apiClient(originalRequest);
          } catch {
            useAuthStore.getState().logout();
            window.dispatchEvent(new CustomEvent('auth:unauthorized'));
          }
        }
      }

      // Solo un 401 significa "esta sesión ya no es válida" (token
      // vencido/revocado). Un 403 es una decisión de autorización sobre esa
      // acción puntual (p. ej. "no puedes borrar una cuenta admin" en
      // AdminAccountsPage) — la sesión sigue siendo válida y NO debe cerrar
      // sesión ni redirigir; la página que hizo la llamada ya sabe mostrar
      // ese error con errorMessage().
      if (status === 401) {
        useAuthStore.getState().logout();
        // App.tsx redirige a /login desde ProtectedRoute al ver este evento.
        window.dispatchEvent(new CustomEvent('auth:unauthorized'));
      }
    }
    return Promise.reject(error);
  }
);
