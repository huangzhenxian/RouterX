import { apiClient } from './client';
import type { CreateProviderInput, Provider, TestResult } from '@/types/provider';

export async function listProviders(): Promise<{ items: Provider[]; total: number }> {
  const { data } = await apiClient.get<{ items: Provider[]; total: number }>('/providers');
  return data;
}

export async function createProvider(input: CreateProviderInput): Promise<Provider> {
  const { data } = await apiClient.post<Provider>('/providers', input);
  return data;
}

export async function deleteProvider(id: number): Promise<void> {
  await apiClient.delete(`/providers/${id}`);
}

export async function enableProvider(id: number): Promise<void> {
  await apiClient.post(`/providers/${id}/enable`);
}

export async function disableProvider(id: number): Promise<void> {
  await apiClient.post(`/providers/${id}/disable`);
}

export async function testProvider(id: number): Promise<TestResult> {
  const { data } = await apiClient.post<TestResult>(`/providers/${id}/test`);
  return data;
}
