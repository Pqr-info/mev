import React from "react";

export default function NodeHealthPanel({ metrics }) {
  if (!metrics) return null;
  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">Node Health</h2>
      <div className="card-row">
        <span className="card-label">Reachability Failures:</span>
        <span className="card-value" style={{ color: metrics.health.NodeReachabilityFailures > 0 ? 'var(--color-red)' : 'inherit' }}>
          {metrics.health.NodeReachabilityFailures}
        </span>
      </div>
      <div className="card-row mt-2">
        <span className="card-label">Manifest Freshness:</span>
        <span className="card-value">{metrics.health.ManifestFreshnessMs} ms</span>
      </div>
      <div className="card-row mt-2">
        <span className="card-label">StateHash Churn:</span>
        <span className="card-value">{metrics.health.StateHashChurnTotal}</span>
      </div>
    </div>
  );
}
