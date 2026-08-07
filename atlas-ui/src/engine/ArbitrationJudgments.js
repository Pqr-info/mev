export class ArbitrationJudgments {
  constructor() {
    this.judgments = new Map();
  }

  record(disputeId, outcome, rationale) {
    const id = `judg_${Date.now()}`;
    const judgment = {
      id,
      disputeId,
      outcome, // UPHOLD, MODIFY, OVERTURN, ESCALATE
      rationale,
      timestamp: Date.now()
    };
    this.judgments.set(id, judgment);
    return judgment;
  }

  getById(id) {
    return this.judgments.get(id);
  }

  getAll() {
    return Array.from(this.judgments.values()).sort((a,b) => b.timestamp - a.timestamp);
  }
}
