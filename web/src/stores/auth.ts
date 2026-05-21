import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Admin } from '@/types/user';

interface AuthState {
  token: string | null;
  admin: Admin | null;
  expiresAt: string | null;
  setAuth: (token: string, admin: Admin, expiresAt: string) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      admin: null,
      expiresAt: null,
      setAuth: (token, admin, expiresAt) => set({ token, admin, expiresAt }),
      clear: () => set({ token: null, admin: null, expiresAt: null }),
    }),
    { name: 'routex_auth' },
  ),
);
