import React from "react";
import './CrossAgentComparisonPanel.css';

export default function CrossAgentComparisonPanel({ agents, onClose }) {
  if (!agents || !agents.length) return null;

  return (
    <div className="cross-agent-panel">
      <div className="cross-agent-header">
        <span>Cross-Agent Trust Comparison</span>
        <button className="close-btn" onClick={onClose} style={{ background: 'transparent', border: 'none', color: '#fff', cursor: 'pointer' }}>X</button>
      </div>

      <div className="cross-agent-body">
        {agents.map(agent => (
          <div key={agent.agentId} className="cross-agent-row">
            <div className="cross-agent-label">{agent.agentId.toUpperCase()}</div>
            <div className="cross-agent-sparkline">
              {agent.trustHistory.map((p, i) => (
                <div
                  key={i}
                  className="cross-trust-point"
                  style={{ height: `${Math.max(10, p.score)}%`, background: p.score < 50 ? '#ff5252' : (p.score < 80 ? '#ffd54f' : '#00ff99') }}
                  title={`Score: ${p.score}`}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
