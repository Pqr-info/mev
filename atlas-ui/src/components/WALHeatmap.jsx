import React from "react";

export default function WALHeatmap({ metrics }) {
  if (!metrics) return null;
  const density = (metrics?.temporal?.WALMutationsTotal || 14) % 50;

  return (
    <div className="glass-card">
      <h2 className="card-title mb-4">WAL Heatmap</h2>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(10, 1fr)', gap: '4px' }}>
        {[...Array(50)].map((_, i) => (
          <div
            key={i}
            style={{ 
              height: '15px', 
              background: 'var(--color-blue)', 
              borderRadius: '2px',
              opacity: i < density ? 1 : 0.1 
            }}
          />
        ))}
      </div>
    </div>
  );
}
