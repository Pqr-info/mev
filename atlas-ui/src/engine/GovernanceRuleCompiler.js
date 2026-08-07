export function compileCondition(conditionSource) {
  try {
    // We pass `fc` (forecast) to the condition evaluator
    // eslint-disable-next-line no-new-func
    const fn = new Function("fc", `return (${conditionSource});`);
    return fn;
  } catch (e) {
    return null;
  }
}

export function compileAction(actionSource) {
  try {
    // We pass `fc`, `quarantine`, `tighten` functions for the action evaluator to call
    // eslint-disable-next-line no-new-func
    const fn = new Function("fc", "quarantine", "tighten", actionSource);
    return fn;
  } catch (e) {
    return null;
  }
}

export function validateConstraints(rule) {
  const c = rule.constraints || {};
  if (!c.reversible || !c.observable || !c.logged) {
    return false;
  }
  return true;
}
