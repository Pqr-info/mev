import { ArbitrationJudgments } from './ArbitrationJudgments';

export class ArbitrationEngine {
  constructor(disputes, judgments) {
    this.disputes = disputes;
    this.judgments = judgments;
  }

  evaluate(disputeId, contextTelemetry = []) {
    const dispute = this.disputes.getById(disputeId);
    if (!dispute) throw new Error("Dispute not found");
    
    this.disputes.update(disputeId, { status: 'EVALUATING' });
    
    // Engine Logic: Analyze context telemetry against dispute
    let outcome = 'UPHOLD';
    let rationale = 'Default rule constraint applied. Action upheld.';
    
    // Trivial heuristic for MVP: if reason contains "override", escalate
    if (dispute.reason.toLowerCase().includes('override')) {
      outcome = 'ESCALATE';
      rationale = 'Manual override request detected. Escalated to Orchestrator.';
    } else if (contextTelemetry.some(t => t.type === 'FALSE_POSITIVE')) {
      outcome = 'OVERTURN';
      rationale = 'False positive telemetry signature detected. Overturning action.';
    }
    
    const judgment = this.judgments.record(disputeId, outcome, rationale);
    this.disputes.update(disputeId, { status: 'CLOSED' });
    
    return judgment;
  }
}
