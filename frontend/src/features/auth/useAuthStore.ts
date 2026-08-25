import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { User } from '../../shared/schemas';

export type { User };

export interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  login: (user: User, token: string, refreshToken?: string | null) => void;
  updateUser: (user: User) => void;
  logout: () => void;
}

// Persiste en sessionStorage (se limpia al cerrar la pestaña), no en
// localStorage: un JWT en localStorage queda expuesto a XSS. Sin Tauri en
// esta fase (F10 no incluye el empaquetado de escritorio, ver
// vite.config.ts) — cuando se añada, esta persistencia deberá conmutar a
// almacenamiento en memoria igual que ../usbi/frontend/src/lib/persistenceStorage.ts.
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,

      login: (user, token, refreshToken = null) => set({ user, token, refreshToken, isAuthenticated: true }),
      updateUser: (user) => set({ user }),
      logout: () => set({ user: null, token: null, refreshToken: null, isAuthenticated: false }),
    }),
    {
      name: 'usbi-anon-auth',
      storage: createJSONStorage(() => sessionStorage),
    }
  )
);
