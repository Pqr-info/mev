/**
 * MeshMemoryClient — Sovereign-27 Eidetic Memory & Queued Recall Client
 */

export class MeshMemoryClient {
  constructor(apiBase = 'http://localhost:4000') {
    this.apiBase = apiBase;
  }

  async storeMemory(memoryId, content, interactionQuality = 0.9, contextAlignment = 0.85, userAffinityScore = 0.95) {
    try {
      const res = await fetch(`${this.apiBase}/api/gmi/memory/store`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ memoryId, content, interactionQuality, contextAlignment, userAffinityScore })
      });
      return await res.json();
    } catch (err) {
      console.error('[MeshMemoryClient] Store Error:', err);
      return { ok: false, error: err.message };
    }
  }

  async recallMemories(queryContext = 'architecture', topK = 3) {
    try {
      const res = await fetch(`${this.apiBase}/api/gmi/memory/recall`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ queryContext, topK })
      });
      return await res.json();
    } catch (err) {
      console.error('[MeshMemoryClient] Recall Error:', err);
      return { ok: false, error: err.message };
    }
  }

  async getMemoryAuditLogs() {
    try {
      const res = await fetch(`${this.apiBase}/api/gmi/memory/audit`);
      return await res.json();
    } catch (err) {
      console.error('[MeshMemoryClient] Audit Error:', err);
      return { ok: false, error: err.message };
    }
  }
}
