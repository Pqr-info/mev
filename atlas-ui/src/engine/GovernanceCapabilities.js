export const GovernanceCapabilities = {
  id: "preemptive_auditor_v1",
  actions: [
    "PRE_EMPTIVE_TIGHTEN",
    "PRE_EMPTIVE_QUARANTINE"
  ],
  constraints: {
    reversible: true,
    observable: true,
    logged: true,
    boundedByTrustRange: [0, 100]
  }
};
