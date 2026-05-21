import { apiClient } from './client';
import type { CreateNodeInput, CreateNodeResult, Node } from '@/types/node';

export async function listNodes(): Promise<{ items: Node[]; total: number }> {
  const { data } = await apiClient.get<{ items: Node[]; total: number }>('/nodes');
  return data;
}

export async function createNode(input: CreateNodeInput): Promise<CreateNodeResult> {
  const { data } = await apiClient.post<CreateNodeResult>('/nodes', input);
  return data;
}

export async function deleteNode(id: number): Promise<void> {
  await apiClient.delete(`/nodes/${id}`);
}
