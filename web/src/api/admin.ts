import { apiClient } from './client';
import type { LoginResult } from '@/types/user';

export async function login(username: string, password: string): Promise<LoginResult> {
  const { data } = await apiClient.post<LoginResult>('/admin/login', { username, password });
  return data;
}
