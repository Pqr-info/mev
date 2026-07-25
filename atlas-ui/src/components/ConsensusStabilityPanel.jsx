import React from "react";

export default function ConsensusStabilityPanel({ metrics }) {
  if (!metrics) return null;
  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">Consensus Stability</h2>
      <div className="card-row">
        <span className="card-label">Avg Confidence:</span>
        <span className="card-value">{(metrics.consensus.AvgConfidence * 100).toFixed(1)}%</span>
      </div>
      <div className="card-row mt-2">
        <span className="card-label">Arbitrations:</span>
        <span className="card-value">{metrics.consensus.ArbitrationTotal}</span>
      </div>
    </div>
  );
}
