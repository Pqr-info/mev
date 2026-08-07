export const GovernanceRegistry = {
  subsystems: new Map(),

  register(subsystem) {
    this.subsystems.set(subsystem.id, subsystem);
  },

  isRatified(id) {
    return this.subsystems.has(id);
  }
};
