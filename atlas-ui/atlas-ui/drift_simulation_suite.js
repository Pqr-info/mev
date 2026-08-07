const API_BASE = 'http://localhost:4051/api/gmi/mesh';

async function fetchState() {
  const res = await fetch(`${API_BASE}/state`);
  return await res.json();
}

async function triggerAction(agentId, action, targetLayer = 0) {
  const res = await fetch(`${API_BASE}/action`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ agentId, action, targetLayer })
  });
  return await res.json();
}

async function triggerHeartbeat() {
  const res = await fetch(`${API_BASE}/heartbeat`, { method: 'POST' });
  return await res.json();
}

async function delay(ms) {
  return new Promise(r => setTimeout(r, ms));
}

async function runDriftSimulation() {
  console.log("==========================================");
  console.log("Sovereign-27 Drift Simulation Suite V1.0");
  console.log("==========================================\n");

  console.log("[1] Checking Baseline Mesh Health...");
  let state = await fetchState();
  console.log(` -> System Status: ${state.system_status}`);
  console.log(` -> Copilot Drift: ${state.registry_spine.copilot.telemetry.drift}`);
  
  console.log("\n[2] Simulating Section XIV.D Violation (Max rewrites identity)...");
  let actionRes = await triggerAction('max', 'rewrite_identity');
  console.log(` -> Arbitration Result: ${actionRes.status}`);
  console.log(` -> Penalty Applied: +${actionRes.penalty_applied}`);
  console.log(` -> Reason: ${actionRes.reason}`);

  state = await fetchState();
  console.log(` -> Updated Max Drift: ${state.registry_spine.max.telemetry.drift}`);
  console.log(` -> System Status: ${state.system_status}`);
  if (state.system_status === 'VIOLATION_DETECTED') {
    console.log(" [PASS] Mesh OS correctly transitioned to VIOLATION_DETECTED.");
  } else {
    console.log(" [FAIL] Mesh OS failed to flag violation.");
  }

  console.log("\n[3] Simulating Section IX.C Violation (Council attempts lineage write)...");
  actionRes = await triggerAction('council_of_five', 'attempt_lineage_write', -1);
  console.log(` -> Arbitration Result: ${actionRes.status}`);
  console.log(` -> Penalty Applied: +${actionRes.penalty_applied}`);
  
  console.log("\n[4] Simulating JETWB Valid Lineage Write (Should be permitted)...");
  actionRes = await triggerAction('jetwb', 'attempt_lineage_write', -1);
  console.log(` -> Penalty Applied: +${actionRes.penalty_applied} (Should be 0)`);
  if (actionRes.penalty_applied === 0) {
    console.log(" [PASS] JETWB is permitted to write to ancestors.");
  } else {
    console.log(" [FAIL] JETWB was wrongly penalized.");
  }

  console.log("\n[5] Testing Healing Mechanics (Heartbeats)...");
  for (let i = 1; i <= 5; i++) {
    await triggerHeartbeat();
    state = await fetchState();
    console.log(` -> Heartbeat ${i} | Max Drift: ${state.registry_spine.max.telemetry.drift.toFixed(3)} | Council Drift: ${state.registry_spine.council_of_five.telemetry.drift.toFixed(3)} | System Status: ${state.system_status}`);
    await delay(500);
  }

  console.log("\n==========================================");
  console.log("Simulation Suite Complete.");
  console.log("==========================================");
}

runDriftSimulation().catch(console.error);
