/**
 * MeshArtifactClient — Sovereign-27 Artifact Introspection & Governance Client
 */

export class MeshArtifactClient {
  constructor(apiBase = 'http://localhost:4000') {
    this.apiBase = apiBase;
  }

  async getArtifactRegistry() {
    try {
      const res = await fetch(`${this.apiBase}/api/gmi/artifacts/registry`);
      return await res.json();
    } catch (err) {
      console.error('[MeshArtifactClient] Registry Error:', err);
      return { ok: false, error: err.message };
    }
  }

  async resolveArtifact(artifactId) {
    try {
      const res = await fetch(`${this.apiBase}/api/gmi/artifacts/resolve`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ artifactId })
      });
      return await res.json();
    } catch (err) {
      console.error('[MeshArtifactClient] Resolve Error:', err);
      return { ok: false, error: err.message };
    }
  }

  async getArtifactAuditLogs() {
    try {
      const res = await fetch(`${this.apiBase}/api/gmi/artifacts/audit`);
      return await res.json();
    } catch (err) {
      console.error('[MeshArtifactClient] Audit Error:', err);
      return { ok: false, error: err.message };
    }
  }
}
