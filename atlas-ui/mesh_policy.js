// Mesh Policy Engine configuration
// This module provides a global runtime state for policies pushed down from Zeta.

export const ActivePolicy = {
    disabledTiers: [],
    quarantinedAgents: [],
    policy_id: null,
    version: null,
    appliedAt: null
};

// Expose a helper to cleanly update policy state
export function updatePolicy(newPolicy) {
    if (newPolicy.disabledTiers && Array.isArray(newPolicy.disabledTiers)) {
        ActivePolicy.disabledTiers = newPolicy.disabledTiers;
    }
    if (newPolicy.quarantinedAgents && Array.isArray(newPolicy.quarantinedAgents)) {
        ActivePolicy.quarantinedAgents = newPolicy.quarantinedAgents;
    }
    ActivePolicy.policy_id = newPolicy.policy_id || null;
    ActivePolicy.version = newPolicy.version || 1;
    ActivePolicy.appliedAt = new Date().toISOString();
    console.log(`[Mesh Policy Engine] Active policy updated:`, ActivePolicy);
    return { policy_id: ActivePolicy.policy_id, version: ActivePolicy.version, appliedAt: ActivePolicy.appliedAt };
}
