import React, { useEffect, useState } from "react";
import { GovernanceTelemetry } from "../engine/GovernanceTelemetry";
import { computeGovernanceMetrics } from "../engine/GovernanceMetrics";

export default function GovernanceDashboard({ onClose }) {
  const [events, setEvents] = useState([...GovernanceTelemetry.events]);
  
  useEffect(() => {
    const unsubscribe = GovernanceTelemetry.subscribe(evt => {
      setEvents([...GovernanceTelemetry.events]);
    });
    return unsubscribe;
  }, []);

  const metrics = computeGovernanceMetrics(events);

  return (
    <div className="governance-dashboard">
      <div className="dashboard-header">
        <span>Governance Oversight</span>
        <button onClick={onClose} style={{ background: 'transparent', border: 'none', color: '#fff', cursor: 'pointer', fontSize: '1.2rem' }}>×</button>
      </div>

      <div className="dashboard-section">
        <div>Total actions: {metrics.total}</div>
      </div>

      <div className="dashboard-section">
        <div className="section-title" style={{ fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '4px' }}>By Type</div>
        {Object.entries(metrics.byType).length === 0 && <div style={{ color: 'var(--text-secondary)' }}>No actions yet.</div>}
        {Object.entries(metrics.byType).map(([type, count]) => (
          <div key={type} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
            <span>{type}</span>
            <span>{count}</span>
          </div>
        ))}
      </div>

      <div className="dashboard-section">
        <div className="section-title" style={{ fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '4px' }}>Recent Events</div>
        <div className="events-list">
          {events.slice(-10).reverse().map(evt => (
            <div key={evt.id} className="event-row">
              <div style={{ color: evt.type.includes('QUARANTINE') ? 'var(--color-red)' : 'var(--color-yellow)' }}>{evt.type}</div>
              <div style={{ fontWeight: 600 }}>{evt.agentId}</div>
              <div style={{ color: 'var(--text-secondary)' }}>{new Date(evt.timestamp).toLocaleTimeString()}</div>
            </div>
          ))}
          {events.length === 0 && <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem' }}>No recent events.</div>}
        </div>
      </div>
    </div>
  );
}
