export function integrateGovernanceIntoMeshSignature(meshSignature, events) {
  meshSignature.governanceLoad = events.length;
  meshSignature.lastGovernanceEvent = events[events.length - 1] || null;
  return meshSignature;
}
