import { GovernanceTelemetry } from './GovernanceTelemetry';
import { GovernanceRules } from './GovernanceRules';
import { compileCondition, compileAction } from './GovernanceRuleCompiler';
import { TimeMachineContext } from './TimeMachineContext';

export class PreEmptiveAuditor {
  constructor(apiBase) {
    this.apiBase = apiBase;
    this.actionHistory = new Set();
  }

  async scanAndAct(forecasts, enabled = false) {
    if (!enabled || !forecasts || forecasts.length === 0) return [];
    
    const actionsTaken = [];
    const rules = GovernanceRules.getAll().filter(r => r.enabled);
    
    const isTemporal = TimeMachineContext.isTemporalMode();

    // Define the callable actions that rules can trigger.
    // CRITICAL: In temporal mode, these are NO-OP stubs that return synthetic results.
    const quarantine = async (agent_id, reason, forecast_id) => {
      if (isTemporal) {
        return { type: 'PRE_EMPTIVE_QUARANTINE', agent_id, reason, temporal: true };
      }
      try {
        await fetch(`${this.apiBase}/pre-emptive/quarantine`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ agent_id, reason, forecast_id })
        });
        return { type: 'PRE_EMPTIVE_QUARANTINE', agent_id, reason };
      } catch (e) {
        console.error("Failed to execute quarantine:", e);
        return null;
      }
    };

    const tighten = async (agent_id, reason, forecast_id) => {
      if (isTemporal) {
        return { type: 'PRE_EMPTIVE_TIGHTEN', agent_id, reason, temporal: true };
      }
      try {
        await fetch(`${this.apiBase}/pre-emptive/tighten`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ agent_id, reason, forecast_id })
        });
        return { type: 'PRE_EMPTIVE_TIGHTEN', agent_id, reason };
      } catch (e) {
        console.error("Failed to execute tighten trust:", e);
        return null;
      }
    };

    for (const fc of forecasts) {
      if (this.actionHistory.has(fc.forecast_id)) continue;
      
      let actionFired = false;
      
      for (const rule of rules) {
        if (actionFired) break; // Execute at most one rule per forecast for now
        
        const condFn = compileCondition(rule.condition);
        const actionFn = compileAction(rule.action);
        
        if (!condFn || !actionFn) continue;

        let conditionMet = false;
        try {
          conditionMet = condFn(fc);
        } catch(e) {
          console.error("Rule condition evaluation error:", e);
        }

        if (conditionMet) {
          try {
            const resultPromise = actionFn(fc, quarantine, tighten);
            if (resultPromise) {
              const res = await Promise.resolve(resultPromise);
              if (res) {
                this.actionHistory.add(fc.forecast_id);
                const evt = {
                  id: `gov-evt-${Date.now()}-${Math.random().toString(36).substr(2, 5)}`,
                  type: res.type,
                  agentId: res.agent_id,
                  clusterId: null,
                  timestamp: Date.now(),
                  reason: isTemporal ? `[TEMPORAL] Rule: ${res.reason}` : `Live Rule: ${res.reason}`,
                  ruleId: rule.id,
                  reversible: rule.constraints?.reversible ?? false,
                  temporal: isTemporal || false
                };
                GovernanceTelemetry.emit(evt);
                actionsTaken.push(evt);
                actionFired = true;
              }
            }
          } catch(e) {
            console.error("Rule action evaluation error:", e);
          }
        }
      }
    }
    
    return actionsTaken;
  }
}
