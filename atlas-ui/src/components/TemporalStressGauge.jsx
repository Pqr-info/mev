import React from "react";

export default function TemporalStressGauge({ metrics }) {
  if (!metrics) return null;
  const stress =
    (metrics?.temporal?.WALMutationsTotal || 0) +
    (metrics?.temporal?.DriftEventsTotal || 0);

  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">Temporal Stress</h2>
      <div style={{ background: 'rgba(255,255,255,0.1)', height: '10px', borderRadius: '5px', overflow: 'hidden' }}>
        <div style={{ background: 'var(--color-red)', width: `${stress % 100}%`, height: '100%', transition: 'width 0.3s' }} />
      </div>
      <p className="mt-2 text-sm card-label">Stress Index: {stress}</p>
    </div>
  );
}
