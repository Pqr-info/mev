export class ArbitrationApplication {
  constructor(engine, telemetryEmitter) {
    this.engine = engine;
    this.telemetryEmitter = telemetryEmitter;
  }

  applyJudgment(disputeId, contextTelemetry = []) {
    const judgment = this.engine.evaluate(disputeId, contextTelemetry);
    
    // Broadcast the sovereign judgment to the mesh
    if (this.telemetryEmitter) {
      this.telemetryEmitter.emit('GOV_ARBITRATION_JUDGMENT', {
        judgmentId: judgment.id,
        disputeId: judgment.disputeId,
        outcome: judgment.outcome,
        rationale: judgment.rationale
      });
    }
    
    return judgment;
  }
}
