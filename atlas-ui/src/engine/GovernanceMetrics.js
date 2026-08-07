export function computeGovernanceMetrics(events) {
  const total = events.length;
  const byType = {};
  const byAgent = {};

  events.forEach(evt => {
    byType[evt.type] = (byType[evt.type] || 0) + 1;
    if (evt.agentId) {
      byAgent[evt.agentId] = (byAgent[evt.agentId] || 0) + 1;
    }
  });

  return { total, byType, byAgent };
}
