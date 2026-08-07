import React, { useState, useEffect } from 'react';
import { Clock, Play, GitCompare, Zap, ArrowLeft } from 'lucide-react';
import { TimeMachineContext } from '../engine/TimeMachineContext';
import { TimeMachineMiddleware } from '../engine/TimeMachineMiddleware';
import { TimeMachineCounterfactual } from '../engine/TimeMachineCounterfactual';
import { GovernanceRules } from '../engine/GovernanceRules';

export default function TimeMachinePanel({ onClose }) {
  const [activeTab, setActiveTab] = useState('epoch'); // 'epoch' | 'diff' | 'counterfactual'
  const [targetDate, setTargetDate] = useState('');
  const [targetTime, setTargetTime] = useState('');
  const [isTemporalActive, setIsTemporalActive] = useState(false);
  const [temporalState, setTemporalState] = useState(null);
  const [checkpointWindow, setCheckpointWindow] = useState(null);

  // Diff state
  const [diffTs1Date, setDiffTs1Date] = useState('');
  const [diffTs1Time, setDiffTs1Time] = useState('');
  const [diffTs2Date, setDiffTs2Date] = useState('');
  const [diffTs2Time, setDiffTs2Time] = useState('');
  const [diffResult, setDiffResult] = useState(null);

  // Counterfactual state
  const [selectedRuleId, setSelectedRuleId] = useState('');
  const [cfTimestamp, setCfTimestamp] = useState('');
  const [cfResult, setCfResult] = useState(null);

  // TSRE state
  const [tsreAgentId, setTsreAgentId] = useState('');
  const [tsreRollbackTs, setTsreRollbackTs] = useState('');
  const [tsreResult, setTsreResult] = useState('');

  useEffect(() => {
    TimeMachineMiddleware.fetchCheckpointWindow().then(setCheckpointWindow);

    const unsub = TimeMachineContext.subscribe(meta => {
      setIsTemporalActive(meta.temporal);
    });
    return unsub;
  }, []);

  // Set default dates to now
  useEffect(() => {
    const now = new Date();
    const dateStr = now.toISOString().split('T')[0];
    const timeStr = now.toTimeString().slice(0, 5);
    if (!targetDate) setTargetDate(dateStr);
    if (!targetTime) setTargetTime(timeStr);
    if (!diffTs2Date) { setDiffTs2Date(dateStr); setDiffTs2Time(timeStr); }
    const hourAgo = new Date(now.getTime() - 3600000);
    if (!diffTs1Date) { setDiffTs1Date(hourAgo.toISOString().split('T')[0]); setDiffTs1Time(hourAgo.toTimeString().slice(0, 5)); }
  }, []);

  const handleEnterTemporal = async () => {
    const ts = new Date(`${targetDate}T${targetTime}`).getTime();
    if (isNaN(ts)) return;
    TimeMachineContext.enterTemporal(ts);
    const state = await TimeMachineMiddleware.fetchTemporalState(ts);
    setTemporalState(state);
  };

  const handleExitTemporal = () => {
    TimeMachineContext.exitTemporal();
    setTemporalState(null);
  };

  const handleDiff = async () => {
    const ts1 = new Date(`${diffTs1Date}T${diffTs1Time}`).getTime();
    const ts2 = new Date(`${diffTs2Date}T${diffTs2Time}`).getTime();
    if (isNaN(ts1) || isNaN(ts2)) return;
    const result = await TimeMachineMiddleware.diffStates(ts1, ts2);
    setDiffResult(result);
  };

  const handleCounterfactual = async () => {
    if (!selectedRuleId || !cfTimestamp) return;
    const ts = new Date(cfTimestamp).getTime();
    if (isNaN(ts)) return;
    const result = await TimeMachineCounterfactual.evaluateRuleAtTime(selectedRuleId, ts);
    setCfResult(result);
  };

  const handleTSRE = async () => {
    if (!tsreAgentId) return;
    const ts = tsreRollbackTs ? new Date(tsreRollbackTs).getTime() : null;
    const success = await TimeMachineContext.initiateSelfRepair(tsreAgentId, ts);
    if (success) {
      setTsreResult(`Successfully initiated TSRE for agent ${tsreAgentId}. L6 commits rolled back and trust restored.`);
    } else {
      setTsreResult(`Failed to initiate TSRE for ${tsreAgentId}.`);
    }
  };

  const rules = GovernanceRules.getAll();

  const tabStyle = (tab) => ({
    flex: 1,
    padding: '0.5rem',
    background: activeTab === tab ? 'rgba(168, 85, 247, 0.2)' : 'transparent',
    color: activeTab === tab ? 'var(--color-purple)' : 'var(--text-secondary)',
    border: 'none',
    borderBottom: activeTab === tab ? '2px solid var(--color-purple)' : '2px solid transparent',
    cursor: 'pointer',
    fontSize: '0.8rem',
    fontWeight: 600
  });

  return (
    <div className="time-machine-panel" style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', borderBottom: '1px solid var(--border)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--color-purple)' }}>
          <Clock size={20} />
          <span style={{ fontWeight: 600 }}>JetWeb Time Machine</span>
          {isTemporalActive && (
            <span className="tag" style={{ background: 'rgba(168, 85, 247, 0.2)', color: 'var(--color-purple)', border: '1px solid var(--color-purple)', fontSize: '0.7rem' }}>TEMPORAL</span>
          )}
        </div>
        <button onClick={onClose} style={{ background: 'transparent', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', fontSize: '1.2rem' }}>×</button>
      </div>

      {/* Tab Bar */}
      <div style={{ display: 'flex', borderBottom: '1px solid var(--border)' }}>
        <button style={tabStyle('epoch')} onClick={() => setActiveTab('epoch')}>
          <Clock size={14} /> Epoch
        </button>
        <button style={tabStyle('diff')} onClick={() => setActiveTab('diff')}>
          <GitCompare size={14} /> Diff
        </button>
        <button style={tabStyle('counterfactual')} onClick={() => setActiveTab('counterfactual')}>
          <Zap size={14} /> What-If
        </button>
        <button style={tabStyle('tsre')} onClick={() => setActiveTab('tsre')}>
          <Zap size={14} /> TSRE
        </button>
      </div>

      {/* Content */}
      <div style={{ flex: 1, padding: '1rem', overflowY: 'auto' }}>

        {/* Epoch Picker Tab */}
        {activeTab === 'epoch' && (
          <div>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '1rem' }}>
              Travel to a past epoch and view the mesh state as it existed at that moment. The Pre-Emptive Auditor operates in <strong>read-only</strong> mode.
            </div>

            {checkpointWindow && (
              <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '1rem', padding: '0.5rem', background: 'var(--bg-1)', borderRadius: '4px', border: '1px solid var(--border)' }}>
                Checkpoint Window: {new Date(checkpointWindow.oldest).toLocaleString()} → {new Date(checkpointWindow.newest).toLocaleString()}
              </div>
            )}

            <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
              <input type="date" value={targetDate} onChange={e => setTargetDate(e.target.value)}
                style={{ flex: 1, background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.5rem', fontSize: '0.85rem' }} />
              <input type="time" value={targetTime} onChange={e => setTargetTime(e.target.value)}
                style={{ flex: 1, background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.5rem', fontSize: '0.85rem' }} />
            </div>

            {!isTemporalActive ? (
              <button onClick={handleEnterTemporal}
                style={{ width: '100%', padding: '0.75rem', background: 'rgba(168, 85, 247, 0.2)', color: 'var(--color-purple)', border: '1px solid var(--color-purple)', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.5rem' }}>
                <Play size={16} /> Enter Temporal Mode
              </button>
            ) : (
              <button onClick={handleExitTemporal}
                style={{ width: '100%', padding: '0.75rem', background: 'rgba(239, 68, 68, 0.2)', color: 'var(--color-red)', border: '1px solid var(--color-red)', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.5rem' }}>
                <ArrowLeft size={16} /> Exit Temporal Mode (Return to LIVE)
              </button>
            )}

            {/* Temporal State Preview */}
            {temporalState && !temporalState.error && (
              <div style={{ marginTop: '1rem', padding: '0.75rem', background: 'var(--bg-1)', borderRadius: '4px', border: '1px dashed var(--color-purple)' }}>
                <div style={{ fontWeight: 600, color: 'var(--color-purple)', marginBottom: '0.5rem', fontSize: '0.85rem' }}>Temporal State @ {new Date(temporalState.meta?.targetTimestamp).toLocaleString()}</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                  Agents: {(temporalState.agents || []).length} | 
                  Proposals: {(temporalState.proposals || []).length} | 
                  Forecasts: {(temporalState.forecasts || []).length}
                </div>
                {(temporalState.agents || []).map(a => (
                  <div key={a.agent_id} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', padding: '0.25rem 0', borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                    <span>{a.agent_id.toUpperCase()}</span>
                    <span style={{ color: a.status === 'QUARANTINE' ? 'var(--color-red)' : a.status === 'PROBATION' ? 'var(--color-yellow)' : 'var(--color-green)' }}>
                      {a.status} ({a.trust_score}/100)
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Diff Tab */}
        {activeTab === 'diff' && (
          <div>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '1rem' }}>
              Compare mesh state between two epochs. See trust score changes, status transitions, and rule deltas.
            </div>

            <div style={{ marginBottom: '0.75rem' }}>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>Epoch A (Before)</div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <input type="date" value={diffTs1Date} onChange={e => setDiffTs1Date(e.target.value)}
                  style={{ flex: 1, background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.4rem', fontSize: '0.8rem' }} />
                <input type="time" value={diffTs1Time} onChange={e => setDiffTs1Time(e.target.value)}
                  style={{ flex: 1, background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.4rem', fontSize: '0.8rem' }} />
              </div>
            </div>

            <div style={{ marginBottom: '0.75rem' }}>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>Epoch B (After)</div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <input type="date" value={diffTs2Date} onChange={e => setDiffTs2Date(e.target.value)}
                  style={{ flex: 1, background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.4rem', fontSize: '0.8rem' }} />
                <input type="time" value={diffTs2Time} onChange={e => setDiffTs2Time(e.target.value)}
                  style={{ flex: 1, background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.4rem', fontSize: '0.8rem' }} />
              </div>
            </div>

            <button onClick={handleDiff}
              style={{ width: '100%', padding: '0.75rem', background: 'rgba(59, 130, 246, 0.2)', color: 'var(--color-blue)', border: '1px solid var(--color-blue)', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.5rem' }}>
              <GitCompare size={16} /> Compute Cross-Epoch Diff
            </button>

            {diffResult && (
              <div className="epoch-diff" style={{ marginTop: '1rem', padding: '0.75rem', background: 'var(--bg-1)', borderRadius: '4px', border: '1px solid var(--border)' }}>
                <div style={{ fontWeight: 600, marginBottom: '0.5rem', fontSize: '0.85rem', color: 'var(--color-blue)' }}>{diffResult.summary}</div>
                {(diffResult.agentDeltas || []).map(d => (
                  <div key={d.agent_id} style={{ padding: '0.5rem', marginBottom: '0.25rem', background: 'rgba(0,0,0,0.2)', borderRadius: '4px', fontSize: '0.8rem', borderLeft: `3px solid ${d.trustDelta > 0 ? 'var(--color-green)' : d.trustDelta < 0 ? 'var(--color-red)' : 'var(--text-secondary)'}` }}>
                    <div style={{ fontWeight: 600 }}>{d.agent_id.toUpperCase()}</div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                      <span>Trust: {d.trustBefore} → {d.trustAfter} ({d.trustDelta > 0 ? '+' : ''}{d.trustDelta})</span>
                      <span>{d.statusBefore} → {d.statusAfter}</span>
                    </div>
                  </div>
                ))}
                {(diffResult.agentDeltas || []).length === 0 && (
                  <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem' }}>No agent changes detected between epochs.</div>
                )}
              </div>
            )}
          </div>
        )}

        {/* Counterfactual Tab */}
        {activeTab === 'counterfactual' && (
          <div>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '1rem' }}>
              Ask "what if" — select a governance rule and a past timestamp to see what the auditor <strong>would have done</strong>.
            </div>

            <div style={{ marginBottom: '0.75rem' }}>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>Rule</div>
              <select value={selectedRuleId} onChange={e => setSelectedRuleId(e.target.value)}
                style={{ width: '100%', background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.5rem', fontSize: '0.85rem' }}>
                <option value="">Select a rule...</option>
                {rules.map(r => <option key={r.id} value={r.id}>{r.id}</option>)}
              </select>
            </div>

            <div style={{ marginBottom: '0.75rem' }}>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>Timestamp</div>
              <input type="datetime-local" value={cfTimestamp} onChange={e => setCfTimestamp(e.target.value)}
                style={{ width: '100%', background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.5rem', fontSize: '0.85rem' }} />
            </div>

            <button onClick={handleCounterfactual} disabled={!selectedRuleId || !cfTimestamp}
              style={{ width: '100%', padding: '0.75rem', background: selectedRuleId && cfTimestamp ? 'rgba(234, 179, 8, 0.2)' : 'var(--bg-1)', color: selectedRuleId && cfTimestamp ? 'var(--color-yellow)' : 'var(--text-secondary)', border: `1px solid ${selectedRuleId && cfTimestamp ? 'var(--color-yellow)' : 'var(--border)'}`, borderRadius: '4px', cursor: selectedRuleId && cfTimestamp ? 'pointer' : 'not-allowed', fontWeight: 600, display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.5rem' }}>
              <Zap size={16} /> Evaluate Counterfactual
            </button>

            {cfResult && (
              <div style={{ marginTop: '1rem', padding: '0.75rem', background: 'var(--bg-1)', borderRadius: '4px', border: '1px dashed var(--color-yellow)' }}>
                <div style={{ fontWeight: 600, color: 'var(--color-yellow)', marginBottom: '0.5rem', fontSize: '0.85rem' }}>{cfResult.summary}</div>
                {(cfResult.results || []).map((r, i) => (
                  <div key={i} style={{ padding: '0.4rem', marginBottom: '0.25rem', background: 'rgba(0,0,0,0.2)', borderRadius: '4px', fontSize: '0.8rem', borderLeft: `3px solid ${r.wouldHaveFired ? 'var(--color-red)' : 'var(--color-green)'}` }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <span style={{ fontWeight: 600 }}>{r.agent_id}</span>
                      <span style={{ color: r.wouldHaveFired ? 'var(--color-red)' : 'var(--color-green)' }}>
                        {r.wouldHaveFired ? 'WOULD FIRE' : 'NO FIRE'}
                      </span>
                    </div>
                    <div style={{ color: 'var(--text-secondary)', fontSize: '0.75rem' }}>
                      Confidence: {((r.confidence || 0) * 100).toFixed(0)}% | {r.type || 'N/A'}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* TSRE Tab */}
        {activeTab === 'tsre' && (
          <div>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '1rem' }}>
              <strong>Temporal Self-Repair Engine (TSRE)</strong> autonomously diagnoses L6 Splinters and applies historical rollback + forward replay to repair corrupted state.
            </div>

            <div style={{ marginBottom: '0.75rem' }}>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>Target Agent (e.g. max, ted)</div>
              <input type="text" value={tsreAgentId} onChange={e => setTsreAgentId(e.target.value)} placeholder="Agent ID"
                style={{ width: '100%', background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.5rem', fontSize: '0.85rem' }} />
            </div>

            <div style={{ marginBottom: '0.75rem' }}>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>Rollback Timestamp (Optional)</div>
              <input type="datetime-local" value={tsreRollbackTs} onChange={e => setTsreRollbackTs(e.target.value)}
                style={{ width: '100%', background: 'rgba(0,0,0,0.3)', color: '#f5f5f7', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.5rem', fontSize: '0.85rem' }} />
            </div>

            <button onClick={handleTSRE} disabled={!tsreAgentId}
              style={{ width: '100%', padding: '0.75rem', background: tsreAgentId ? 'rgba(16, 185, 129, 0.2)' : 'var(--bg-1)', color: tsreAgentId ? 'var(--color-green)' : 'var(--text-secondary)', border: `1px solid ${tsreAgentId ? 'var(--color-green)' : 'var(--border)'}`, borderRadius: '4px', cursor: tsreAgentId ? 'pointer' : 'not-allowed', fontWeight: 600, display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.5rem' }}>
              <Zap size={16} /> Execute TSRE Repair
            </button>

            {tsreResult && (
              <div style={{ marginTop: '1rem', padding: '0.75rem', background: 'var(--bg-1)', borderRadius: '4px', border: '1px dashed var(--color-green)' }}>
                <div style={{ fontSize: '0.85rem', color: 'var(--color-green)' }}>{tsreResult}</div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
