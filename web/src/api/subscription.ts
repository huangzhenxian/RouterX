import { apiClient } from './client';

export interface PerNodeLink {
  node_name: string;
  host: string;
  port: number;
  url: string;
}

export interface SubscriptionPayload {
  username: string;
  uuid: string;
  links: PerNodeLink[];
  base64: string;
  raw_text: string;
}

export interface SubscriptionResp {
  subscription_url: string;
  payload: SubscriptionPayload;
}

export async function getUserSubscription(id: number): Promise<SubscriptionResp> {
  const { data } = await apiClient.get<SubscriptionResp>(`/users/${id}/subscription`);
  return data;
}
