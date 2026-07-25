import React, { useState, useEffect } from 'react';
import { Activity, Server, Clock, ShieldAlert, CheckCircle2, AlertTriangle, Workflow, ExternalLink, ActivitySquare } from 'lucide-react';
import './index.css';
import VitalsSidebar from './components/VitalsSidebar';
import TemporalStressGauge from './components/TemporalStressGauge';
import InferenceThroughputGraph from './components/InferenceThroughputGraph';
import ConsensusStabilityPanel from './components/ConsensusStabilityPanel';
import WALHeatmap from './components/WALHeatmap';
import NodeHealthPanel from './components/NodeHealthPanel';
import AntigravityChat from './components/AntigravityChat';

const API_BASE = 'http://localhost:8080';

const NodeCard = ({ node }) => {
  const getStatusIcon = (status) => {
    switch (status) {
      case 'EQUIVALENT': return <CheckCircle2 size={18} className="status-EQUIVALENT" />;
      case 'DRIFT_DETECTED': return <AlertTriangle size={18} className="status-DRIFT_DETECTED" />;
      case 'UNREACHABLE': return <ShieldAlert size={18} className="status-UNREACHABLE" />;
      default: return <Activity size={18} />;
    }
  };

  const hashShort = (hash) => hash ? hash.substring(0, 10) + '...' : 'N/A';

  return (
    <div className="glass-card">
      <div className="card-header">
        <div className="card-title">
          <Server size={20} />
          {node.name}
          <span className="tag" style={{ background: 'rgba(255,255,255,0.1)', marginLeft: '8px' }}>{node.role}</span>
        </div>
        <div className="card-title">
          {getStatusIcon(node.drift)}
        </div>
      </div>
      
      <div className="card-row">
        <span className="card-label">State Hash:</span>
        <span className="card-value">{hashShort(node.state_hash)}</span>
      </div>
      <div className="card-row">
        <span className="card-label">Expected:</span>
        <span className="card-value" style={{ color: node.drift === 'DRIFT_DETECTED' ? 'var(--color-yellow)' : 'inherit' }}>
          {hashShort(node.expected)}
        </span>
      </div>
      <div className="card-row">
        <span className="card-label">Health:</span>
        <span className="card-value">{node.health}</span>
      </div>

      {node.models && node.models.length > 0 && (
        <div style={{ marginTop: '0.5rem' }}>
          <div className="card-label" style={{ marginBottom: '0.5rem', fontSize: '0.75rem', textTransform: 'uppercase' }}>Active Models</div>
          {node.models.map((m, i) => (
            <div key={i} className="card-row" style={{ paddingLeft: '0.5rem', borderLeft: '2px solid var(--border-glow)' }}>
              <span>{m.name}</span>
              <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>{m.quant}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default function App() {
  const [atlasState, setAtlasState] = useState({ nodes: {} });
  const [lineage, setLineage] = useState([]);
  const [consensus, setConsensus] = useState([]);
  const [metrics, setMetrics] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [stRes, linRes, conRes, metRes] = await Promise.all([
          fetch(`${API_BASE}/atlas/state`),
          fetch(`${API_BASE}/atlas/lineage`),
          fetch(`${API_BASE}/atlas/consensus`),
          fetch(`${API_BASE}/atlas/metrics`)
        ]);
        
        if (stRes.ok) setAtlasState(await stRes.json());
        if (linRes.ok) setLineage(await linRes.json());
        if (conRes.ok) setConsensus(await conRes.json());
        if (metRes.ok) setMetrics(await metRes.json());
      } catch (err) {
        console.error("Failed to fetch atlas data", err);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 2000);
    return () => clearInterval(interval);
  }, []);

  const getTagClass = (action) => {
    if (action === 'ROLLBACK') return 'tag-rollback';
    if (action === 'ROLLFORWARD') return 'tag-rollforward';
    if (action === 'PROMOTE') return 'tag-promote';
    return '';
  };

  return (
    <div className="dashboard-container">
      <header className="header">
        <h1 className="header-title">Organ Atlas</h1>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
          <Activity size={18} className="live-pulse" style={{ color: 'var(--color-green)' }} />
          <span>System Live</span>
        </div>
      </header>

      <main className="main-grid">
        <div className="nodes-grid">
          {Object.values(atlasState.nodes || {}).map((node) => (
            <NodeCard key={node.name} node={node} />
          ))}
          {Object.keys(atlasState.nodes || {}).length === 0 && (
            <div style={{ color: 'var(--text-secondary)', padding: '2rem' }}>Awaiting nodes...</div>
          )}

          {/* Phase 122 Telemetry Grid below the Node Cards */}
          {metrics && (
            <>
              <InferenceThroughputGraph metrics={metrics} />
              <TemporalStressGauge metrics={metrics} />
              <WALHeatmap metrics={metrics} />
              <ConsensusStabilityPanel metrics={metrics} />
              <NodeHealthPanel metrics={metrics} />
            </>
          )}
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
          {metrics && <VitalsSidebar metrics={metrics} />}
          <div className="glass-card side-panel" style={{ flex: 1 }}>
            <div className="card-title"><Workflow size={20} /> Consensus Decisions</div>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {consensus.map((c, i) => (
                <div key={i} className="list-item">
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <span style={{ fontWeight: 600 }}>{c.node}</span>
                    <span className={`tag ${getTagClass(c.decision)}`}>{c.decision}</span>
                  </div>
                  <div className="card-row">
                    <span className="card-label">Confidence</span>
                    <span>{(c.confidence * 100).toFixed(1)}%</span>
                  </div>
                  <div className="card-row">
                    <span className="card-label">Proposal ID</span>
                    <span style={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>{c.proposal_id.substring(0,8)}</span>
                  </div>
                </div>
              ))}
              {consensus.length === 0 && <div style={{color:'var(--text-secondary)'}}>No active decisions.</div>}
            </div>
          </div>

          <div className="glass-card side-panel" style={{ flex: 1 }}>
            <div className="card-title"><Clock size={20} /> Temporal Lineage</div>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {lineage.map((l, i) => (
                <div key={i} className="list-item">
                  <div style={{ fontWeight: 500 }}>{l.type}</div>
                  <div className="card-row">
                    <span className="card-label">{new Date(l.timestamp).toLocaleTimeString()}</span>
                    {l.hash && <span className="card-value" style={{fontSize: '0.75rem'}}>{l.hash}</span>}
                    {l.node && <span style={{fontSize: '0.875rem'}}>{l.node}</span>}
                  </div>
                </div>
              ))}
              {lineage.length === 0 && <div style={{color:'var(--text-secondary)'}}>No events recorded.</div>}
            </div>
          </div>

          <AntigravityChat />
        </div>
      </main>
    </div>
  );
}
