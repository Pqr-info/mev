import crypto from 'crypto';

const fetch = globalThis.fetch;

async function runDriftTest() {
  console.log("=== S27 GOVERNANCE DRIFT SIMULATION ===");

  const maxPropagate = 'http://localhost:4050/api/lpv2/propagate/from-max';
  
  const region = 12;
  const slot = 55555;
  const payloadStr = "DRIFT-SIMULATION-PAYLOAD";
  
  console.log("\\n1. Seeding valid lineage into Max and Zeta...");
  const seedRes = await fetch(maxPropagate, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      region,
      slot,
      payload: payloadStr,
      payloadClass: 'state',
      version: 1
    })
  }).then(r => r.json());

  const lineageId = seedRes.lineageId;
  console.log("Seeded Lineage ID:", lineageId);

  console.log("\\n2. Simulating a corrupted LPV2 propagation to Ted...");
  
  const envelope = {
    lineageId: lineageId,
    source: 'max',
    region,
    slot,
    payloadClass: 'state',
    version: 1,
    payload: payloadStr,
    identity: 'MXMAX.N0.L0',
    checksum: 'bad-checksum-deadbeef'
  };

  try {
    const res = await fetch('http://localhost:4051/api/lpv2/receive', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(envelope)
    });
    console.log("Ted response to corrupt envelope:", res.status, await res.text());
  } catch (e) {
    console.error("Ted fetch error:", e.message);
  }

  console.log("\\n3. Waiting for Zeta Governance Engine to detect and repropagate...");
  console.log("Check the Zeta terminal to watch the auto-resolution loop in action!");

  let resolved = false;
  for (let i = 0; i < 6; i++) {
    await new Promise(r => setTimeout(r, 3000)); // check every 3s
    const reports = await fetch('http://localhost:4052/api/lpv2/drift/report').then(r => r.json());
    if (reports.ok && reports.data) {
        console.log("Current drift entries:", reports.data.length);
        const anomaly = reports.data.find(d => d.lineage_id === lineageId);
        if (anomaly) {
            console.log(`Anomaly ${anomaly.lineage_id} drift_detected = ${anomaly.drift_detected}`);
            if (anomaly.drift_detected === 0) {
                console.log("✅ Zeta Auto-Repropagate successfully resolved the drift!");
                resolved = true;
                break;
            }
        }
    }
  }

  if (!resolved) {
      console.log("❌ Governance engine did not resolve the drift in time.");
  }
}

runDriftTest().catch(console.error);
