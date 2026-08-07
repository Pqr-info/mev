const NODES = {
  Max: 'http://localhost:4050/api/gmi/mesh',
  Ted: 'http://localhost:4051/api/gmi/mesh'
};

async function fetchState(node) {
  const res = await fetch(`${NODES[node]}/state`);
  return await res.json();
}

async function triggerAction(node, agentId, action, targetLayer = 0) {
  const res = await fetch(`${NODES[node]}/action`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ agentId, action, targetLayer })
  });
  return await res.json();
}

async function triggerHeartbeat(node) {
  const res = await fetch(`${NODES[node]}/heartbeat`, { method: 'POST' });
  return await res.json();
}

async function delay(ms) {
  return new Promise(r => setTimeout(r, ms));
}

async function runDriftSimulation() {
  console.log("==========================================");
  console.log("Sovereign-27 Drift Simulation Suite V2.0 (Multi-Node)");
  console.log("==========================================\n");

  console.log("[1] Checking Baseline Mesh Health on both nodes...");
  const maxState = await fetchState('Max');
  const tedState = await fetchState('Ted');
  
  console.log(` [Max] System Status: ${maxState.system_status} | Copilot Drift: ${maxState.registry_spine.copilot.telemetry.drift}`);
  console.log(` [Ted] System Status: ${tedState.system_status} | Copilot Drift: ${tedState.registry_spine.copilot.telemetry.drift}`);
  
  console.log("\n[2] Simulating Section XIV.D Violation on Max (Max rewrites identity)...");
  let actionRes = await triggerAction('Max', 'max', 'rewrite_identity');
  console.log(` -> [Max] Arbitration Result: ${actionRes.status}`);
  console.log(` -> [Max] Penalty Applied: +${actionRes.penalty_applied}`);

  console.log("\n[3] Simulating Section IX.C Violation on Ted (Council attempts lineage write)...");
  let actionResTed = await triggerAction('Ted', 'council_of_five', 'attempt_lineage_write', -1);
  console.log(` -> [Ted] Arbitration Result: ${actionResTed.status}`);
  console.log(` -> [Ted] Penalty Applied: +${actionResTed.penalty_applied}`);

  console.log("\n[4] Validating states post-violation...");
  const maxStateUpdated = await fetchState('Max');
  const tedStateUpdated = await fetchState('Ted');
  
  console.log(` [Max] Updated Max Drift: ${maxStateUpdated.registry_spine.max.telemetry.drift}`);
  console.log(` [Ted] Updated Council Drift: ${tedStateUpdated.registry_spine.council_of_five.telemetry.drift}`);
  
  if (maxStateUpdated.system_status === 'VIOLATION_DETECTED' && tedStateUpdated.system_status === 'VIOLATION_DETECTED') {
    console.log(" [PASS] Both nodes correctly transitioned to VIOLATION_DETECTED.");
  } else {
    console.log(" [FAIL] Nodes failed to flag violations correctly.");
  }

  console.log("\n[5] Testing Healing Mechanics (Heartbeats)...");
  for (let i = 1; i <= 5; i++) {
    await triggerHeartbeat('Max');
    await triggerHeartbeat('Ted');
    let mState = await fetchState('Max');
    let tState = await fetchState('Ted');
    
    console.log(` -> Heartbeat ${i}`);
    console.log(`    [Max] Drift: ${mState.registry_spine.max.telemetry.drift.toFixed(3)} | Status: ${mState.system_status}`);
    console.log(`    [Ted] Drift: ${tState.registry_spine.council_of_five.telemetry.drift.toFixed(3)} | Status: ${tState.system_status}`);
    await delay(500);
  }

  console.log("\n==========================================");
  console.log("Multi-Node Simulation Suite Complete.");
  console.log("==========================================");
}

runDriftSimulation().catch(console.error);

