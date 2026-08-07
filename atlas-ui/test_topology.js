import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import path from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function sleep(ms) {
  return new Promise(r => setTimeout(r, ms));
}

function spawnAgent(name, port, script = 'server.js') {
  console.log(`[Test] Spawning ${name} on port ${port}...`);
  const child = spawn('node', [script], {
    cwd: __dirname,
    env: { ...process.env, PORT: port, AGENT_ID: name, ENDPOINT_URL: `http://localhost:${port}` },
    stdio: 'inherit'
  });
  return child;
}

async function runTests() {
  console.log('--- Starting Topology Test Harness ---');
  
  // Clean slate: reset the database before the test
  console.log('[Test] Wiping database for a clean test run...');
  const fs = (await import('fs')).default;
  const dbPath1 = path.join(__dirname, 's27_mesh_cockroach_sim.db');
  const dbPath2 = path.join(__dirname, 'data', 'pqlite_gmi_mesh.db');
  for (const p of [dbPath1, dbPath2]) {
    try { fs.unlinkSync(p); } catch(e) {}
    try { fs.unlinkSync(p + '-wal'); } catch(e) {}
    try { fs.unlinkSync(p + '-shm'); } catch(e) {}
  }
  console.log('[Test] Wiped databases');

  const zeta = spawnAgent('zeta', 4052, 'zeta_l7_service.js');
  const max = spawnAgent('max', 4050, 'server.js');
  const ted = spawnAgent('ted', 4051, 'server.js');
  const leo = spawnAgent('leo', 4053, 'server.js');
  const brain = spawnAgent('brain', 4060, 'semantic_brain.js');

  const procs = [zeta, max, ted, leo, brain];

  try {
    console.log('[Test] Waiting 5 seconds for agents to boot and register...');
    await sleep(5000);

    // 1. Verify Topology
    console.log('\n--- Test 1: Verify Topology ---');
    const topoRes = await fetch('http://localhost:4052/api/mesh/topology');
    const topoData = await topoRes.json();
    console.log(`[Test] Topology from Zeta:`, topoData.topology.map(t => t.agent_id));
    if (!['max', 'ted', 'leo'].every(a => topoData.topology.find(t => t.agent_id === a))) {
      throw new Error('Topology verification failed. Not all agents registered.');
    }
    console.log('[Test] Topology verification PASSED.');

    // 2. Verify Multi-Peer Propagation
    console.log('\n--- Test 2: Verify Multi-Peer Propagation ---');
    const payload = {
      region: 1,
      slot: 100,
      payload: { instruction: 'test_topology_broadcast' },
      payloadClass: 'test',
      version: 1
    };

    console.log('[Test] Sending propagate request to Max (port 4050)...');
    const propRes = await fetch('http://localhost:4050/api/lpv2/propagate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    
    if (!propRes.ok) throw new Error(`Max propagate failed: ${await propRes.text()}`);
    console.log('[Test] Max propagated successfully. Waiting for Zeta L7 Registry convergence...');
    await sleep(2000);

    const regRes = await fetch('http://localhost:4052/api/lpv2/registry');
    const regData = await regRes.json();
    const lineage = regData.data[0];
    console.log(`[Test] Zeta Registry saw lineage: ${lineage?.lineage_id} from ${lineage?.source_agent}`);
    if (!lineage || lineage.source_agent !== 'max') {
      throw new Error('Registry did not record Max propagation properly.');
    }
    console.log('[Test] Multi-Peer Propagation tracking PASSED.');

    // 3. Verify Trust Evaluation Across All Agents
    console.log('\n--- Test 3: Verify Trust Evaluation ---');
    console.log('[Test] Waiting 12 seconds for Sentinel Governance loop (runs every 10s)...');
    await sleep(12000);

    const trustRes = await fetch('http://localhost:4052/api/governance/trust');
    const trustData = await trustRes.json();
    console.log('[Test] Trust Ledger:');
    trustData.agents.forEach(a => console.log(`  - Agent ${a.agent_id}: Score=${a.trust_score}, Status=${a.status}`));
    
    if (!['max', 'ted', 'leo'].every(a => trustData.agents.find(t => t.agent_id === a))) {
      throw new Error('Trust ledger missing an agent.');
    }
    console.log('[Test] Trust Evaluation PASSED.');

    // 4. Verify L6 Consensus Spine
    console.log('\n--- Test 4: Verify L6 Consensus Spine ---');
    const historyRes = await fetch('http://localhost:4052/api/l6/history/max');
    const historyData = await historyRes.json();
    console.log(`[Test] Max's L6 Spine Length: ${historyData.history.length}`);
    if (historyData.history.length === 0) {
      throw new Error('L6 spine for Max was not populated.');
    }
    const latestRoot = historyData.history[historyData.history.length - 1];
    console.log(`[Test] Latest L6 Root for Max: height ${latestRoot.height}, hash ${latestRoot.root}`);
    
    const verifyRes = await fetch(`http://localhost:4052/api/l6/verify/max/${latestRoot.lineage_id}`);
    const verifyData = await verifyRes.json();
    if (!verifyData.verified) {
      throw new Error('Verification endpoint failed to verify lineage.');
    }
    console.log('[Test] L6 Consensus Spine Validation & Verification PASSED.');

    // 5. Verify Cryptographic Tampering Quarantine
    console.log('\n--- Test 5: Verify Cryptographic Tampering Quarantine ---');
    console.log('[Test] Max is injecting a forged L6 root (rewriting history)...');
    await fetch('http://localhost:4052/api/mesh/auditor/commit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        agent_id: 'max',
        lineage_id: 'lpv2-forgery-1234',
        height: latestRoot.height + 1,
        prev_root: '00000000000badbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbad',
        root: '11111111111badbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbadbad',
        timestamp: Date.now(),
        version: 1
      })
    });
    
    console.log('[Test] Waiting 12 seconds for Sentinel Governance loop to catch tampering...');
    await sleep(12000);
    
    const trustRes2 = await fetch('http://localhost:4052/api/governance/trust');
    const trustData2 = await trustRes2.json();
    const maxTrust = trustData2.agents.find(a => a.agent_id === 'max');
    console.log(`[Test] Max Trust Status post-tamper: Score=${maxTrust.trust_score}, Status=${maxTrust.status}`);
    
    if (maxTrust.status !== 'QUARANTINE') {
      throw new Error(`Sentinel failed to quarantine Max for L6 tampering. Status is: ${maxTrust.status}`);
    }
    console.log('[Test] Cryptographic Tampering Quarantine PASSED.');

    // 6. Verify Semantic Governance Ratification
    console.log('\n--- Test 6: Verify Autonomous Semantic Governance Ratification ---');
    console.log('[Test] Waiting 20 seconds for Semantic Brain to analyze telemetry and propose quarantine relaxation...');
    await sleep(20000);
    
    const trustRes3 = await fetch('http://localhost:4052/api/governance/trust');
    const trustData3 = await trustRes3.json();
    const maxTrustFinal = trustData3.agents.find(a => a.agent_id === 'max');
    console.log(`[Test] Max Trust Status post-ratification: Score=${maxTrustFinal.trust_score}, Status=${maxTrustFinal.status}`);
    
    if (maxTrustFinal.status !== 'PROBATION') {
      throw new Error(`Zeta Ratification Engine failed to accept Semantic Brain proposal. Status is: ${maxTrustFinal.status}`);
    }
    console.log('[Test] Autonomous Semantic Governance Ratification PASSED.');

    console.log('\n--- ALL TESTS PASSED ---\n');

    console.log('--- The Story of the Mesh (Phase 9 Timeline) ---');
    const timelineRes = await fetch('http://localhost:4052/api/governance/timeline');
    const timelineData = await timelineRes.json();
    if (timelineData.ok) {
      timelineData.timeline.forEach(evt => {
        const time = new Date(evt.timestamp).toISOString();
        console.log(`[${time}] [${evt.event_type}] Agent: ${evt.agent_id} | ${evt.description}`);
      });
    }
    console.log('------------------------------------------------\n');

  } catch (err) {
    console.error(`\n[Test] FAILED: ${err.message}\n`);
  } finally {
    console.log('\n[Test] Shutting down agents...');
    procs.forEach(p => p.kill());
  }
}

runTests();
