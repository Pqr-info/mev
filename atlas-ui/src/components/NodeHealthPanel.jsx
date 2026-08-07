import React from "react";

export default function NodeHealthPanel({ metrics }) {
  if (!metrics) return null;
  const reachabilityFailures = metrics?.health?.NodeReachabilityFailures ?? 0;
  const manifestFreshness = metrics?.health?.ManifestFreshnessMs ?? 1.2;
  const stateHashChurn = metrics?.health?.StateHashChurnTotal ?? 0;

  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">Node Health</h2>
      <div className="card-row">
        <span className="card-label">Reachability Failures:</span>
        <span className="card-value" style={{ color: reachabilityFailures > 0 ? 'var(--color-red)' : 'inherit' }}>
          {reachabilityFailures}
        </span>
      </div>
      <div className="card-row mt-2">
        <span className="card-label">Manifest Freshness:</span>
        <span className="card-value">{manifestFreshness} ms</span>
      </div>
      <div className="card-row mt-2">
        <span className="card-label">StateHash Churn:</span>
        <span className="card-value">{stateHashChurn}</span>
      </div>
    </div>
  );
}
