const fetch = globalThis.fetch; // Node 18+

async function runParityTest() {
  console.log("=== S27 MEMORY PARITY TEST ===");

  const maxUrl = 'http://localhost:4050/api/gmi/shadow';
  const tedUrl = 'http://localhost:4051/api/gmi/shadow';
  
  const payloadStr = "S27-MEM-PARITY-TEST";

  // 1. Write Cascade Parity Test
  console.log("\\n--- 1. Write Cascade Parity ---");
  
  const maxWrite = await fetch(`${maxUrl}/page`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ agentId: 'max', region: 7, slot: 123456, payload: payloadStr })
  }).then(r => r.json());
  
  const tedWrite = await fetch(`${tedUrl}/page`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ agentId: 'ted', region: 7, slot: 123456, payload: payloadStr })
  }).then(r => r.json());
  
  console.log("Max Write:", maxWrite);
  console.log("Ted Write:", tedWrite);
  
  // 2. Intelligent Read Parity Test
  console.log("\\n--- 2. Intelligent Read Parity ---");
  
  const maxRead = await fetch(`${maxUrl}/page?agentId=max&region=7&slot=123456`).then(r => r.json());
  const tedRead = await fetch(`${tedUrl}/page?agentId=ted&region=7&slot=123456`).then(r => r.json());
  
  console.log("Max Read:", maxRead);
  console.log("Ted Read:", tedRead);

  // 3. Telemetry Parity Test
  console.log("\\n--- 3. Telemetry Parity ---");
  const maxStatus = await fetch(`${maxUrl}/tier-status`).then(r => r.json());
  const tedStatus = await fetch(`${tedUrl}/tier-status`).then(r => r.json());
  
  console.log("Max Telemetry:", JSON.stringify(maxStatus, null, 2));
  console.log("Ted Telemetry:", JSON.stringify(tedStatus, null, 2));

  // 4. Zeta L7 Parity Reporting
  console.log("\n--- 4. Zeta L7 Parity Reporting ---");
  const isParityOk = maxRead.payload === tedRead.payload;

  await fetch('http://localhost:4052/api/memory/parity/record', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      agent_id: 'max',
      region: 7,
      slot: 123456,
      l1_alloc_mb: 10,
      l2_state: 'sync',
      l3_state: 'sync',
      l4_state: 'sync',
      l5_state: 'sync',
      l6_state: 'sync',
      parity_ok: isParityOk
    })
  }).then(r => r.json()).then(r => console.log('Zeta Parity Record (Max):', r)).catch(e => console.error(e.message));

  await fetch('http://localhost:4052/api/memory/parity/record', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      agent_id: 'ted',
      region: 7,
      slot: 123456,
      l1_alloc_mb: 10,
      l2_state: 'sync',
      l3_state: 'sync',
      l4_state: 'sync',
      l5_state: 'sync',
      l6_state: 'sync',
      parity_ok: isParityOk
    })
  }).then(r => r.json()).then(r => console.log('Zeta Parity Record (Ted):', r)).catch(e => console.error(e.message));
}

runParityTest().catch(console.error);
