import React, { useState, useEffect } from 'react';
import { Zap, Activity, RefreshCw, Play, ShieldAlert, Cpu, ArrowRight } from 'lucide-react';
import { MEVMultiLegEngine } from '../engine/MEVMultiLegEngine';

export default function MEVArbitragePanel({ onClose }) {
  const [routes, setRoutes] = useState([]);
  const [maxLegs, setMaxLegs] = useState(7);
  const [networkKey, setNetworkKey] = useState('BASE_MAINNET');
  const [selectedRoute, setSelectedRoute] = useState(null);
  const [simResult, setSimResult] = useState(null);
  const [broadcastResult, setBroadcastResult] = useState(null);
  const [loading, setLoading] = useState(false);

  const handleLiveBroadcast = async (route) => {
    if (!route) return;
    try {
      const res = await fetch('/api/mev/broadcast', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ route, network: networkKey })
      });
      const data = await res.json();
      setBroadcastResult(data);
    } catch (e) {
      console.error('Live broadcast failed:', e);
      setBroadcastResult({ ok: false, error: e.message });
    }
  };

  const fetchOpportunities = async () => {
    try {
      setLoading(true);
      const res = await fetch(`/api/mev/opportunities?maxLegs=${maxLegs}`);
      const data = await res.json();
      if (data.ok) {
        setRoutes(data.routes);
        if (data.routes.length > 0 && !selectedRoute) {
          setSelectedRoute(data.routes[0]);
        }
      }
    } catch (e) {
      console.error('Failed to fetch MEV opportunities:', e);
      // Fallback local engine simulation
      const fallbackRoutes = MEVMultiLegEngine.generateCandidateRoutes(maxLegs);
      setRoutes(fallbackRoutes);
      if (fallbackRoutes.length > 0) setSelectedRoute(fallbackRoutes[0]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOpportunities();
  }, [maxLegs]);

  const handleSimulateShadow = async (route) => {
    if (!route) return;
    try {
      const res = await fetch('/api/mev/trial', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          route_id: route.route_id,
          leg_count: route.leg_count,
          risk_category: route.risk_category
        })
      });
      const data = await res.json();
      if (data.ok) {
        setSimResult(data);
      }
    } catch (e) {
      console.error('Shadow trial execution failed:', e);
      setSimResult(MEVMultiLegEngine.simulateShadowTrial(route));
    }
  };

  const getRiskColor = (cat) => {
    if (cat === 'LOW') return '#10b981';
    if (cat === 'MEDIUM') return '#3b82f6';
    return '#ef4444';
  };

  return (
    <div className="mev-panel" style={{ display: 'flex', flexDirection: 'column', height: '100%', color: '#f5f5f7' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', borderBottom: '1px solid var(--border)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#10b981' }}>
          <Zap size={20} />
          <span style={{ fontWeight: 700, fontSize: '1rem' }}>Variable Multi-Leg MEV Optimization Engine</span>
          <span className="tag" style={{ background: 'rgba(16, 185, 129, 0.2)', color: '#10b981', border: '1px solid #10b981', fontSize: '0.7rem', padding: '0.1rem 0.4rem', borderRadius: '4px' }}>2 TO 7 LEGS</span>
        </div>
        <button onClick={onClose} style={{ background: 'transparent', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', fontSize: '1.2rem' }}>×</button>
      </div>

      {/* Controls Bar */}
      <div style={{ padding: '0.75rem 1rem', background: 'rgba(0,0,0,0.3)', borderBottom: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.3rem' }}>
            <span style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Target Chain:</span>
            <select
              value={networkKey}
              onChange={(e) => setNetworkKey(e.target.value)}
              style={{ background: '#09090b', color: '#10b981', border: '1px solid #10b981', padding: '0.25rem 0.4rem', borderRadius: '4px', fontSize: '0.75rem', fontWeight: 700, cursor: 'pointer' }}>
              <option value="BASE_MAINNET">Base L2 (Coinbase Native, ~$0.03/tx - RECOMMENDED)</option>
              <option value="ARBITRUM_ONE">Arbitrum One L2 (~$0.05/tx)</option>
              <option value="ETH_FLASHBOTS">Ethereum L1 Flashbots Protect (~$12.00+/tx)</option>
            </select>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.3rem' }}>
            <span style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Max Depth:</span>
            {[2, 3, 4, 5, 6, 7].map(l => (
              <button
                key={l}
                onClick={() => setMaxLegs(l)}
                style={{
                  padding: '0.2rem 0.4rem',
                  borderRadius: '4px',
                  fontSize: '0.75rem',
                  fontWeight: 700,
                  cursor: 'pointer',
                  background: maxLegs === l ? '#10b981' : 'rgba(255,255,255,0.05)',
                  color: maxLegs === l ? '#000' : 'var(--text-secondary)',
                  border: maxLegs === l ? '1px solid #10b981' : '1px solid var(--border)'
                }}>
                {l}L
              </button>
            ))}
          </div>
        </div>

        <button onClick={fetchOpportunities} style={{ background: 'transparent', border: '1px solid var(--border)', color: '#f5f5f7', padding: '0.2rem 0.5rem', borderRadius: '4px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.3rem', fontSize: '0.75rem' }}>
          <RefreshCw size={12} /> Scan Arbitrage Graph
        </button>
      </div>

      {/* Main Content: Left Route List, Right Inspector */}
      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Left: Ranked Multi-Leg Opportunities */}
        <div style={{ width: '50%', padding: '1rem', borderRight: '1px solid var(--border)', overflowY: 'auto' }}>
          <div style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.75rem' }}>
            Ranked Arbitrage Routes ({routes.length} Available)
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            {routes.map(r => {
              const isSel = selectedRoute?.route_id === r.route_id;
              const riskColor = getRiskColor(r.risk_category);

              return (
                <div
                  key={r.route_id}
                  onClick={() => { setSelectedRoute(r); setSimResult(null); }}
                  style={{
                    background: isSel ? 'rgba(16, 185, 129, 0.15)' : 'var(--bg-1)',
                    border: `1px solid ${isSel ? '#10b981' : 'var(--border)'}`,
                    borderRadius: '6px',
                    padding: '0.75rem',
                    cursor: 'pointer',
                    transition: 'all 0.15s ease'
                  }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.3rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                      <span style={{ fontWeight: 700, fontSize: '0.85rem' }}>{r.leg_count}-Leg Route</span>
                      <span style={{ fontSize: '0.65rem', padding: '0.1rem 0.3rem', borderRadius: '3px', background: `${riskColor}22`, color: riskColor, border: `1px solid ${riskColor}` }}>
                        {r.risk_category} RISK
                      </span>
                    </div>
                    <span style={{ fontWeight: 700, color: r.net_profit_eth > 0 ? '#10b981' : '#ef4444', fontSize: '0.9rem' }}>
                      {r.net_profit_eth > 0 ? '+' : ''}{r.net_profit_eth} ETH
                    </span>
                  </div>

                  {/* LPV Machine-Native Header Preview */}
                  <pre style={{ background: '#09090b', padding: '0.4rem', borderRadius: '4px', fontSize: '0.65rem', color: '#60a5fa', overflowX: 'auto', margin: 0 }}>
                    {r.lpv_header}
                  </pre>
                </div>
              );
            })}
          </div>
        </div>

        {/* Right: Route Details & Shadow Auditor Simulator */}
        <div style={{ flex: 1, padding: '1rem', overflowY: 'auto', background: 'rgba(0,0,0,0.1)' }}>
          {selectedRoute ? (
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
                <div>
                  <div style={{ fontSize: '0.75rem', color: '#10b981', fontWeight: 600 }}>ROUTE ID: {selectedRoute.route_id}</div>
                  <div style={{ fontSize: '1rem', fontWeight: 700 }}>{selectedRoute.leg_count}-Leg Arbitrage Path</div>
                </div>
                <button
                  onClick={() => handleSimulateShadow(selectedRoute)}
                  style={{ background: 'rgba(16, 185, 129, 0.2)', color: '#10b981', border: '1px solid #10b981', padding: '0.4rem 0.75rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.8rem', display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                  <Play size={14} /> Run Shadow Trial
                </button>
              </div>

              {/* Financial Metrics */}
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '0.5rem', marginBottom: '1rem' }}>
                <div style={{ background: 'var(--bg-1)', padding: '0.5rem', borderRadius: '4px', border: '1px solid var(--border)' }}>
                  <div style={{ fontSize: '0.7rem', color: 'var(--text-secondary)' }}>Input Principal</div>
                  <div style={{ fontSize: '0.9rem', fontWeight: 700 }}>{selectedRoute.input_eth} ETH</div>
                </div>
                <div style={{ background: 'var(--bg-1)', padding: '0.5rem', borderRadius: '4px', border: '1px solid var(--border)' }}>
                  <div style={{ fontSize: '0.7rem', color: 'var(--text-secondary)' }}>Gas Overhead</div>
                  <div style={{ fontSize: '0.9rem', fontWeight: 700, color: '#ef4444' }}>-{selectedRoute.gas_cost_eth} ETH</div>
                </div>
                <div style={{ background: 'var(--bg-1)', padding: '0.5rem', borderRadius: '4px', border: '1px solid var(--border)' }}>
                  <div style={{ fontSize: '0.7rem', color: 'var(--text-secondary)' }}>Net Yield</div>
                  <div style={{ fontSize: '0.9rem', fontWeight: 700, color: '#10b981' }}>+{selectedRoute.net_profit_eth} ETH</div>
                </div>
              </div>

              {/* Multi-Leg Pool Sequence Breakdown */}
              <div style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.4rem' }}>Swaps Pipeline ({selectedRoute.pools.length} Legs):</div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.3rem', marginBottom: '1rem' }}>
                {selectedRoute.pools.map((p, idx) => (
                  <div key={idx} style={{ background: 'var(--bg-1)', padding: '0.4rem 0.6rem', borderRadius: '4px', border: '1px solid var(--border)', fontSize: '0.75rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <span style={{ fontWeight: 700 }}>Leg {idx + 1}: {p.dex} ({p.pair})</span>
                    <span style={{ color: 'var(--text-secondary)' }}>Fee: {(p.fee * 100).toFixed(2)}%</span>
                  </div>
                ))}
              </div>

              {/* Shadow Trial Simulation Output & Live Broadcast */}
              {simResult && (
                <div style={{ background: 'var(--bg-1)', padding: '0.75rem', borderRadius: '6px', border: `1px solid ${simResult.substrate_batch_ready ? '#10b981' : '#ef4444'}`, marginTop: '1rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.3rem' }}>
                    <span style={{ fontSize: '0.8rem', fontWeight: 700, color: simResult.substrate_batch_ready ? '#10b981' : '#ef4444' }}>
                      Auditor Shadow Trial: {simResult.auditor_status}
                    </span>

                    {simResult.substrate_batch_ready && (
                      <button
                        onClick={() => handleLiveBroadcast(selectedRoute)}
                        style={{ background: '#10b981', color: '#000', border: 'none', padding: '0.35rem 0.65rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 700, fontSize: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.3rem' }}>
                        <Zap size={12} /> Broadcast Live MEV Relayer
                      </button>
                    )}
                  </div>

                  <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '0.4rem' }}>
                    Shadow Latency: {simResult.shadow_latency_ms}ms | Target Network: {networkKey}
                  </div>
                  <pre style={{ background: '#09090b', padding: '0.4rem', borderRadius: '4px', fontSize: '0.65rem', color: simResult.substrate_batch_ready ? '#34d399' : '#f87171', margin: 0 }}>
                    {simResult.lpv_status}
                  </pre>
                </div>
              )}

              {/* Broadcast Result Feedback */}
              {broadcastResult && (
                <div style={{ background: broadcastResult.ok ? 'rgba(16, 185, 129, 0.15)' : 'rgba(239, 68, 68, 0.15)', border: `1px solid ${broadcastResult.ok ? '#10b981' : '#ef4444'}`, borderRadius: '6px', padding: '0.75rem', marginTop: '0.75rem' }}>
                  <div style={{ fontSize: '0.8rem', fontWeight: 700, color: broadcastResult.ok ? '#10b981' : '#ef4444' }}>
                    {broadcastResult.ok ? '✅ MEV Batch Relayed Successfully' : '⚠️ Relay Broadcast Result'}
                  </div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.2rem' }}>
                    {broadcastResult.error || `TX Hash: ${broadcastResult.tx_hash} | Nonce: ${broadcastResult.nonce}`}
                  </div>
                  {broadcastResult.lpv_header && (
                    <pre style={{ background: '#09090b', padding: '0.4rem', borderRadius: '4px', fontSize: '0.65rem', color: '#60a5fa', marginTop: '0.4rem', margin: 0 }}>
                      {broadcastResult.lpv_header}
                    </pre>
                  )}
                </div>
              )}
            </div>
          ) : (
            <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem' }}>Select an arbitrage route to inspect multi-leg swap details and trigger pre-emptive shadow trials.</div>
          )}
        </div>
      </div>
    </div>
  );
}
