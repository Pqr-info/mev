/**
 * Mesh-Wide Skill Synchronization Client (KV-Backed)
 * Resolves content-addressed, versioned skills & capability graphs across the 5D Mesh.
 */

export class MeshSkillClient {
  constructor(apiBase = 'http://localhost:4000') {
    this.apiBase = apiBase;
    this.registry = {};
    this.capabilities = new Set();
  }

  /**
   * Fetch active skill manifests from mesh KV
   */
  async bootSequence() {
    try {
      const resp = await fetch(`${this.apiBase}/api/gmi/skills/registry`);
      const data = await resp.json();
      if (data.ok) {
        this.registry = data.registry;
        this.capabilities = new Set(data.activeCapabilities);
      }
      return data;
    } catch (e) {
      console.warn('[MeshSkillClient Boot Failed]', e.message);
      return { ok: false, error: e.message };
    }
  }

  /**
   * Resolve an agent capability against Council governance & manifest rules
   */
  async resolveCapability(agent = 'max', need = 'midi.snapshot.write') {
    try {
      const resp = await fetch(`${this.apiBase}/api/gmi/skills/resolve`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agent, need })
      });
      return await resp.json();
    } catch (e) {
      return { ok: false, error: e.message };
    }
  }
}

export const meshSkillClient = new MeshSkillClient();
