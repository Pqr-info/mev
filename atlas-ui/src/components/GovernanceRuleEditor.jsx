import React, { useState, useEffect } from "react";
import { GovernanceRules } from "../engine/GovernanceRules";
import { applyRuleEdit } from "../engine/GovernanceRuleApplier";

export default function GovernanceRuleEditor({ ruleId, operatorId = "human-operator" }) {
  const [rule, setRule] = useState(GovernanceRules.getById(ruleId));
  const [condition, setCondition] = useState("");
  const [action, setAction] = useState("");
  const [enabled, setEnabled] = useState(true);
  const [status, setStatus] = useState(null);

  useEffect(() => {
    const r = GovernanceRules.getById(ruleId);
    setRule(r);
    if (r) {
      setCondition(r.condition || "");
      setAction(r.action || "");
      setEnabled(r.enabled ?? true);
      setStatus(null);
    }
  }, [ruleId]);

  if (!rule) return <div style={{ color: 'var(--text-secondary)' }}>Select a rule to edit.</div>;

  const onSave = () => {
    const patch = { condition, action, enabled };
    const result = applyRuleEdit(ruleId, patch, operatorId);
    if (result.ok) {
      setStatus({ text: "Saved successfully", type: "success" });
      setRule(GovernanceRules.getById(ruleId)); // Refresh
    } else {
      setStatus({ text: `Error: ${result.error}`, type: "error" });
    }
  };

  return (
    <div className="rule-editor">
      <div className="rule-header" style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '12px' }}>
        <span style={{ fontWeight: 'bold', fontSize: '1.1rem' }}>{rule.id}</span>
        <span className="rule-provenance tag" style={{ background: 'rgba(255,255,255,0.1)', padding: '2px 6px', borderRadius: '4px' }}>{rule.provenance}</span>
      </div>

      <div className="rule-field">
        <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '4px', color: 'var(--text-secondary)' }}>Condition (JS Expression)</label>
        <textarea
          value={condition}
          onChange={e => setCondition(e.target.value)}
        />
      </div>

      <div className="rule-field">
        <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: '4px', color: 'var(--text-secondary)' }}>Action (JS Function Call)</label>
        <textarea
          value={action}
          onChange={e => setAction(e.target.value)}
        />
      </div>

      <div className="rule-field" style={{ margin: '12px 0' }}>
        <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={enabled}
            onChange={e => setEnabled(e.target.checked)}
            style={{ width: '16px', height: '16px', accentColor: 'var(--color-blue)' }}
          />
          Enabled
        </label>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
        <button 
          onClick={onSave}
          style={{ background: 'var(--color-blue)', color: '#fff', border: 'none', padding: '6px 16px', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold' }}
        >
          Apply Rule
        </button>
        {status && (
          <div className="rule-status" style={{ color: status.type === 'success' ? 'var(--color-green)' : 'var(--color-red)', fontSize: '0.85rem' }}>
            {status.text}
          </div>
        )}
      </div>

      <div className="rule-versions" style={{ marginTop: '20px', borderTop: '1px solid rgba(255,255,255,0.1)', paddingTop: '10px' }}>
        <div className="section-title" style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', marginBottom: '8px' }}>Version History</div>
        {(rule.versions || []).slice().reverse().map((v, i) => (
          <div key={i} className="rule-version-row" style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
            <span>{new Date(v.timestamp).toLocaleString()}</span>
            <span>Op: {v.operatorId}</span>
          </div>
        ))}
        {!(rule.versions && rule.versions.length) && <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>No edits yet.</div>}
      </div>
    </div>
  );
}
