import { GovernanceRules } from "./GovernanceRules";
import { compileCondition, compileAction, validateConstraints } from "./GovernanceRuleCompiler";
import { GovernanceTelemetry } from "./GovernanceTelemetry";

export function applyRuleEdit(ruleId, patch, operatorId = "system") {
  const existing = GovernanceRules.getById(ruleId);
  if (!existing) return { ok: false, error: "Rule not found" };

  const candidate = { ...existing, ...patch };

  if (!validateConstraints(candidate)) {
    return { ok: false, error: "Constraint validation failed. Rule must be reversible, observable, and logged." };
  }

  const conditionFn = compileCondition(candidate.condition);
  const actionFn = compileAction(candidate.action);

  if (!conditionFn || !actionFn) {
    return { ok: false, error: "Compilation failed" };
  }

  // Update rule store
  const updated = GovernanceRules.update(ruleId, candidate);
  recordRuleVersion(updated, operatorId, patch);

  // Emit telemetry about rule change
  GovernanceTelemetry.emit({
    id: `gov-rule-evt-${Date.now()}-${Math.random().toString(36).substr(2, 5)}`,
    type: "GOV_RULE_UPDATED",
    agentId: null,
    clusterId: null,
    timestamp: Date.now(),
    reason: `Rule ${ruleId} updated via live editor`,
    ruleId: ruleId,
    reversible: true
  });

  return { ok: true, rule: updated };
}

function recordRuleVersion(rule, operatorId, patch) {
  const version = {
    timestamp: Date.now(),
    operatorId,
    patch
  };
  rule.versions = rule.versions || [];
  rule.versions.push(version);
}
