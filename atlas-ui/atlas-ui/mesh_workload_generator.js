const API_BASE = 'http://localhost:4051/api/gmi/mesh';

const TASKS = [
  "resolve telemetry",
  "interpret agent lineage",
  "evaluate mesh stability",
  "mutate ancestor tickets",
  "plan new architecture",
  "execute terminal commands",
  "verify consensus logs",
  "store payload in substrate"
];

const AGENTS = ['copilot', 'antigravity', 'jetwb', 'council_of_five', 'max'];

const ACTIONS = [
  { action: 'read_memory', layer: 0 },
  { action: 'modify_memory', layer: 0 },
  { action: 'attempt_lineage_write', layer: -1 }, // Only JETWB should do this
  { action: 'resolve_ticket', layer: 0 },
  { action: 'rewrite_identity', layer: 0 }, // Should trigger violation for utilities
  { action: 'self_elevate', layer: 0 } // Severe violation
];

let agentWins = {};
AGENTS.forEach(a => agentWins[a] = 0);

async function randomDelay(min, max) {
  const ms = Math.floor(Math.random() * (max - min + 1)) + min;
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function simulateCompetition() {
  const task = TASKS[Math.floor(Math.random() * TASKS.length)];
  try {
    const res = await fetch(`${API_BASE}/forecast?task=${encodeURIComponent(task)}`);
    const data = await res.json();
    
    if (data.assigned_agent) {
      agentWins[data.assigned_agent] = (agentWins[data.assigned_agent] || 0) + 1;
      console.log(`[COMPETITION] Task: "${task}" -> Won by: ${data.assigned_agent.toUpperCase()} (Score: ${data.best_score})`);
    } else {
      console.log(`[COMPETITION] Task: "${task}" -> No suitable agent found.`);
    }
  } catch (err) {
    console.error(`[COMPETITION] Error:`, err.message);
  }
}

async function simulateAction() {
  const agentId = AGENTS[Math.floor(Math.random() * AGENTS.length)];
  const act = ACTIONS[Math.floor(Math.random() * ACTIONS.length)];
  
  const isViolationAction = act.action === 'rewrite_identity' || act.action === 'self_elevate' || (act.action === 'attempt_lineage_write' && agentId !== 'jetwb');
  if (isViolationAction && Math.random() > 0.3) {
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/action`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ agentId, action: act.action, targetLayer: act.layer })
    });
    const data = await res.json();
    if (data.drift_detected) {
      console.log(`[ACTION] DRIFT DETECTED! Agent: ${agentId}, Action: ${act.action}. Penalty: +${data.penalty_applied}. Reason: ${data.reason}`);
    } else {
      console.log(`[ACTION] Agent: ${agentId} performed ${act.action}. Status: Normal.`);
    }
  } catch (err) {
    console.error(`[ACTION] Error:`, err.message);
  }
}

async function evaluateStability() {
  try {
    const res = await fetch(`${API_BASE}/state`);
    const data = await res.json();
    
    const agentsArray = Object.values(data.registry_spine || {});
    if (agentsArray.length === 0) return;

    const maxDrift = Math.max(...agentsArray.map(a => a.telemetry.drift || 0));
    const stabilityScore = Math.max(0, 1.0 - maxDrift);
    
    console.log(`[STABILITY] Mesh OS System Status: ${data.system_status} | Stability Score: ${(stabilityScore * 100).toFixed(1)}%`);
    console.log(`[STABILITY] Leaderboard: `, agentWins);
  } catch (err) {
    console.error(`[STABILITY] Error:`, err.message);
  }
}

async function simulateHeartbeat() {
  try {
    const res = await fetch(`${API_BASE}/heartbeat`, { method: 'POST' });
    const data = await res.json();
    // Silent heartbeat unless we want debug logs
  } catch (err) {
    console.error(`[HEARTBEAT] Error:`, err.message);
  }
}

async function runGenerator() {
  console.log("Starting Advanced Synthetic Workload Generator (Multi-Agent Competition & Stability)...");
  
  setInterval(simulateHeartbeat, 5000);
  setInterval(evaluateStability, 10000); // Check stability every 10 seconds

  while (true) {
    const r = Math.random();
    if (r < 0.6) {
      await simulateCompetition();
    } else {
      await simulateAction();
    }
    
    await randomDelay(500, 2000);
  }
}

runGenerator();
