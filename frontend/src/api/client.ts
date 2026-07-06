const basePath = window.location.pathname.replace(/\/$/, '');
const API_BASE = `${basePath}/api/traffilk`;

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (res.status === 401) {
    window.location.hash = '#/login';
    throw new Error('Unauthorized');
  }

  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(data.error || `HTTP ${res.status}`);
  }

  return res.json();
}

export interface Node {
  id: number;
  name: string;
  url: string;
  status: string;
  trafficUsedBytes: number;
  trafficLimitBytes: number;
  isTrafficTrackingActive: boolean;
  trafficResetDay: number;
  rxBytesPerSec: number;
  txBytesPerSec: number;
  cpuLoadPercent: number;
  loadAvg1: number;
  loadAvg5: number;
  loadAvg15: number;
  memTotalBytes: number;
  memUsedBytes: number;
  uptimeSeconds: number;
  netDropsRx: number;
  netDropsTx: number;
  fileDescriptors: number;
}

export interface DailyTraffic {
  date: string;
  rx_bytes: number;
  tx_bytes: number;
}

export interface APIToken {
  id: number;
  name: string;
  created_at: string;
}

// Auth
export const login = (username: string, password: string) =>
  request<{ status: string; token: string }>('/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });

export const logout = () =>
  request<{ status: string }>('/logout', { method: 'POST' });

// Nodes
export const getNodes = () => request<Node[]>('/nodes');
export const addNode = (node: { name: string; url: string }) =>
  request<{ status: string }>('/nodes', { method: 'POST', body: JSON.stringify(node) });
export const updateNode = (id: number, node: { name: string; url: string }) =>
  request<{ status: string }>(`/nodes/${id}`, { method: 'PUT', body: JSON.stringify(node) });
export const deleteNode = (id: number) =>
  request<{ status: string }>(`/nodes/${id}`, { method: 'DELETE' });
export const getNodeTraffic = (id: number) =>
  request<DailyTraffic[]>(`/nodes/${id}/traffic`);
export const pollNode = (id: number) =>
  request<{ status: string }>(`/nodes/${id}/poll`, { method: 'POST' });

// Tokens
export const getTokens = () => request<APIToken[]>('/tokens');
export const createToken = (name: string) =>
  request<{ token: string }>('/tokens', { method: 'POST', body: JSON.stringify({ name }) });
export const deleteToken = (id: number) =>
  request<{ status: string }>(`/tokens/${id}`, { method: 'DELETE' });
