export interface Provider {
  id: number;
  name: string;
  type: 'socks5' | 'http' | 'https';
  host: string;
  port: number;
  username: string;
  region: string;
  tags: string;
  priority: number;
  status: number; // 1=enabled, 0=disabled
  healthy: boolean;
  latency_ms: number;
  fail_count: number;
  last_checked_at: string | null;
  last_error: string;
  created_at: string;
  updated_at: string;
}

export interface CreateProviderInput {
  name: string;
  type: 'socks5' | 'http' | 'https';
  host: string;
  port: number;
  username?: string;
  password?: string;
  region?: string;
  tags?: string;
  priority?: number;
}

export interface TestResult {
  healthy: boolean;
  latency_ms: number;
  error: string;
}
