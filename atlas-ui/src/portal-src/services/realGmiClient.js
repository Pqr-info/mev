// src/services/realGmiClient.js
const BASE = '';

async function call(path, options = {}) {
  let url = `${BASE}${path}`;
  let res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    }
  });

  const text = await res.text();
  let json;
  try { json = JSON.parse(text); } catch {
    json = { raw: text };
  }

  if (!res.ok) {
    const err = new Error(`HTTP ${res.status} ${res.statusText}`);
    err.response = json;
    throw err;
  }
  return json;
}

export async function health() {
  return call('/api/health', { method: 'GET' });
}

export async function fetchTelemetryInspector() {
  return call('/api/telemetry/inspector', { method: 'GET' });
}

export async function registerAgent(payload) {
  return call('/api/gmi/register', {
    method: 'POST',
    body: JSON.stringify(payload || {})
  });
}

export async function bindSubstrate(payload) {
  return call('/api/gmi/bindSubstrate', {
    method: 'POST',
    body: JSON.stringify(payload || {})
  });
}

export async function savePage(payload) {
  return call('/api/gmi/savePage', {
    method: 'POST',
    body: JSON.stringify(payload || {})
  });
}

export async function ensureTicket(payload) {
  return call('/api/gmi/ensureTicket', {
    method: 'POST',
    body: JSON.stringify(payload || {})
  });
}

export async function mapPageToTickets(payload) {
  return call('/api/gmi/mapPageToTickets', {
    method: 'POST',
    body: JSON.stringify(payload || {})
  });
}

export async function buildAgentCube(payload) {
  return call('/api/gmi/buildAgentCube', {
    method: 'POST',
    body: JSON.stringify(payload || {})
  });
}

export async function searchMemory(params) {
  const q = new URLSearchParams(params).toString();
  return call(`/api/gmi/searchMemory?${q}`, { method: 'GET' });
}

export async function pqliteQuery(sql, params = []) {
  return call('/api/pqlite/query', {
    method: 'POST',
    body: JSON.stringify({ sql, params })
  });
}

export async function ingestFilesystem(rootPath, agentId) {
  return call('/api/gmi/ingestFilesystem', {
    method: 'POST',
    body: JSON.stringify({ rootPath, agentId })
  });
}

export const realGmiClient = {
  health,
  fetchTelemetryInspector,
  registerAgent,
  bindSubstrate,
  savePage,
  ensureTicket,
  mapPageToTickets,
  buildAgentCube,
  searchMemory,
  pqliteQuery,
  ingestFilesystem
};
