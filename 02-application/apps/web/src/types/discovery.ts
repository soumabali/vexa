export interface PortResult {
  port: number;
  service?: string;
  banner?: string;
  protocol: 'tcp' | 'udp';
  tls?: boolean;
}

export interface ScanResult {
  id: string;
  ip_address: string;
  hostname?: string;
  open_ports: PortResult[];
  os_guess?: string;
  response_time_ms: number;
  status: 'up' | 'down' | 'filtered';
  scanned_at: string;
}

export interface ScanOptions {
  timeout_ms: number;
  rate_limit_ms: number;
  concurrency: number;
  icmp_first: boolean;
  os_scan: boolean;
  resolve_hostname: boolean;
}

export interface ScanJob {
  id: string;
  network: string;
  ports: number[];
  status: 'pending' | 'running' | 'completed' | 'cancelled' | 'failed';
  progress: number;
  results: ScanResult[];
  total_hosts: number;
  scanned_hosts: number;
  found_hosts?: number;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  error?: string;
  options: ScanOptions;
}

export interface StartScanRequest {
  network: string;
  ports: number[];
  options: Partial<ScanOptions>;
}

export interface StartScanResponse {
  id: string;
  status: string;
  network: string;
  message: string;
  created_at: string;
}

export interface ScanResultsResponse {
  scan: ScanJob;
  results: ScanResult[];
}

export interface ScanHistoryResponse {
  jobs: ScanJob[];
}

export interface ImportResultsRequest {
  import_all: boolean;
  selected_ips?: string[];
}

export interface ImportResultsResponse {
  imported: number;
  imported_count?: number;
  failed: number;
  host_ids: string[];
}
