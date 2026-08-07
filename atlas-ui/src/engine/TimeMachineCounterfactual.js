/**
 * TimeMachineCounterfactual — "What-if" analysis engine.
 * 
 * Evaluates governance rules against historical forecasts in pure read-only mode.
 * No side effects. No POST requests. Only synthetic result generation.
 */
import { GovernanceRules } from './GovernanceRules';
import { compileCondition, compileAction } from './GovernanceRuleCompiler';
import { TimeMachineMiddleware } from './TimeMachineMiddleware';

export const TimeMachineCounterfactual = {

  /**
   * Evaluate a specific rule against all forecasts that existed at a given timestamp.
   * Returns synthetic results showing what WOULD have happened.
   */
  async evaluateRuleAtTime(ruleId, timestamp) {
    const rule = GovernanceRules.getById(ruleId);
    if (!rule) return { error: 'Rule not found', results: [] };

    // Fetch historical state to get forecasts at that time
    const state = await TimeMachineMiddleware.fetchTemporalState(timestamp);
    const forecasts = state.forecasts || [];

    if (forecasts.length === 0) {
      return {
        ruleId,
        timestamp,
        results: [],
        summary: 'No forecasts existed at this timestamp.'
      };
    }

    const condFn = compileCondition(rule.condition);
    if (!condFn) return { error: 'Failed to compile rule condition', results: [] };

    const results = [];

    for (const fc of forecasts) {
      let conditionMet = false;
      try {
        conditionMet = condFn(fc);
      } catch (e) {
        results.push({
          forecast_id: fc.forecast_id,
          agent_id: fc.agent_id,
          conditionMet: false,
          error: e.message,
          wouldHaveFired: false
        });
        continue;
      }

      results.push({
        forecast_id: fc.forecast_id,
        agent_id: fc.agent_id,
        confidence: fc.confidence,
        type: fc.type,
        reasoning: fc.reasoning,
        conditionMet,
        wouldHaveFired: conditionMet,
        syntheticAction: conditionMet ? rule.action : null,
        temporal: true
      });
    }

    const fired = results.filter(r => r.wouldHaveFired).length;

    return {
      ruleId,
      ruleName: rule.id,
      timestamp,
      results,
      summary: `Rule "${rule.id}" would have fired on ${fired}/${forecasts.length} forecasts at ${new Date(timestamp).toLocaleString()}.`
    };
  },

  /**
   * Evaluate ALL enabled rules against historical forecasts at a given timestamp.
   */
  async evaluateAllRulesAtTime(timestamp) {
    const rules = GovernanceRules.getAll().filter(r => r.enabled);
    const outcomes = [];

    for (const rule of rules) {
      const result = await this.evaluateRuleAtTime(rule.id, timestamp);
      outcomes.push(result);
    }

    return {
      timestamp,
      ruleCount: rules.length,
      outcomes,
      summary: `Evaluated ${rules.length} rules against historical state at ${new Date(timestamp).toLocaleString()}.`
    };
  }
};
