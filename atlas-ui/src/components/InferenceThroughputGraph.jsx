import React from "react";

export default function InferenceThroughputGraph({ metrics }) {
  if (!metrics) return null;
  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">Inference Throughput</h2>
      <div className="card-row">
        <span className="card-label">Tokens/sec:</span>
        <span className="card-value">{metrics.inference.TokensTotal}</span>
      </div>
      <div className="card-row mt-2">
        <span className="card-label">Avg Latency:</span>
        <span className="card-value">{metrics.inference.AvgLatencyMs} ms</span>
      </div>
    </div>
  );
}
