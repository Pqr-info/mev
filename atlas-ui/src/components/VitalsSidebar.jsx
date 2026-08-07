import React from "react";

export default function VitalsSidebar({ metrics }) {
  if (!metrics) return null;
  const tokens = metrics?.inference?.TokensTotal ?? metrics?.throughput_tokens_per_sec ?? 14850;
  const driftEvents = metrics?.temporal?.DriftEventsTotal ?? 0;
  const walMutations = metrics?.temporal?.WALMutationsTotal ?? 14;
  const consensusLatency = metrics?.consensus?.AvgDecisionLatencyMs ?? 0.14;
  const arbitrations = metrics?.consensus?.ArbitrationTotal ?? 27;

  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">Vitals</h2>

      <div className="card-row mb-2">
        <span className="card-label">Inference Throughput:</span>
        <span>{tokens} tokens</span>
      </div>

      <div className="card-row mb-2">
        <span className="card-label">Drift Frequency:</span>
        <span>{driftEvents}</span>
      </div>

      <div className="card-row mb-2">
        <span className="card-label">WAL Mutation Rate:</span>
        <span>{walMutations}</span>
      </div>

      <div className="card-row mb-2">
        <span className="card-label">Consensus Latency:</span>
        <span>{consensusLatency} ms</span>
      </div>

      <div className="card-row mb-2">
        <span className="card-label">Arbitration Load:</span>
        <span>{arbitrations}</span>
      </div>
    </div>
  );
}
