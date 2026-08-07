import React from "react";
import { GovernanceTelemetry } from "../engine/GovernanceTelemetry";

export default function AgentGovernanceHistory({ agentId }) {
  const events = GovernanceTelemetry.events.filter(
    evt => evt.agentId === agentId
  );

  if (!events.length) return <div style={{ color: 'var(--text-secondary)', padding: '0.5rem 0' }}>No governance actions recorded.</div>;

  return (
    <div className="agent-governance-history" style={{ marginTop: '0.5rem', display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
      {events.map(evt => (
        <div key={evt.id} className="agent-gov-row">
          <div style={{ color: evt.type.includes('QUARANTINE') ? 'var(--color-red)' : 'var(--color-yellow)', fontWeight: 600 }}>{evt.type}</div>
          <div style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>{evt.reason}</div>
          <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)', fontSize: '0.75rem', marginTop: '2px' }}>
            <span>{new Date(evt.timestamp).toLocaleString()}</span>
            <span>Rule: {evt.ruleId}</span>
          </div>
        </div>
      ))}
    </div>
  );
}
