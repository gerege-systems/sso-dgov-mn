// API Gateway-ийн BFF (/api/gateway/*) хариунуудын TypeScript хэлбэрүүд. Эдгээр
// нь backend-ийн responses_gateway.go-той тэнцүү (json tag-аар).

export interface GwService {
  id: string;
  name: string;
  protocol: string;
  host: string;
  port: number;
  path: string;
  retries: number;
  connect_timeout_ms: number;
  tags: string[];
  enabled: boolean;
  created_at: string;
  updated_at: string | null;
}

export interface GwRoute {
  id: string;
  service_id: string;
  service_name: string;
  name: string;
  methods: string[];
  paths: string[];
  strip_path: boolean;
  preserve_host: boolean;
  enabled: boolean;
  created_at: string;
  updated_at: string | null;
}

export interface GwConsumer {
  id: string;
  username: string;
  custom_id: string;
  tags: string[];
  enabled: boolean;
  key_count: number;
  created_at: string;
  updated_at: string | null;
}

export interface GwKey {
  id: string;
  consumer_id: string;
  label: string;
  prefix: string;
  plaintext?: string;
  last_used_at: string | null;
  expires_at: string | null;
  revoked: boolean;
  created_at: string;
}

export type GwPolicyType = 'rate-limit' | 'key-auth' | 'cors' | 'ip-restriction' | 'request-transform';

export interface GwPolicy {
  id: string;
  route_id: string | null;
  route_name: string;
  type: GwPolicyType;
  config: unknown;
  enabled: boolean;
  created_at: string;
  updated_at: string | null;
}

export interface GwLog {
  id: string;
  route_name: string;
  consumer: string;
  method: string;
  path: string;
  status: number;
  latency_ms: number;
  client_ip: string;
  created_at: string;
}

export interface GwOverview {
  services: number;
  routes: number;
  consumers: number;
  active_keys: number;
  requests_24h: number;
  errors_24h: number;
  rate_limited_24h: number;
  error_rate: number;
  avg_latency_ms: number;
  p95_latency_ms: number;
  status_buckets: { class: string; count: number }[];
  top_routes: { route_name: string; count: number }[];
}
