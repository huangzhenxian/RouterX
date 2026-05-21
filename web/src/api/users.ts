import { apiClient } from './client';
import type { CreateUserInput, User } from '@/types/user';

interface ListResp {
  items: User[];
  total: number;
}

export async function listUsers(): Promise<ListResp> {
  const { data } = await apiClient.get<ListResp>('/users');
  return data;
}

export async function getUser(id: number): Promise<User> {
  const { data } = await apiClient.get<User>(`/users/${id}`);
  return data;
}

export async function createUser(input: CreateUserInput): Promise<User> {
  const { data } = await apiClient.post<User>('/users', input);
  return data;
}

export async function deleteUser(id: number): Promise<void> {
  await apiClient.delete(`/users/${id}`);
}

export async function enableUser(id: number): Promise<void> {
  await apiClient.post(`/users/${id}/enable`);
}

export async function disableUser(id: number): Promise<void> {
  await apiClient.post(`/users/${id}/disable`);
}
