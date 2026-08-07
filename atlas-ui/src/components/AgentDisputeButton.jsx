import React, { useState } from 'react';
import { AlertCircle } from 'lucide-react';

export default function AgentDisputeButton({ agentId, lastGovernanceEvent }) {
  const [isDisputing, setIsDisputing] = useState(false);
  const [reason, setReason] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!reason.trim()) return;

    // Send dispute to the Arbitration Engine (via telemetry mock for now)
    try {
      await fetch('/api/governance/telemetry', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: 'AGENT_DISPUTE',
          agentId: agentId,
          reason: reason,
          actionRef: lastGovernanceEvent ? (lastGovernanceEvent.event_id || lastGovernanceEvent.id) : null,
          timestamp: Date.now()
        })
      });
      setIsDisputing(false);
      setReason('');
      alert(`Dispute filed for ${agentId}. The Sovereign Arbitration Engine will evaluate.`);
    } catch (err) {
      console.error('Failed to file dispute:', err);
    }
  };

  if (!lastGovernanceEvent) return null; // Only show if there's something to dispute

  return (
    <div style={{ position: 'relative' }}>
      <button 
        onClick={() => setIsDisputing(!isDisputing)}
        style={{ 
          background: 'transparent', 
          color: 'var(--color-yellow)', 
          border: '1px solid var(--color-yellow)', 
          padding: '0.2rem 0.5rem', 
          borderRadius: '4px', 
          cursor: 'pointer', 
          fontSize: '0.75rem',
          display: 'flex',
          alignItems: 'center',
          gap: '0.25rem'
        }}
        title="Raise Dispute"
      >
        <AlertCircle size={12} /> Dispute Action
      </button>

      {isDisputing && (
        <div style={{
          position: 'absolute',
          top: '110%',
          right: 0,
          background: 'var(--bg-0)',
          border: '1px solid var(--border)',
          borderRadius: '4px',
          padding: '0.5rem',
          zIndex: 10,
          width: '250px',
          boxShadow: '0 4px 6px rgba(0,0,0,0.3)'
        }}>
          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            <label style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
              Reason for dispute (Agent {agentId}):
            </label>
            <input 
              type="text" 
              value={reason} 
              onChange={e => setReason(e.target.value)} 
              placeholder="e.g. False Positive FP_032"
              style={{ padding: '0.25rem', fontSize: '0.75rem', background: 'rgba(255,255,255,0.05)', color: '#fff', border: '1px solid var(--border)' }}
            />
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.5rem' }}>
              <button type="button" onClick={() => setIsDisputing(false)} style={{ background: 'transparent', color: 'var(--text-secondary)', border: 'none', cursor: 'pointer', fontSize: '0.75rem' }}>Cancel</button>
              <button type="submit" style={{ background: 'var(--color-yellow)', color: '#000', border: 'none', borderRadius: '2px', cursor: 'pointer', fontSize: '0.75rem', padding: '0.1rem 0.4rem' }}>Submit</button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
