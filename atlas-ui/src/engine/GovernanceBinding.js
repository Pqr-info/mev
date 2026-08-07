import { GovernanceRegistry } from "./GovernanceRegistry.js";

export function bindGovernanceToConstitution(constitution, capabilities) {
  // In a full implementation, the constitution object would validate constraints.
  // For Phase 15.1, we register the capabilities with the GovernanceRegistry to signify ratification.
  GovernanceRegistry.register({
    id: capabilities.id,
    status: "ratified",
    actions: capabilities.actions,
    constraints: capabilities.constraints
  });

  return constitution;
}
