import React, { useState, useEffect } from 'react';
import { Gavel, Clock, CheckCircle, XCircle, AlertTriangle } from 'lucide-react';

export default function ArbitrationPanel({ operatorId, onClose }) {
  const [disputes, setDisputes] = useState([]);
  const [judgments, setJudgments] = useState([]);

  // Mocking the engine integration for the UI since this runs in browser
  useEffect(() => {
    // In a real app, this would poll the backend which runs the ArbitrationEngine
    // For this UI, we listen to telemetry for DISPUTES and JUDGMENTS
    import('../engine/GovernanceTelemetry').then(({ GovernanceTelemetry }) => {
      const unsub = GovernanceTelemetry.subscribe(evt => {
        if (evt.type === 'AGENT_DISPUTE') {
          setDisputes(prev => [{
            id: `disp_${Date.now()}`,
            agentId: evt.agentId,
            actionRef: evt.actionRef,
            reason: evt.reason,
            status: 'OPEN',
            timestamp: evt.timestamp
          }, ...prev]);
        }
        if (evt.type === 'GOV_ARBITRATION_JUDGMENT') {
          setJudgments(prev => [evt, ...prev]);
          setDisputes(prev => prev.map(d => d.id === evt.disputeId ? { ...d, status: 'CLOSED' } : d));
        }
      });
      return () => unsub(); // cleanup mock
    });
  }, []);

  // For Demo purposes, allow the operator to manually resolve an open dispute
  const handleResolve = (disputeId, outcome, rationale) => {
    const evt = {
      type: 'GOV_ARBITRATION_JUDGMENT',
      judgmentId: `judg_${Date.now()}`,
      disputeId,
      outcome,
      rationale,
      timestamp: Date.now()
    };
    
    // Dispatch to our local mock receiver
    import('../engine/GovernanceTelemetry').then(({ GovernanceTelemetry }) => {
       GovernanceTelemetry.emit('GOV_ARBITRATION_JUDGMENT', evt);
    });
  };

  const getOutcomeColor = (outcome) => {
    switch (outcome) {
      case 'UPHOLD': return 'var(--color-green)';
      case 'MODIFY': return 'var(--color-yellow)';
      case 'OVERTURN': return 'var(--color-blue)';
      case 'ESCALATE': return 'var(--color-red)';
      default: return 'var(--text-primary)';
    }
  };

  return (
    <div className="arbitration-panel" style={{ display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }}>
      <div className="panel-header" style={{ padding: '1rem', borderBottom: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0, fontSize: '1.2rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--color-purple)' }}>
          <Gavel size={20} /> Sovereign Arbitration
        </h2>
        <button onClick={onClose} style={{ background: 'transparent', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer' }}>✕</button>
      </div>

      <div style={{ flex: 1, overflowY: 'auto', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        
        <section>
          <h3 style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
            <Clock size={14} /> Open Disputes
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            {disputes.filter(d => d.status === 'OPEN').length === 0 && (
              <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontStyle: 'italic' }}>No active disputes.</div>
            )}
            {disputes.filter(d => d.status === 'OPEN').map(d => (
              <div key={d.id} className="dispute-row" style={{ background: 'rgba(255,255,255,0.02)', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.75rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.25rem' }}>
                  <strong style={{ color: 'var(--color-yellow)' }}>{d.agentId}</strong>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>Ref: {d.actionRef?.substring(0,8)}</span>
                </div>
                <div style={{ fontSize: '0.85rem', marginBottom: '0.75rem' }}>"{d.reason}"</div>
                
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button onClick={() => handleResolve(d.id, 'UPHOLD', 'Rule constraint correctly applied.')} style={{ flex: 1, padding: '0.25rem', background: 'rgba(34, 197, 94, 0.1)', color: 'var(--color-green)', border: '1px solid var(--color-green)', borderRadius: '2px', cursor: 'pointer', fontSize: '0.75rem' }}>Uphold</button>
                  <button onClick={() => handleResolve(d.id, 'OVERTURN', 'False positive confirmed.')} style={{ flex: 1, padding: '0.25rem', background: 'rgba(59, 130, 246, 0.1)', color: 'var(--color-blue)', border: '1px solid var(--color-blue)', borderRadius: '2px', cursor: 'pointer', fontSize: '0.75rem' }}>Overturn</button>
                  <button onClick={() => handleResolve(d.id, 'ESCALATE', 'Requires human orchestrator review.')} style={{ flex: 1, padding: '0.25rem', background: 'rgba(239, 68, 68, 0.1)', color: 'var(--color-red)', border: '1px dashed var(--color-red)', borderRadius: '2px', cursor: 'pointer', fontSize: '0.75rem' }}>Escalate</button>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section>
          <h3 style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
            <Gavel size={14} /> Recent Judgments
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            {judgments.length === 0 && (
              <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontStyle: 'italic' }}>No recent judgments.</div>
            )}
            {judgments.map(j => (
              <div key={j.judgmentId} className="judgment-row" style={{ background: 'rgba(255,255,255,0.02)', borderLeft: `3px solid ${getOutcomeColor(j.outcome)}`, borderRadius: '0 4px 4px 0', padding: '0.75rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.25rem' }}>
                  <strong style={{ color: getOutcomeColor(j.outcome) }}>{j.outcome}</strong>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>Disp: {j.disputeId?.substring(0,8)}</span>
                </div>
                <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>{j.rationale}</div>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  );
}
