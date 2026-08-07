export const GovernanceRules = {
  rules: new Map(),

  register(rule) {
    this.rules.set(rule.id, rule);
  },

  getAll() {
    return Array.from(this.rules.values());
  },

  getById(id) {
    return this.rules.get(id) || null;
  },

  update(id, patch) {
    const existing = this.rules.get(id);
    if (!existing) return null;
    const updated = { ...existing, ...patch };
    this.rules.set(id, updated);
    return updated;
  },

  setEnabled(id, enabled) {
    const existing = this.rules.get(id);
    if (!existing) return null;
    existing.enabled = enabled;
    this.rules.set(id, existing);
    return existing;
  }
};

// Seed initial rules based on Phase 15 logic
GovernanceRules.register({
  id: "quarantine-high-confidence",
  scope: "agent",
  condition: "fc.confidence > 0.95",
  action: "quarantine(fc.agent_id, fc.reasoning, fc.forecast_id)",
  constraints: { reversible: true, observable: true, logged: true },
  enabled: true,
  provenance: "system",
  versions: []
});

GovernanceRules.register({
  id: "tighten-medium-confidence",
  scope: "agent",
  condition: "fc.confidence > 0.8 && fc.confidence <= 0.95",
  action: "tighten(fc.agent_id, fc.reasoning, fc.forecast_id)",
  constraints: { reversible: true, observable: true, logged: true },
  enabled: true,
  provenance: "system",
  versions: []
});
