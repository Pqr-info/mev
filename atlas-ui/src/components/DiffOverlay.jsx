import React from 'react';
import './DiffOverlay.css'; // I will add the CSS in index.css, or DiffOverlay.css

export default function DiffOverlay({ markers }) {
  if (!markers || markers.length === 0) return null;

  return (
    <div className="diff-overlay" style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, pointerEvents: 'none' }}>
      {markers.map((marker, idx) => (
        <div
          key={idx}
          className={`diff-marker ${marker.type} severity-${marker.severity}`}
          style={{
            position: 'absolute',
            left: `${marker.x}%`,
            top: `${marker.y}%`,
            width: '20px',
            height: '20px',
            borderRadius: '50%',
            transform: 'translate(-50%, -50%)',
            pointerEvents: 'auto'
          }}
          title={`${marker.type} (${marker.severity})`}
        />
      ))}
    </div>
  );
}
