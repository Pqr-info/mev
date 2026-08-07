import React, { useState } from "react";
import { GovernanceRules } from "../engine/GovernanceRules";
import GovernanceRuleEditor from "./GovernanceRuleEditor";
import { Shield } from 'lucide-react';

export default function GovernanceRuleList({ operatorId, onClose }) {
  const [selectedRuleId, setSelectedRuleId] = useState(null);
  const rules = GovernanceRules.getAll();

  return (
    <div className="rule-list-panel" style={{ display: 'flex', flexDirection: 'column', height: '100%', background: 'var(--bg-1)', borderLeft: '1px solid var(--border)' }}>
      <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', borderBottom: '1px solid var(--border)' }}>
        <div className="card-title" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Shield size={18} /> Live Rule Introspection
        </div>
        {onClose && (
          <button onClick={onClose} style={{ background: 'transparent', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', fontSize: '1.2rem' }}>×</button>
        )}
      </div>
      
      <div style={{ padding: '1rem', flex: 1, overflowY: 'auto' }}>
        <div className="rule-list" style={{ marginBottom: '1.5rem' }}>
          {rules.map(rule => (
            <div
              key={rule.id}
              className={`rule-row ${selectedRuleId === rule.id ? 'active' : ''}`}
              onClick={() => setSelectedRuleId(rule.id)}
              style={{ 
                background: selectedRuleId === rule.id ? 'rgba(255,255,255,0.05)' : 'transparent',
                borderLeft: `3px solid ${rule.enabled ? 'var(--color-green)' : 'var(--text-secondary)'}`
              }}
            >
              <span style={{ fontWeight: 'bold' }}>{rule.id}</span>
              <span style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>{rule.scope}</span>
              <span style={{ color: rule.enabled ? 'var(--color-green)' : 'var(--text-secondary)' }}>
                {rule.enabled ? "Enabled" : "Disabled"}
              </span>
            </div>
          ))}
        </div>

        {selectedRuleId && (
          <div className="rule-editor-container" style={{ background: 'var(--bg-0)', padding: '1rem', borderRadius: '8px', border: '1px solid var(--border)' }}>
            <GovernanceRuleEditor ruleId={selectedRuleId} operatorId={operatorId} />
          </div>
        )}
      </div>
    </div>
  );
}
