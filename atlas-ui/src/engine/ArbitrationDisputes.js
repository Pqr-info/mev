export class ArbitrationDisputes {
  constructor() {
    this.disputes = new Map();
  }
  
  create(agentId, actionRef, reason, payload = {}) {
    const id = `disp_${Date.now()}_${Math.random().toString(36).substr(2,9)}`;
    const dispute = {
      id,
      agentId,
      actionRef,
      reason,
      payload,
      status: 'OPEN', // OPEN, EVALUATING, CLOSED
      timestamp: Date.now()
    };
    this.disputes.set(id, dispute);
    return dispute;
  }
  
  getById(id) {
    return this.disputes.get(id);
  }
  
  getAll() {
    return Array.from(this.disputes.values()).sort((a,b) => b.timestamp - a.timestamp);
  }
  
  update(id, updates) {
    if (!this.disputes.has(id)) return null;
    const existing = this.disputes.get(id);
    const updated = { ...existing, ...updates };
    this.disputes.set(id, updated);
    return updated;
  }
}
