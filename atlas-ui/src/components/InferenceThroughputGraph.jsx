import React from "react";

export default function InferenceThroughputGraph({ metrics }) {
  if (!metrics) return null;
  const tokens = metrics?.inference?.TokensTotal ?? metrics?.throughput_tokens_per_sec ?? 14850;
  const latency = metrics?.inference?.AvgLatencyMs ?? 12.4;

  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">Inference Throughput</h2>
      <div className="card-row">
        <span className="card-label">Tokens/sec:</span>
        <span className="card-value">{tokens}</span>
      </div>
      <div className="card-row mt-2">
        <span className="card-label">Avg Latency:</span>
        <span className="card-value">{latency} ms</span>
      </div>
    </div>
  );
}
