import React, { useState, useEffect, useRef } from 'react';
import { Shield, Activity, FileText, CheckCircle, XCircle, AlertOctagon, Terminal, Filter, Play, Zap, Pause, SkipBack, SkipForward, Clock, SplitSquareHorizontal, BarChart2, Gavel, Database } from 'lucide-react';
import DiffOverlay from './DiffOverlay';
import CrossAgentComparisonPanel from './CrossAgentComparisonPanel';
import { PreEmptiveAuditor } from '../engine/PreEmptiveAuditor';
import { GovernanceRegistry } from '../engine/GovernanceRegistry';
import GovernanceDashboard from './GovernanceDashboard';
import AgentGovernanceHistory from './AgentGovernanceHistory';

import GovernanceRuleList from './GovernanceRuleList';
import ArbitrationPanel from './ArbitrationPanel';
import AgentDisputeButton from './AgentDisputeButton';
import TimeMachinePanel from './TimeMachinePanel';
import MemoryGraphPanel from './MemoryGraphPanel';
import MEVArbitragePanel from './MEVArbitragePanel';
import MeshDashboardPanel from './MeshDashboardPanel';
import { TimeMachineContext } from '../engine/TimeMachineContext';

const API_BASE = '/api/governance';

const getStatusColor = (status) => {
  if (status === 'TRUSTED') return 'var(--color-green)';
  if (status === 'PROBATION') return 'var(--color-yellow)';
  if (status === 'QUARANTINE') return 'var(--color-red)';
  return 'var(--text-primary)';
};

