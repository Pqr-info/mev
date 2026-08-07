import React, { useState, useEffect } from 'react';

export default function MeshDashboardPanel({ onClose }) {
  const [nodes, setNodes] = useState([]);
  const [arbitrationResult, setArbitrationResult] = useState(null);
  const [history, setHistory] = useState([]);
  const [personality, setPersonality] = useState(null);

  useEffect(() => {
    fetchNodes();
    fetchHistory();
    fetchPersonality();
    const interval = setInterval(() => {
      fetchNodes();
      fetchHistory();
      fetchPersonality();
    }, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchPersonality = async () => {
    try {
      const res = await fetch('/sos/mesh/personality');
      const data = await res.json();
      if (data.ok) setPersonality(data.personality);
    } catch (err) {
      console.error("Error fetching personality:", err);
    }
  };

  const fetchHistory = async () => {
    try {
      const res = await fetch('/sos/mesh/history');
      const data = await res.json();
      if (data.ok) {
        setHistory(data.history);
      }
    } catch (err) {
      console.error("Error fetching history:", err);
    }
  };

  const fetchNodes = async () => {
    try {
      const res = await fetch('/sos/mesh/nodes');
      const data = await res.json();
      if (data.ok) {
        setNodes(Object.values(data.registry));
      }
    } catch (err) {
      console.error("Error fetching mesh nodes:", err);
    }
  };

  const [healResult, setHealResult] = useState(null);
  const [optimizerResult, setOptimizerResult] = useState(null);
  const [behaviorResult, setBehaviorResult] = useState(null);
  const [forecastResult, setForecastResult] = useState(null);

  const [apexResult, setApexResult] = useState(null);

  const handleApexPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/sovereign27', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setApexResult(data.sovereign27);
    } catch (err) {
      console.error("Apex error:", err);
    }
  };
  const [faucetResult, setFaucetResult] = useState(null);

  const handleFaucetDrip = async (nodeId = 'fra', network = 'BASE_SEPOLIA') => {
    try {
      const res = await fetch('/sos/faucet/drip', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ node_id: nodeId, network })
      });
      const data = await res.json();
      if (data.ok) setFaucetResult(data);
    } catch (err) {
      console.error("Faucet error:", err);
    }
  };

  const handleJudiciaryPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/judiciary', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setJudiciaryResult(data.judiciary);
    } catch (err) {
      console.error("Judiciary error:", err);
    }
  };

  const handleEvolutionPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/evolution', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setEvolutionResult(data.evolution);
    } catch (err) {
      console.error("Evolution error:", err);
    }
  };

  const handleEconomicsPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/economics', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setEconomicsResult(data.economics);
    } catch (err) {
      console.error("Economics error:", err);
    }
  };

  const handleCulturePulse = async () => {
    try {
      const res = await fetch('/sos/mesh/culture', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setCultureResult(data.culture);
    } catch (err) {
      console.error("Culture error:", err);
    }
  };

  const handleSovereigntyPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/sovereignty', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setSovereigntyResult(data.sovereignty);
    } catch (err) {
      console.error("Sovereignty error:", err);
    }
  };

  const handleDiplomacyPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/diplomacy', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setDiplomacyResult(data.diplomacy);
    } catch (err) {
      console.error("Diplomacy error:", err);
    }
  };

  const handleConsciencePulse = async () => {
    try {
      const res = await fetch('/sos/mesh/conscience', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setConscienceResult(data.conscience);
    } catch (err) {
      console.error("Conscience error:", err);
    }
  };

  const handleIntentPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/intent', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setIntentResult(data.intent);
    } catch (err) {
      console.error("Intent error:", err);
    }
  };

  const handleBehaviorPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/behavior', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setBehaviorResult(data.behavior_memory);
    } catch (err) {
      console.error("Behavior error:", err);
    }
  };

  const handleForesightPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/forecast', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setForecastResult(data.forecast);
    } catch (err) {
      console.error("Forecast error:", err);
    }
  };

  const handleSelfHealPulse = async () => {
    try {
      const res = await fetch('/sos/mesh/self-heal', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ failed_node: 'fra' })
      });
      const data = await res.json();
      if (data.ok) setHealResult(data.self_heal);
    } catch (err) {
      console.error("Self heal error:", err);
    }
  };

  const handleProfitOptimize = async () => {
    try {
      const res = await fetch('/sos/mesh/optimize', { method: 'POST' });
      const data = await res.json();
      if (data.ok) setOptimizerResult(data.profit_optimizer);
    } catch (err) {
      console.error("Optimizer error:", err);
    }
  };

  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      backgroundColor: 'rgba(5, 7, 15, 0.92)', backdropFilter: 'blur(12px)',
      zIndex: 9999, display: 'flex', flexDirection: 'column', padding: '24px',
      color: '#e2e8f0', fontFamily: 'Inter, monospace', overflowY: 'auto'
    }}>
      {/* Swarm Behavioral Personality & Trust Index Header */}
      {personality && (
        <div style={{
          background: 'rgba(168, 85, 247, 0.12)', border: '1px solid #a855f7',
          borderRadius: '10px', padding: '14px 20px', marginBottom: '20px', display: 'flex', justifyContent: 'space-between', alignItems: 'center'
        }}>
          <div>
            <div style={{ color: '#c084fc', fontSize: '15px', fontWeight: 'bold', display: 'flex', alignItems: 'center', gap: '8px' }}>
              <span>🧠</span> Swarm Behavioral Trait: {personality.swarm_trait}
            </div>
            <div style={{ color: '#e9d5ff', fontSize: '12px', marginTop: '4px' }}>
              Global Swarm Trust Index: <strong style={{ color: '#4ade80' }}>{(personality.global_trust_index * 100).toFixed(1)}%</strong> | Machine Personality Header:
            </div>
            <div style={{ background: '#020617', padding: '4px 8px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#c084fc', marginTop: '4px', display: 'inline-block' }}>
              {personality.lpv_personality_header}
            </div>
          </div>
          <div style={{ textAlign: 'right' }}>
            <span style={{ fontSize: '12px', color: '#a855f7', border: '1px solid #a855f7', padding: '4px 10px', borderRadius: '12px', fontWeight: 'bold' }}>
              PHASE 9: EVOLVING
            </span>
          </div>
        </div>
      )}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px', borderBottom: '1px solid rgba(255, 255, 255, 0.1)', paddingBottom: '12px' }}>
        <div>
          <h2 style={{ margin: 0, color: '#38bdf8', fontSize: '22px', display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span>🌐</span> Sovereign-27 Tri-Region Mesh Dashboard & LPV Arbitration
          </h2>
          <span style={{ fontSize: '12px', color: '#94a3b8' }}>
            [LPV-SWARM-ACTIVE|NODES:4|TOPOLOGY:TRI_REGION|LATENCY_CLASS:EU_ULTRA]
          </span>
        </div>
        <button onClick={onClose} style={{
          background: 'none', border: '1px solid rgba(255,255,255,0.2)', color: '#94a3b8',
          padding: '6px 16px', borderRadius: '6px', cursor: 'pointer', fontSize: '14px'
        }}>✕ Close</button>
      </div>

      {/* Roster & Node Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '16px', marginBottom: '24px' }}>
        {nodes.map(n => (
          <div key={n.node_id} style={{
            background: n.node_id === 'fra' ? 'rgba(14, 165, 233, 0.12)' : 'rgba(30, 41, 59, 0.6)',
            border: n.node_id === 'fra' ? '1px solid #38bdf8' : '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '10px', padding: '16px', display: 'flex', flexDirection: 'column', gap: '8px'
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontWeight: 'bold', fontSize: '16px', color: n.node_id === 'fra' ? '#38bdf8' : '#f1f5f9' }}>
                {n.hostname} ({n.node_id.toUpperCase()})
              </span>
              <span style={{
                fontSize: '10px', padding: '2px 8px', borderRadius: '12px', fontWeight: 'bold',
                backgroundColor: n.status === 'ONLINE' ? 'rgba(34, 197, 94, 0.2)' : 'rgba(239, 68, 68, 0.2)',
                color: n.status === 'ONLINE' ? '#4ade80' : '#f87171',
                border: n.status === 'ONLINE' ? '1px solid #22c55e' : '1px solid #ef4444'
              }}>{n.status}</span>
            </div>
            <div style={{ fontSize: '12px', color: '#94a3b8' }}>IP: <code style={{ color: '#cbd5e1' }}>{n.ip}</code></div>
            <div style={{ fontSize: '12px', color: '#94a3b8' }}>Role: <span style={{ color: '#e2e8f0' }}>{n.role}</span></div>
            <div style={{ fontSize: '12px', color: '#94a3b8' }}>CPU: <span style={{ color: '#e2e8f0' }}>{n.cpu}</span></div>
            <div style={{ fontSize: '12px', color: '#94a3b8' }}>Latency: <span style={{ color: '#38bdf8', fontWeight: 'bold' }}>{n.latency || '<1.8ms'}</span></div>
          </div>
        ))}
      </div>

      {/* LPV Arbitration Engine & Swarm Behaviors Section */}
      <div style={{
        background: 'rgba(15, 23, 42, 0.8)', border: '1px solid rgba(255, 255, 255, 0.1)',
        borderRadius: '10px', padding: '20px', marginBottom: '24px'
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px' }}>
          <div>
            <h3 style={{ margin: 0, color: '#f59e0b', fontSize: '16px' }}>⚡ LPV Orchestration & Self-Healing Behaviors (Phases 5-8)</h3>
            <span style={{ fontSize: '12px', color: '#94a3b8' }}>Dynamic node arbitration, profit optimization & autonomous failover</span>
          </div>
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            <button onClick={() => handleFaucetDrip('fra', 'BASE_SEPOLIA')} style={{
              backgroundColor: '#06b6d4', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>💧 Drip Testnet Funds</button>
            <button onClick={handleApexPulse} style={{
              backgroundColor: '#3b82f6', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer',
              boxShadow: '0 0 10px rgba(59, 130, 246, 0.5)'
            }}>👑 Phase 27: NBEP 2.0 Apex</button>
            <button onClick={handleEvolutionPulse} style={{
              backgroundColor: '#8b5cf6', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 17: Evolution</button>
            <button onClick={handleJudiciaryPulse} style={{
              backgroundColor: '#0284c7', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 16: Judiciary</button>
            <button onClick={handleEconomicsPulse} style={{
              backgroundColor: '#eab308', color: '#0f172a', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 15: Economics</button>
            <button onClick={handleCulturePulse} style={{
              backgroundColor: '#ec4899', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 14: Culture</button>
            <button onClick={handleDiplomacyPulse} style={{
              backgroundColor: '#6366f1', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 13: Diplomacy</button>
            <button onClick={handleConsciencePulse} style={{
              backgroundColor: '#e11d48', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 12: Conscience</button>
            <button onClick={handleIntentPulse} style={{
              backgroundColor: '#f43f5e', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 11: Intent</button>
            <button onClick={handleForesightPulse} style={{
              backgroundColor: '#38bdf8', color: '#0f172a', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 10: Foresight</button>
            <button onClick={handleBehaviorPulse} style={{
              backgroundColor: '#a855f7', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 9: Memory</button>
            <button onClick={handleProfitOptimize} style={{
              backgroundColor: '#10b981', color: '#0f172a', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 7: Optimize</button>
            <button onClick={handleSelfHealPulse} style={{
              backgroundColor: '#ef4444', color: '#fff', border: 'none',
              padding: '8px 12px', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer'
            }}>Phase 8: Failover</button>
          </div>
        </div>

        {faucetResult && (
          <div style={{ background: 'rgba(6, 182, 212, 0.15)', borderRadius: '6px', padding: '12px', border: '1px solid #06b6d4', marginBottom: '10px' }}>
            <div style={{ color: '#22d3ee', fontSize: '13px', fontWeight: 'bold' }}>
              💧 Dripped {faucetResult.dripped_amount} ({faucetResult.network}) to Node [{faucetResult.node_id.toUpperCase()}] ({faucetResult.wallet_address})
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#22d3ee', marginTop: '6px' }}>
              {faucetResult.lpv_header}
            </div>
          </div>
        )}

        {apexResult && (
          <div style={{ background: 'rgba(59, 130, 246, 0.15)', borderRadius: '6px', padding: '12px', border: '1px solid #3b82f6', marginBottom: '10px' }}>
            <div style={{ color: '#60a5fa', fontSize: '14px', fontWeight: 'bold' }}>
              👑 Apex Status: {apexResult.apex_status} | Active Phases: {apexResult.active_phases}/27 | Swarm Cohesion: {(apexResult.cohesion * 100).toFixed(1)}%
            </div>
            <div style={{ background: '#020617', padding: '8px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#60a5fa', marginTop: '6px' }}>
              {apexResult.lpv_apex_header}
            </div>
          </div>
        )}

        {evolutionResult && (
          <div style={{ background: 'rgba(139, 92, 246, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #8b5cf6', marginBottom: '10px' }}>
            <div style={{ color: '#c084fc', fontSize: '13px', fontWeight: 'bold' }}>
              Epoch: {evolutionResult.evolution_epoch} | Hel Fast Mutation: {evolutionResult.node_mutations.hel_fast}
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#c084fc', marginTop: '6px' }}>
              {evolutionResult.lpv_evolution_header}
            </div>
          </div>
        )}

        {judiciaryResult && (
          <div style={{ background: 'rgba(2, 132, 199, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #0284c7', marginBottom: '10px' }}>
            <div style={{ color: '#38bdf8', fontSize: '13px', fontWeight: 'bold' }}>
              Court: {judiciaryResult.judiciary_status} | Adjudication: {judiciaryResult.recent_adjudications[0].case_id} ({judiciaryResult.recent_adjudications[0].verdict})
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#38bdf8', marginTop: '6px' }}>
              {judiciaryResult.lpv_judiciary_header}
            </div>
          </div>
        )}

        {economicsResult && (
          <div style={{ background: 'rgba(234, 179, 8, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #eab308', marginBottom: '10px' }}>
            <div style={{ color: '#fde047', fontSize: '13px', fontWeight: 'bold' }}>
              Market: {economicsResult.compute_market_status} | Gas Oracle: {economicsResult.gas_price_oracle_gwei} Gwei (FRA Tariff: {economicsResult.node_compute_tariffs.fra_ultra})
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#fde047', marginTop: '6px' }}>
              {economicsResult.lpv_economics_header}
            </div>
          </div>
        )}

        {cultureResult && (
          <div style={{ background: 'rgba(236, 72, 153, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #ec4899', marginBottom: '10px' }}>
            <div style={{ color: '#f472b6', fontSize: '13px', fontWeight: 'bold' }}>
              Cooperative Norm: {cultureResult.cooperative_norm} | Swarm Cohesion: {(cultureResult.swarm_cohesion_score * 100).toFixed(1)}%
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#f472b6', marginTop: '6px' }}>
              {cultureResult.lpv_culture_header}
            </div>
          </div>
        )}

        {sovereigntyResult && (
          <div style={{ background: 'rgba(139, 92, 246, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #8b5cf6', marginBottom: '10px' }}>
            <div style={{ color: '#a78bfa', fontSize: '13px', fontWeight: 'bold' }}>
              Swarm Sovereignty: {sovereigntyResult.mesh_state} | Consensus: {sovereigntyResult.global_consensus} ({sovereigntyResult.active_nodes_count}/7 Nodes)
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#a78bfa', marginTop: '6px' }}>
              {sovereigntyResult.lpv_sovereignty_header}
            </div>
          </div>
        )}

        {diplomacyResult && (
          <div style={{ background: 'rgba(99, 102, 241, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #6366f1', marginBottom: '10px' }}>
            <div style={{ color: '#818cf8', fontSize: '13px', fontWeight: 'bold' }}>
              Diplomatic Status: {diplomacyResult.negotiation_status} | Offload: {diplomacyResult.resource_rebalance.fra}
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#818cf8', marginTop: '6px' }}>
              {diplomacyResult.lpv_diplomacy_header}
            </div>
          </div>
        )}

        {conscienceResult && (
          <div style={{ background: 'rgba(225, 29, 72, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #e11d48', marginBottom: '10px' }}>
            <div style={{ color: '#fda4af', fontSize: '13px', fontWeight: 'bold' }}>
              Safety Envelope: {conscienceResult.safety_envelope} | Anti-Toxic MEV: {conscienceResult.anti_toxic_mev} (Max Drawdown: {conscienceResult.max_drawdown_per_block_eth} ETH)
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#fda4af', marginTop: '6px' }}>
              {conscienceResult.lpv_conscience_header}
            </div>
          </div>
        )}

        {intentResult && (
          <div style={{ background: 'rgba(244, 63, 94, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #f43f5e', marginBottom: '10px' }}>
            <div style={{ color: '#fb7185', fontSize: '13px', fontWeight: 'bold' }}>
              Swarm Autonomous Goal: {intentResult.primary_goal} (Target: {intentResult.target_yield_24h_eth} ETH | Policy: {intentResult.active_intent_policy})
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#fb7185', marginTop: '6px' }}>
              {intentResult.lpv_intent_header}
            </div>
          </div>
        )}

        {behaviorResult && (
          <div style={{ background: 'rgba(168, 85, 247, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #a855f7', marginBottom: '10px' }}>
            <div style={{ color: '#c084fc', fontSize: '13px', fontWeight: 'bold' }}>
              7d Swarm Memory Yield: +{behaviorResult.yield_memory_7d_eth} ETH (Historical Bias: FRA 62%, HEL_FAST 21%, DAL 17%)
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#c084fc', marginTop: '6px' }}>
              {behaviorResult.lpv_behavior_header}
            </div>
          </div>
        )}

        {forecastResult && (
          <div style={{ background: 'rgba(56, 189, 248, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #38bdf8', marginBottom: '10px' }}>
            <div style={{ color: '#38bdf8', fontSize: '13px', fontWeight: 'bold' }}>
              Predictive Foresight: Next Block Yield +{forecastResult.predicted_next_block_yield_eth} ETH (Risk: {forecastResult.forecasted_congestion_risk})
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#38bdf8', marginTop: '6px' }}>
              {forecastResult.lpv_forecast_header}
            </div>
          </div>
        )}

        {optimizerResult && (
          <div style={{ background: 'rgba(16, 185, 129, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #10b981', marginBottom: '10px' }}>
            <div style={{ color: '#34d399', fontSize: '13px', fontWeight: 'bold' }}>
              24h Mesh Yield: +{optimizerResult.mesh_yield_24h_eth} ETH (Avg Latency: {optimizerResult.avg_mesh_latency_ms}ms)
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#34d399', marginTop: '6px' }}>
              {optimizerResult.lpv_opt_header}
            </div>
          </div>
        )}

        {healResult && (
          <div style={{ background: 'rgba(239, 68, 68, 0.1)', borderRadius: '6px', padding: '12px', border: '1px solid #ef4444' }}>
            <div style={{ color: '#f87171', fontSize: '13px', fontWeight: 'bold' }}>
              Simulated Failure: {healResult.detected_failure} &rarr; Autonomous Reroute: <strong style={{ color: '#4ade80' }}>{healResult.failover_node}</strong> ({healResult.recovery_latency_ms}ms recovery)
            </div>
            <div style={{ background: '#020617', padding: '6px', borderRadius: '4px', fontSize: '11px', fontFamily: 'monospace', color: '#f87171', marginTop: '6px' }}>
              {healResult.lpv_heal_header}
            </div>
          </div>
        )}
      </div>

      {/* Real-time LPV Bundle Stream Monitor */}
      <div style={{
        background: 'rgba(15, 23, 42, 0.8)', border: '1px solid rgba(255, 255, 255, 0.1)',
        borderRadius: '10px', padding: '20px'
      }}>
        <h3 style={{ margin: '0 0 10px 0', color: '#4ade80', fontSize: '16px' }}>📜 Phase 6 Adaptive Arbitration History & Historical Yield Bias Log</h3>
        <div style={{
          background: '#020617', borderRadius: '6px', padding: '12px', height: '140px', overflowY: 'auto',
          fontFamily: 'monospace', fontSize: '12px', color: '#4ade80', display: 'flex', flexDirection: 'column', gap: '4px'
        }}>
          {history.map(item => (
            <div key={item.id} style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px rgba(255,255,255,0.05) solid', paddingBottom: '2px' }}>
              <span>[{item.timestamp.slice(11, 19)}] {item.route_id} &rarr; <strong style={{ color: '#38bdf8' }}>{item.assigned_node.toUpperCase()}</strong></span>
              <span style={{ color: '#f59e0b' }}>Net: +{item.net_eth} ETH | Latency: {item.latency_ms}ms</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
