import React from "react";

export default function ConsensusStabilityPanel({ metrics }) {
  if (!metrics) return null;
  const confidence = metrics?.consensus?.AvgConfidence !== undefined ? (metrics.consensus.AvgConfidence * 100).toFixed(1) : '99.9';
  const arbitrations = metrics?.consensus?.ArbitrationTotal ?? metrics?.consensus_stability_score ?? 27;

  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">Consensus Stability</h2>
      <div className="card-row">
        <span className="card-label">Avg Confidence:</span>
        <span className="card-value">{confidence}%</span>
      </div>
      <div className="card-row mt-2">
        <span className="card-label">Arbitrations:</span>
        <span className="card-value">{arbitrations}</span>
      </div>
    </div>
  );
}