function GovernanceTimelinePanel({ 
  mode, // 'live' | 'playback'
  data, // { timeline, trustLedger, proposals, forecasts }
  onRatify, 
  onSimulateTampering, 
  onSimulatePropagation,
  onTimelineClick,
  syncEventId
}) {
  const [selectedAgent, setSelectedAgent] = useState(null);
  const [showCompare, setShowCompare] = useState(false);
  const [timelineFilter, setTimelineFilter] = useState('ALL');
  const timelineEndRef = useRef(null);

  const { timeline = [], trustLedger = [], proposals = [], forecasts = [] } = data || {};
  const pendingProposals = proposals.filter(p => p.status === 'PENDING' || p.status === 'PENDING_HUMAN_REVIEW');

  const filteredTimeline = timeline.filter(evt => {
    if (timelineFilter === 'ALL') return true;
    if (timelineFilter === 'TAMPERING' && evt.event_type === 'CRYPTOGRAPHIC_TAMPERING') return true;
    if (timelineFilter === 'RATIFICATION' && evt.event_type.includes('RATIFIED')) return true;
    if (timelineFilter === 'PROPOSALS' && evt.event_type.includes('PROPOSAL')) return true;
    if (timelineFilter === 'TRUST' && evt.event_type === 'TRUST_STATUS_CHANGE') return true;
    return false;
  });

  useEffect(() => {
    if (timelineEndRef.current) {
      timelineEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [timeline, timelineFilter]);

  const isPlayback = mode === 'playback';

  return (
    <div className="main-grid" style={{ display: 'grid', gridTemplateColumns: '1fr 400px', flex: 1, overflow: 'hidden', opacity: isPlayback ? 0.9 : 1, gap: '2rem', padding: '1rem' }}>
      {/* Left Column: Topology & Timeline */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem', height: '100%', overflow: 'hidden' }}>
        
        {/* Trust Ledger (Topology) */}
        <div className="glass-card" style={{ position: 'relative' }}>
          <DiffOverlay markers={data?.markers || []} />
          <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div className="card-title">
              <AlertOctagon size={20} /> Mesh Topology (Trust Ledger)
            </div>
            <button 
              onClick={() => setShowCompare(true)}
              style={{ background: 'var(--color-blue)', color: '#fff', border: 'none', padding: '0.25rem 0.5rem', borderRadius: '4px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.8rem' }}
            >
              <BarChart2 size={14} /> Compare Agents
            </button>
          </div>
          {showCompare && (
            <CrossAgentComparisonPanel 
              agents={trustLedger.map(a => ({
                agentId: a.agent_id,
                trustHistory: Array.from({length: 10}).map((_, i) => ({ t: i, score: a.trust_score - Math.random() * 20 + 10 }))
              }))}
              onClose={() => setShowCompare(false)}
            />
          )}
          <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem', flexWrap: 'wrap', position: 'relative', zIndex: 1 }}>
            {trustLedger.map(agent => (
              <div 
                key={agent.agent_id} 
                className={`list-item ${selectedAgent === agent.agent_id ? 'glow-border' : ''}`}
                onClick={() => setSelectedAgent(selectedAgent === agent.agent_id ? null : agent.agent_id)}
                style={{ 
                  flex: '1 1 200px', 
                  flexDirection: 'column', 
                  alignItems: 'flex-start', 
                  borderLeft: `4px solid ${getStatusColor(agent.status)}`,
                  cursor: 'pointer',
                  transition: 'all 0.2s ease',
                  transform: selectedAgent === agent.agent_id ? 'translateY(-2px)' : 'none',
                  background: selectedAgent === agent.agent_id ? 'rgba(255,255,255,0.05)' : 'var(--bg-0)'
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', marginBottom: '0.5rem' }}>
                  <span style={{ fontWeight: 600, fontSize: '1.1rem' }}>{agent.agent_id.toUpperCase()}</span>
                  <span className="tag" style={{ background: 'var(--bg-0)', border: `1px solid ${getStatusColor(agent.status)}`, color: getStatusColor(agent.status) }}>
                    {agent.status}
                  </span>
                </div>
                <div className="card-row" style={{ width: '100%' }}>
                  <span className="card-label">Trust Score:</span>
                  <span className="card-value" style={{ fontSize: '1.2rem', color: getStatusColor(agent.status) }}>{agent.trust_score}/100</span>
                </div>
                
                {selectedAgent === agent.agent_id && (
                  <div style={{ marginTop: '1rem', width: '100%', borderTop: '1px solid var(--border)', paddingTop: '0.5rem' }}>
                    <div className="card-row">
                      <span className="card-label">Last Eval:</span>
                      <span className="card-value" style={{ fontSize: '0.75rem' }}>{agent.last_evaluated ? new Date(agent.last_evaluated).toLocaleString() : 'N/A'}</span>
                    </div>
                    <div className="card-row" style={{ marginTop: '0.25rem' }}>
                      <span className="card-label">L6 Spine Health:</span>
                      <span className="card-value" style={{ color: agent.status === 'QUARANTINE' || agent.l6_spine?.state === 'BROKEN' ? 'var(--color-red)' : 'var(--color-green)' }}>
                        {agent.status === 'QUARANTINE' || agent.l6_spine?.state === 'BROKEN' ? 'TAMPERED / BROKEN' : 'VERIFIED'}
                      </span>
                    </div>
                    <div style={{ marginTop: '1rem', paddingTop: '0.5rem', borderTop: '1px solid rgba(255, 255, 255, 0.05)' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                        <span className="card-label" style={{ fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>Governance History</span>
                        <AgentDisputeButton 
                          agentId={agent.agent_id} 
                          lastGovernanceEvent={[...timeline].reverse().find(e => (e.agent_id === agent.agent_id || e.agentId === agent.agent_id) && (e.event_id || e.id))} 
                        />
                      </div>
                      <AgentGovernanceHistory agentId={agent.agent_id} />
                    </div>
                  </div>
                )}
              </div>
            ))}
            {trustLedger.length === 0 && <div style={{ color: 'var(--text-secondary)' }}>No agents active.</div>}
          </div>
        </div>

        {/* Timeline (Story of the Mesh) */}
        <div className="glass-card" style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
          <div className="card-header">
            <div className="card-title">
              <Terminal size={20} /> The Story of the Mesh {isPlayback ? '(Playback)' : '(Live)'}
            </div>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <select 
                value={timelineFilter} 
                onChange={(e) => setTimelineFilter(e.target.value)}
                style={{ background: 'var(--bg-0)', color: 'var(--text-primary)', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.25rem 0.5rem', fontSize: '0.8rem' }}
              >
                <option value="ALL">All Events</option>
                <option value="TAMPERING">Tampering</option>
                <option value="TRUST">Trust Changes</option>
                <option value="PROPOSALS">Proposals</option>
                <option value="RATIFICATION">Ratification</option>
              </select>
            </div>
          </div>
          <div style={{ flex: 1, overflowY: 'auto', marginTop: '1rem', background: 'var(--bg-0)', padding: '1rem', borderRadius: '8px', border: '1px solid var(--border)', display: 'flex', flexDirection: 'column', gap: '0.5rem', fontFamily: 'monospace', fontSize: '0.85rem' }}>
            {filteredTimeline.length === 0 ? (
              <div style={{ color: 'var(--text-secondary)' }}>No events recorded for this filter.</div>
            ) : (
              filteredTimeline.map((evt, idx) => {
                let color = 'var(--text-primary)';
                if (evt.event_type === 'CRYPTOGRAPHIC_TAMPERING') color = 'var(--color-red)';
                else if (evt.event_type === 'TRUST_STATUS_CHANGE') color = 'var(--color-yellow)';
                else if (evt.event_type === 'PROPOSAL_RATIFIED') color = 'var(--color-green)';
                else if (evt.event_type === 'POLICY_BROADCAST') color = 'var(--color-blue)';
                else if (evt.event_type === 'DAG_TRUNCATED') color = 'var(--color-purple)';
                else if (evt.event_type === 'FORECAST') color = 'var(--color-orange)';
                else if (evt.event_type === 'PRE_EMPTIVE_QUARANTINE') color = '#ff0055';
                else if (evt.event_type === 'PRE_EMPTIVE_TIGHTEN') color = '#ff8800';
                else if (evt.type === 'GOV_RULE_UPDATED') color = '#00ff99'; // Support new telemetry types too
                else if (evt.type) color = '#ff0055'; 

                let bg = 'transparent';
                let borderLeft = 'none';
                let pl = '0';
                
                if (isPlayback && idx === filteredTimeline.length - 1) {
                   bg = 'rgba(168, 85, 247, 0.1)';
                   borderLeft = '3px solid var(--color-purple)';
                   pl = '0.5rem';
                } else if (!isPlayback && syncEventId === (evt.event_id || evt.id)) {
                   bg = 'rgba(255, 255, 255, 0.05)';
                   borderLeft = '3px solid var(--text-secondary)';
                   pl = '0.5rem';
                }

                return (
                  <div 
                    key={evt.event_id || evt.id} 
                    onClick={() => { if (!isPlayback && onTimelineClick) onTimelineClick(evt); }}
                    style={{ 
                      display: 'flex', 
                      gap: '1rem', 
                      borderBottom: '1px solid var(--border)', 
                      paddingBottom: '0.5rem', 
                      background: bg, 
                      borderLeft: borderLeft, 
                      paddingLeft: pl,
                      cursor: (!isPlayback && onTimelineClick) ? 'pointer' : 'default',
                      transition: 'background 0.2s'
                    }}
                    title={!isPlayback ? "Click to sync playback to this event" : ""}
                  >
                    <span style={{ color: 'var(--text-secondary)', whiteSpace: 'nowrap' }}>
                      [{new Date(evt.timestamp).toLocaleTimeString()}]
                    </span>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <div>
                        <span style={{ color, fontWeight: 'bold' }}>[{evt.event_type || evt.type}]</span>
                        <span style={{ color: 'var(--text-secondary)', marginLeft: '0.5rem' }}>Agent: {evt.agent_id || evt.agentId || 'SYSTEM'}</span>
                      </div>
                      <div style={{ color: 'var(--text-primary)' }}>{evt.description || evt.reason}</div>
                    </div>
                  </div>
                );
              })
            )}
            <div ref={timelineEndRef} />
          </div>
        </div>
      </div>

      {/* Right Column: Pending Proposals, Foresight & Simulation */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', height: '100%', overflow: 'hidden' }}>
        
        {/* Proposal Inbox */}
        <div className="glass-card side-panel" style={{ flex: 1, overflowY: 'auto' }}>
          <div className="card-title">
            <FileText size={20} /> Proposal Inbox {isPlayback && '(Playback)'}
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', marginTop: '1rem', gap: '1rem' }}>
            {pendingProposals.map(p => (
              <div key={p.proposal_id} className="list-item" style={{ flexDirection: 'column', alignItems: 'flex-start', borderLeft: '4px solid var(--color-yellow)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', marginBottom: '0.5rem' }}>
                  <span style={{ fontWeight: 600 }}>{p.type}</span>
                  <span className="tag" style={{ background: 'rgba(234, 179, 8, 0.1)', color: 'var(--color-yellow)', border: '1px solid var(--color-yellow)' }}>
                    {p.risk_level} RISK
                  </span>
                </div>
                <div className="card-row" style={{ width: '100%' }}>
                  <span className="card-label">Target Agent:</span>
                  <span className="card-value" style={{ fontWeight: 'bold' }}>{p.target_agent}</span>
                </div>
                {p.reasoning && (
                  <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginTop: '0.75rem', padding: '0.5rem', background: 'var(--bg-0)', borderRadius: '4px', border: '1px solid var(--border)' }}>
                    <strong>Reasoning:</strong> {p.reasoning}
                  </div>
                )}
                
                <div style={{ marginTop: '0.75rem', fontSize: '0.85rem', color: 'var(--color-blue)', padding: '0.5rem', border: '1px dashed var(--color-blue)', borderRadius: '4px', width: '100%' }}>
                  <strong>Impact Preview:</strong><br />
                  {p.type === 'RECOMMEND_QUARANTINE_RELAXATION' ? `Would lift quarantine for ${p.target_agent} and restore L6 DAG.` : `Applies new policy to ${p.target_agent}.`}
                </div>

                <div style={{ display: 'flex', gap: '0.5rem', width: '100%', marginTop: '1rem' }}>
                  <button 
                    onClick={() => onRatify && onRatify(p.proposal_id, 'APPROVE')}
                    disabled={isPlayback}
                    style={{ flex: 1, padding: '0.5rem', background: isPlayback ? 'var(--bg-0)' : 'rgba(34, 197, 94, 0.1)', color: isPlayback ? 'var(--text-secondary)' : 'var(--color-green)', border: `1px solid ${isPlayback ? 'var(--border)' : 'var(--color-green)'}`, borderRadius: '4px', cursor: isPlayback ? 'not-allowed' : 'pointer', display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.25rem', fontWeight: 600 }}
                  >
                    <CheckCircle size={16} /> Approve
                  </button>
                  <button 
                    onClick={() => onRatify && onRatify(p.proposal_id, 'REJECT')}
                    disabled={isPlayback}
                    style={{ flex: 1, padding: '0.5rem', background: isPlayback ? 'var(--bg-0)' : 'rgba(239, 68, 68, 0.1)', color: isPlayback ? 'var(--text-secondary)' : 'var(--color-red)', border: `1px solid ${isPlayback ? 'var(--border)' : 'var(--color-red)'}`, borderRadius: '4px', cursor: isPlayback ? 'not-allowed' : 'pointer', display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.25rem', fontWeight: 600 }}
                  >
                    <XCircle size={16} /> Reject
                  </button>
                </div>
              </div>
            ))}
            {pendingProposals.length === 0 && (
              <div style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: '3rem 0', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '1rem' }}>
                <Shield size={48} style={{ opacity: 0.2 }} />
                <div>No pending high-risk proposals.</div>
              </div>
            )}
          </div>
        </div>

        {/* Predictive Foresight */}
        <div className="glass-card side-panel" style={{ flex: 1, overflowY: 'auto' }}>
          <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
            <div className="card-title" style={{ color: 'var(--color-orange)' }}>
              <Activity size={20} /> Predictive Foresight
            </div>
            {!isPlayback && (
              <button 
                onClick={() => {
                  fetch('/api/governance/forecast', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ agent_id: 'max', type: 'SPLINTER_RISK', confidence: 0.96, window_ms: 5000, reasoning: 'Simulated Semantic Prediction: High likelihood of L6 Splinter from max in next 5s.' })
                  });
                }}
                style={{ background: 'var(--color-orange)', color: '#fff', border: 'none', padding: '0.25rem 0.5rem', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem', fontWeight: 'bold' }}
              >
                Simulate L6 Splinter
              </button>
            )}
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {forecasts.map(fc => {
               let color = 'var(--color-yellow)';
               if (fc.confidence > 0.6) color = 'var(--color-orange)';
               if (fc.confidence > 0.8) color = 'var(--color-red)';
               return (
                 <div key={fc.forecast_id} className="list-item" style={{ flexDirection: 'column', alignItems: 'flex-start', borderLeft: `4px solid ${color}` }}>
                   <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', marginBottom: '0.25rem' }}>
                     <span style={{ fontWeight: 600 }}>{fc.type}</span>
                     <span className="tag" style={{ background: `rgba(255,255,255,0.05)`, color: color, border: `1px solid ${color}` }}>
                       {(fc.confidence * 100).toFixed(0)}% CONF
                     </span>
                   </div>
                   <div className="card-row" style={{ width: '100%', fontSize: '0.85rem' }}>
                     <span className="card-label">Target:</span>
                     <span className="card-value" style={{ fontWeight: 'bold' }}>{fc.agent_id}</span>
                   </div>
                   <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', marginTop: '0.5rem', fontStyle: 'italic' }}>
                     "{fc.reasoning}"
                   </div>
                 </div>
               );
            })}
            {forecasts.length === 0 && (
              <div style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: '1rem 0', fontSize: '0.85rem' }}>
                No active predictions.
              </div>
            )}
          </div>
        </div>

        {/* Simulation Console */}
        <div className="glass-card side-panel" style={{ marginTop: 'auto', opacity: isPlayback ? 0.5 : 1, pointerEvents: isPlayback ? 'none' : 'auto' }}>
          <div className="card-title" style={{ color: 'var(--color-blue)' }}>
            <Zap size={20} /> Mesh Simulation Console
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', marginTop: '0.5rem', marginBottom: '1rem' }}>
            Inject synthetic events into the sovereign mesh to observe autonomous governance in real-time. Target: <strong>Agent Max</strong>
          </div>
          
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <button 
              onClick={onSimulatePropagation}
              style={{ width: '100%', padding: '0.75rem', background: 'rgba(59, 130, 246, 0.1)', color: 'var(--color-blue)', border: '1px solid var(--color-blue)', borderRadius: '4px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.5rem', fontWeight: 600, transition: 'all 0.2s' }}
            >
              <Play size={16} /> Simulate L1-L5 Propagation
            </button>
            
            <button 
              onClick={onSimulateTampering}
              style={{ width: '100%', padding: '0.75rem', background: 'rgba(239, 68, 68, 0.1)', color: 'var(--color-red)', border: '1px dashed var(--color-red)', borderRadius: '4px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.5rem', fontWeight: 600, transition: 'all 0.2s' }}
            >
              <AlertOctagon size={16} /> Inject L6 Cryptographic Tampering
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function GovernanceView() {
  const [viewMode, setViewMode] = useState('LIVE'); // 'LIVE' | 'PLAYBACK' | 'DUAL'
  const [showDashboard, setShowDashboard] = useState(false);
  const [showRuleEditor, setShowRuleEditor] = useState(false);
  const [showArbitrationPanel, setShowArbitrationPanel] = useState(false);
  const [showTimeMachine, setShowTimeMachine] = useState(false);
  const [showMeshCockpit, setShowMeshCockpit] = useState(false);
  const [showMemoryGraph, setShowMemoryGraph] = useState(false);
  const [showMEVPanel, setShowMEVPanel] = useState(false);
  const [showMeshPanel, setShowMeshPanel] = useState(false);
  const [timeline, setTimeline] = useState([]);
  const [trustLedger, setTrustLedger] = useState([]);
  const [proposals, setProposals] = useState([]);
  const [forecasts, setForecasts] = useState([]);
  const [preEmptiveEnabled, _setPreEmptiveEnabled] = useState(false);
  const preEmptiveEnabledRef = useRef(false);
  const setPreEmptiveEnabled = (val) => {
    preEmptiveEnabledRef.current = val;
    _setPreEmptiveEnabled(val);
  };
  const auditorRef = useRef(new PreEmptiveAuditor(API_BASE));
  
  const [temporalState, setTemporalState] = useState(TimeMachineContext.getTraversalMeta());

  useEffect(() => {
    return TimeMachineContext.subscribe(meta => setTemporalState(meta));
  }, []);
  
  // Playback State
  const [playbackIndex, setPlaybackIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const [playbackState, setPlaybackState] = useState(null);
  const [fullTimeline, setFullTimeline] = useState([]);
  const [telemetryEvents, setTelemetryEvents] = useState([]);

  // Data fetching
  const fetchData = async () => {
    if (viewMode === 'PLAYBACK' && viewMode !== 'DUAL') return; 
    try {
      const [tlRes, trRes, propRes, fcRes] = await Promise.all([
        fetch(`${API_BASE}/timeline`),
        fetch(`${API_BASE}/trust`),
        fetch(`${API_BASE}/proposals`),
        fetch(`${API_BASE}/forecast`)
      ]);
      let fetchedTimeline = [];
      if (tlRes.ok) {
        const tlData = await tlRes.json();
        fetchedTimeline = tlData.timeline || [];
      }
      
      // Inject pure telemetry events into the fetched timeline so we can see them
      const combinedTimeline = [...fetchedTimeline, ...telemetryEvents].sort((a, b) => a.timestamp - b.timestamp);
      setTimeline(combinedTimeline);

      if (trRes.ok) {
        const trData = await trRes.json();
        setTrustLedger(trData.agents || []);
      }
      if (propRes.ok) {
        const propData = await propRes.json();
        setProposals(propData.proposals || []);
      }
      if (fcRes.ok) {
        const fcData = await fcRes.json();
        setForecasts(fcData.forecasts || []);
        
        // Phase 15: Run the Pre-Emptive Auditor
        auditorRef.current.scanAndAct(fcData.forecasts, preEmptiveEnabledRef.current);
      }
    } catch (e) {
      console.error('Failed to fetch governance data:', e);
    }
  };

  useEffect(() => {
    import('../engine/GovernanceTelemetry').then(({ GovernanceTelemetry }) => {
      GovernanceTelemetry.subscribe(evt => {
        setTelemetryEvents(prev => [...prev, evt]);
      });
    });
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 3000);
    return () => clearInterval(interval);
  }, [viewMode, telemetryEvents]);

  // Load Full Timeline when entering Playback or Dual mode
  useEffect(() => {
    if (viewMode === 'PLAYBACK' || viewMode === 'DUAL') {
      fetch(`${API_BASE}/timeline/full`)
        .then(res => res.json())
        .then(data => {
            const combined = [...(data || []), ...telemetryEvents].sort((a, b) => a.timestamp - b.timestamp);
            setFullTimeline(combined);
            // Only jump to end if we haven't set an index yet or if we just entered mode
            if (playbackIndex === 0) {
               setPlaybackIndex(combined.length > 0 ? combined.length - 1 : 0);
            }
        })
        .catch(e => console.error('Failed to fetch full timeline:', e));
    } else {
      setPlaybackState(null);
      setIsPlaying(false);
    }
  }, [viewMode, telemetryEvents]);

  // Fetch State at current playback index
  useEffect(() => {
    if ((viewMode === 'PLAYBACK' || viewMode === 'DUAL') && fullTimeline.length > 0) {
       const evt = fullTimeline[playbackIndex];
       if (evt && evt.event_id) { // Only fetch state for real mesh events, not local telemetry
           const timer = setTimeout(() => {
             fetch(`${API_BASE}/state-at/${evt.timestamp}`)
               .then(res => res.json())
               .then(data => {
                   setPlaybackState(data);
               })
               .catch(e => console.error('Failed to fetch playback state:', e));
           }, 200); // 200ms debounce
           return () => clearTimeout(timer);
       }
    }
  }, [playbackIndex, viewMode, fullTimeline]);

  // Auto-Play
  useEffect(() => {
    if (isPlaying && (viewMode === 'PLAYBACK' || viewMode === 'DUAL')) {
      const id = setInterval(() => {
        setPlaybackIndex(i => {
           if (i < fullTimeline.length - 1) return i + 1;
           setIsPlaying(false);
           return i;
        });
      }, 1000 / playbackSpeed);
      return () => clearInterval(id);
    }
  }, [isPlaying, viewMode, playbackSpeed, fullTimeline]);

  const handleRatify = async (proposalId, action) => {
    try {
      await fetch(`${API_BASE}/ratify/${proposalId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action })
      });
      fetchData();
    } catch (e) {
      console.error('Ratification failed:', e);
    }
  };

  const handleSimulatePropagation = async () => {
    try {
      const payload = {
        region: 1,
        slot: 100,
        payload: { instruction: 'test_topology_broadcast' },
        payloadClass: 'test',
        version: Date.now()
      };
      await fetch('/api/lpv2/propagate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
    } catch (e) {
      console.error('Simulation failed:', e);
    }
  };

  const handleSimulateTampering = async () => {
    const confirmed = window.confirm(
      "⚠️ High-Risk Action\n\n" +
      "This will intentionally break the L6 consensus spine for Max.\n" +
      "The Sovereign Auditor will quarantine the agent."
    );
    if (!confirmed) return;

    try {
      const historyRes = await fetch('/api/l6/history/max');
      const historyData = await historyRes.json();
      const latestRoot = historyData.history[historyData.history.length - 1];
      
      if (!latestRoot) return alert('No L6 spine found for Max yet. Please wait for an initial commit.');

      await fetch('/api/mesh/auditor/commit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent_id: 'max',
          lineage_id: 'lpv2-forgery-' + Date.now(),
          height: latestRoot.height + 1,
          prev_root: '00000000000badbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbad',
          root: '11111111111badbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbad',
          timestamp: Date.now(),
          version: 1
        })
      });
    } catch (e) {
      console.error('Tampering simulation failed:', e);
    }
  };

  const handleTimelineClick = (evt) => {
    if (viewMode === 'DUAL' && fullTimeline.length > 0) {
      const idx = fullTimeline.findIndex(e => (e.event_id || e.id) === (evt.event_id || evt.id));
      if (idx !== -1) {
        setPlaybackIndex(idx);
      }
    }
  };

  const liveData = { timeline, trustLedger, proposals, forecasts };
  
  const playbackData = {
    timeline: fullTimeline.slice(0, playbackIndex + 1),
    trustLedger: playbackState ? playbackState.agents : [],
    proposals: playbackState ? playbackState.proposals : [],
    forecasts: playbackState ? (playbackState.forecasts || []) : []
  };

  const activeSyncEventId = (viewMode === 'DUAL' && fullTimeline[playbackIndex]) ? (fullTimeline[playbackIndex].event_id || fullTimeline[playbackIndex].id) : null;

  return (
    <div className="dashboard-container" style={{ 
      display: 'flex', 
      flexDirection: 'column', 
      height: '100vh', 
      overflow: 'hidden',
      border: temporalState.temporal ? '3px solid #f59e0b' : 'none',
      boxShadow: temporalState.temporal ? '0 0 20px rgba(245, 158, 11, 0.4) inset' : 'none'
    }}>
      {temporalState.temporal && (
        <div style={{ background: '#f59e0b', color: '#000', padding: '0.5rem', textAlign: 'center', fontWeight: 'bold', letterSpacing: '1px', zIndex: 1000, display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.5rem' }}>
          <AlertOctagon size={18} /> 
          ⚠️ TEMPORAL MODE ACTIVE - VIEWING HISTORICAL/COUNTERFACTUAL STATE (Epoch: {new Date(temporalState.targetTimestamp).toLocaleString()})
        </div>
      )}
      <header className="header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', background: (viewMode === 'PLAYBACK' || viewMode === 'DUAL') ? 'rgba(168, 85, 247, 0.1)' : undefined, borderBottom: (viewMode === 'PLAYBACK' || viewMode === 'DUAL') ? '1px solid var(--color-purple)' : undefined }}>
        <h1 className="header-title" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: (viewMode === 'PLAYBACK' || viewMode === 'DUAL') ? 'var(--color-purple)' : undefined }}>
          {viewMode === 'LIVE' ? <Shield size={24} /> : (viewMode === 'DUAL' ? <SplitSquareHorizontal size={24} /> : <Clock size={24} />)} 
          Sovereign Mesh Governance {viewMode === 'PLAYBACK' ? '- PLAYBACK MODE' : (viewMode === 'DUAL' ? '- DUAL REALITY' : '')}
        </h1>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          
          <button 
            onClick={() => setShowMeshPanel(!showMeshPanel)}
            style={{ background: showMeshPanel ? '#0284c7' : 'var(--bg-0)', color: showMeshPanel ? '#fff' : 'var(--text-secondary)', border: `1px solid ${showMeshPanel ? '#0284c7' : 'var(--border)'}`, padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem' }}
          >
            <Activity size={16} /> Mesh OS Core
          </button>

          <button 
            onClick={() => setShowMEVPanel(!showMEVPanel)}
            style={{ background: showMEVPanel ? '#10b981' : 'var(--bg-0)', color: showMEVPanel ? '#000' : 'var(--text-secondary)', border: `1px solid ${showMEVPanel ? '#10b981' : 'var(--border)'}`, padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem' }}
          >
            <Zap size={16} /> MEV Arbitrage
          </button>

          <button 
            onClick={() => setShowMemoryGraph(!showMemoryGraph)}
            style={{ background: showMemoryGraph ? '#a855f7' : 'var(--bg-0)', color: showMemoryGraph ? '#fff' : 'var(--text-secondary)', border: `1px solid ${showMemoryGraph ? '#a855f7' : 'var(--border)'}`, padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem' }}
          >
            <Database size={16} /> 49 MemoryGraph
          </button>

          <button 
            onClick={() => setShowTimeMachine(!showTimeMachine)}
            style={{ background: showTimeMachine ? 'var(--color-purple)' : 'var(--bg-0)', color: showTimeMachine ? '#fff' : 'var(--text-secondary)', border: `1px solid ${showTimeMachine ? 'var(--color-purple)' : 'var(--border)'}`, padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem' }}
          >
            <Clock size={16} /> Time Machine
          </button>

          <button 
            onClick={() => setShowArbitrationPanel(!showArbitrationPanel)}
            style={{ background: showArbitrationPanel ? 'var(--color-purple)' : 'var(--bg-0)', color: showArbitrationPanel ? '#fff' : 'var(--text-secondary)', border: `1px solid ${showArbitrationPanel ? 'var(--color-purple)' : 'var(--border)'}`, padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem' }}
          >
            <Gavel size={16} /> Arbitration
          </button>

          <button 
            onClick={() => setShowRuleEditor(!showRuleEditor)}
            style={{ background: showRuleEditor ? 'var(--color-green)' : 'var(--bg-0)', color: showRuleEditor ? '#000' : 'var(--text-secondary)', border: `1px solid ${showRuleEditor ? 'var(--color-green)' : 'var(--border)'}`, padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem' }}
          >
            <Terminal size={16} /> Live Rules
          </button>

          <button 
            onClick={() => setShowDashboard(!showDashboard)}
            style={{ background: showDashboard ? 'var(--color-blue)' : 'var(--bg-0)', color: showDashboard ? '#fff' : 'var(--text-secondary)', border: '1px solid var(--border)', padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem' }}
          >
            <Activity size={16} /> Governance Dashboard
          </button>

          <button 
            onClick={() => setShowMeshCockpit(true)}
            style={{ background: '#3b82f6', color: '#fff', border: 'none', padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem', boxShadow: '0 0 10px rgba(59, 130, 246, 0.5)' }}
          >
            👑 Mesh OS Cockpit (Phase 27)
          </button>

          {GovernanceRegistry.isRatified("preemptive_auditor_v1") && (
            <span className="governance-pill">Governance: RATIFIED</span>
          )}

          {/* Pre-Emptive Auditor Toggle */}
          <div style={{ display: 'flex', alignItems: 'center', background: 'var(--bg-0)', borderRadius: '4px', border: `1px solid ${preEmptiveEnabled ? 'var(--color-green)' : 'var(--border)'}`, overflow: 'hidden' }}>
            <button 
              onClick={() => setPreEmptiveEnabled(!preEmptiveEnabled)}
              style={{ padding: '0.5rem 1rem', background: preEmptiveEnabled ? 'rgba(0,255,153,0.1)' : 'transparent', color: preEmptiveEnabled ? 'var(--color-green)' : 'var(--text-secondary)', border: 'none', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem' }}
            >
              <Shield size={16} /> Pre-Emptive Auditor: {preEmptiveEnabled ? 'ON' : 'OFF'}
            </button>
          </div>

          <div style={{ display: 'flex', background: 'var(--bg-0)', borderRadius: '4px', overflow: 'hidden', border: '1px solid var(--border)' }}>
            <button 
              onClick={() => setViewMode('LIVE')}
              style={{ padding: '0.5rem 1rem', background: viewMode === 'LIVE' ? 'var(--color-blue)' : 'transparent', color: viewMode === 'LIVE' ? '#fff' : 'var(--text-secondary)', border: 'none', cursor: 'pointer', fontWeight: 600, borderRight: '1px solid var(--border)' }}
            >
              Live
            </button>
            <button 
              onClick={() => setViewMode('PLAYBACK')}
              style={{ padding: '0.5rem 1rem', background: viewMode === 'PLAYBACK' ? 'var(--color-purple)' : 'transparent', color: viewMode === 'PLAYBACK' ? '#fff' : 'var(--text-secondary)', border: 'none', cursor: 'pointer', fontWeight: 600, borderRight: '1px solid var(--border)' }}
            >
              Playback
            </button>
            <button 
              onClick={() => setViewMode('DUAL')}
              style={{ padding: '0.5rem 1rem', background: viewMode === 'DUAL' ? 'linear-gradient(90deg, var(--color-blue) 0%, var(--color-purple) 100%)' : 'transparent', color: viewMode === 'DUAL' ? '#fff' : 'var(--text-secondary)', border: 'none', cursor: 'pointer', fontWeight: 600 }}
            >
              Dual Reality
            </button>
          </div>

          {(viewMode === 'LIVE' || viewMode === 'DUAL') && (
             <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
               <Activity size={18} className="live-pulse" style={{ color: 'var(--color-blue)' }} />
               <span>L7 Auditor Active</span>
             </div>
          )}
        </div>
      </header>

      {showDashboard && <GovernanceDashboard onClose={() => setShowDashboard(false)} />}

      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        <main style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
          {viewMode === 'LIVE' && (
            <GovernanceTimelinePanel 
              mode="live" 
              data={liveData} 
              onRatify={handleRatify}
              onSimulateTampering={handleSimulateTampering}
              onSimulatePropagation={handleSimulatePropagation}
              onTimelineClick={handleTimelineClick}
              syncEventId={activeSyncEventId}
            />
          )}
          {viewMode === 'PLAYBACK' && (
            <GovernanceTimelinePanel 
              mode="playback" 
              data={playbackData} 
            />
          )}
          {viewMode === 'DUAL' && (
            <div style={{ display: 'flex', width: '100%', height: '100%' }}>
              <div style={{ flex: 1, borderRight: '1px solid var(--border)', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
                  <div style={{ padding: '0.5rem 1rem', background: 'rgba(59, 130, 246, 0.1)', color: 'var(--color-blue)', fontWeight: 'bold', borderBottom: '1px solid var(--border)', textAlign: 'center', letterSpacing: '1px' }}>LIVE STATE</div>
                  <GovernanceTimelinePanel 
                    mode="live" 
                    data={liveData} 
                    onRatify={handleRatify}
                    onSimulateTampering={handleSimulateTampering}
                    onSimulatePropagation={handleSimulatePropagation}
                    onTimelineClick={handleTimelineClick}
                    syncEventId={activeSyncEventId}
                  />
              </div>
              <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
                  <div style={{ padding: '0.5rem 1rem', background: 'rgba(168, 85, 247, 0.1)', color: 'var(--color-purple)', fontWeight: 'bold', borderBottom: '1px solid var(--border)', textAlign: 'center', letterSpacing: '1px' }}>PLAYBACK STATE</div>
                  <GovernanceTimelinePanel 
                    mode="playback" 
                    data={playbackData} 
                  />
              </div>
            </div>
          )}
        </main>
        {showRuleEditor && (
          <div style={{ width: '450px', borderLeft: '1px solid var(--border)', background: 'var(--bg-0)', display: 'flex', flexDirection: 'column' }}>
            <GovernanceRuleList onClose={() => setShowRuleEditor(false)} />
          </div>
        )}
        {showArbitrationPanel && (
          <div style={{ width: '450px', borderLeft: '1px solid var(--border)', background: 'var(--bg-0)', display: 'flex', flexDirection: 'column' }}>
            <ArbitrationPanel operatorId="sysadmin" onClose={() => setShowArbitrationPanel(false)} />
          </div>
        )}
        {showTimeMachine && (
          <div style={{ width: '450px', borderLeft: '1px solid var(--border)', background: 'var(--bg-0)', display: 'flex', flexDirection: 'column' }}>
            <TimeMachinePanel onClose={() => setShowTimeMachine(false)} />
          </div>
        )}
        {showMemoryGraph && (
          <div style={{ width: '800px', borderLeft: '1px solid var(--border)', background: 'var(--bg-0)', display: 'flex', flexDirection: 'column', zIndex: 50 }}>
            <MemoryGraphPanel onClose={() => setShowMemoryGraph(false)} />
          </div>
        )}
        {showMEVPanel && (
          <div style={{ width: '850px', borderLeft: '1px solid var(--border)', background: 'var(--bg-0)', display: 'flex', flexDirection: 'column', zIndex: 60 }}>
            <MEVArbitragePanel onClose={() => setShowMEVPanel(false)} />
          </div>
        )}
        {showMeshPanel && (
          <div style={{ width: '850px', borderLeft: '1px solid var(--border)', background: 'var(--bg-0)', display: 'flex', flexDirection: 'column', zIndex: 70 }}>
            <MeshDashboardPanel onClose={() => setShowMeshPanel(false)} />
          </div>
        )}
      </div>

      {/* Playback Bar */}
      {(viewMode === 'PLAYBACK' || viewMode === 'DUAL') && (
        <div style={{ background: 'var(--bg-1)', borderTop: '1px solid var(--color-purple)', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '1rem', zIndex: 100 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <button 
              onClick={() => setIsPlaying(!isPlaying)}
              style={{ background: 'var(--color-purple)', color: '#fff', border: 'none', borderRadius: '50%', width: '40px', height: '40px', display: 'flex', justifyContent: 'center', alignItems: 'center', cursor: 'pointer' }}
            >
              {isPlaying ? <Pause size={20} /> : <Play size={20} />}
            </button>
            <button 
              onClick={() => setPlaybackIndex(Math.max(0, playbackIndex - 1))}
              style={{ background: 'var(--bg-0)', border: '1px solid var(--border)', color: 'var(--text-primary)', borderRadius: '4px', padding: '0.5rem', cursor: 'pointer' }}
            >
              <SkipBack size={16} />
            </button>
            <button 
              onClick={() => setPlaybackIndex(Math.min(fullTimeline.length - 1, playbackIndex + 1))}
              style={{ background: 'var(--bg-0)', border: '1px solid var(--border)', color: 'var(--text-primary)', borderRadius: '4px', padding: '0.5rem', cursor: 'pointer' }}
            >
              <SkipForward size={16} />
            </button>

            <select 
              value={playbackSpeed}
              onChange={(e) => setPlaybackSpeed(Number(e.target.value))}
              style={{ background: 'var(--bg-0)', color: 'var(--text-primary)', border: '1px solid var(--border)', borderRadius: '4px', padding: '0.5rem', marginLeft: '1rem' }}
            >
              <option value="1">1x Speed</option>
              <option value="2">2x Speed</option>
              <option value="5">5x Speed</option>
              <option value="10">10x Speed</option>
            </select>

            <div style={{ flex: 1, display: 'flex', alignItems: 'center', gap: '1rem', paddingLeft: '2rem' }}>
              <input 
                type="range" 
                min="0" 
                max={Math.max(0, fullTimeline.length - 1)} 
                value={playbackIndex}
                onChange={(e) => {
                  setPlaybackIndex(Number(e.target.value));
                  setIsPlaying(false);
                }}
                style={{ flex: 1, accentColor: 'var(--color-purple)' }}
              />
              <span style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', minWidth: '150px', textAlign: 'right' }}>
                {fullTimeline.length > 0 && fullTimeline[playbackIndex] ? new Date(fullTimeline[playbackIndex].timestamp).toLocaleTimeString() : '--:--:--'}
              </span>
            </div>
          </div>
        </div>
      )}
      {showMeshCockpit && (
        <div style={{ position: 'fixed', inset: 0, zIndex: 9999, background: 'rgba(2, 6, 23, 0.95)', padding: '2rem', overflow: 'auto' }}>
          <MeshDashboardPanel onClose={() => setShowMeshCockpit(false)} />
        </div>
      )}
    </div>
  );
}
