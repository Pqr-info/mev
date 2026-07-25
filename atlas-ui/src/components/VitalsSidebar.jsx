import React from "react";

export default function VitalsSidebar({ metrics }) {
  if (!metrics) return null;
  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">Vitals</h2>

      <div className="card-row mb-2">
        <span className="card-label">Inference Throughput:</span>
        <span>{metrics.inference.TokensTotal} tokens</span>
      </div>

      <div className="card-row mb-2">
        <span className="card-label">Drift Frequency:</span>
        <span>{metrics.temporal.DriftEventsTotal}</span>
      </div>

      <div className="card-row mb-2">
        <span className="card-label">WAL Mutation Rate:</span>
        <span>{metrics.temporal.WALMutationsTotal}</span>
      </div>

      <div className="card-row mb-2">
        <span className="card-label">Consensus Latency:</span>
        <span>{metrics.consensus.AvgDecisionLatencyMs} ms</span>
      </div>

      <div className="card-row mb-2">
        <span className="card-label">Arbitration Load:</span>
        <span>{metrics.consensus.ArbitrationTotal}</span>
      </div>
    </div>
  );
}
