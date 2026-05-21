export interface User {
  id: number;
  username: string;
  uuid: string;
  status: number; // 1=enabled, 0=disabled
  traffic_limit: number;
  used_traffic: number;
  expire_time: string;
  created_at: string;
  updated_at: string;
}

export interface Admin {
  id: number;
  username: string;
  role: string;
  created_at: string;
}

export interface LoginResult {
  token: string;
  expires_at: string;
  admin: Admin;
}

export interface CreateUserInput {
  username: string;
  traffic_limit?: number;
  expire_time?: string;
}
