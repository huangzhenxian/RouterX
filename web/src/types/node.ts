export interface Node {
  id: number;
  name: string;
  ip: string;
  region: string;
  status: number;
  cpu: number;
  memory: number;
  bandwidth: number;
  version: string;
  last_seen: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateNodeInput {
  name: string;
  ip?: string;
  region?: string;
}

export interface CreateNodeResult {
  node: Node;
  auth_token: string;
}
