const fetch = globalThis.fetch;

async function runTrustTest() {
  console.log("=== S27 GOVERNANCE TRUST & QUARANTINE ===");

  console.log("\\n1. Checking Initial Trust State...");
  const initialTrust = await fetch('http://localhost:4052/api/governance/trust').then(r => r.json());
  console.log("Initial Trust Ledger:", JSON.stringify(initialTrust.agents, null, 2));

  console.log("\\n2. Injecting 3 synthetic drift anomalies for 'max' to crater trust score...");
  for (let i = 0; i < 3; i++) {
      await fetch('http://localhost:4052/api/lpv2/drift/record', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            lineage_id: `lpv2-synthetic-drift-${Date.now()}-${i}`,
            source_agent: 'max',
            max_version: 1,
            ted_version: 1,
            max_checksum: `bad-${i}`,
            ted_checksum: `good-${i}`,
            drift_detected: true
          })
      });
  }
  
  console.log("\\n3. Waiting 12 seconds for Zeta Sentinel Loop to compute trust and orchestrate quarantine...");
  await new Promise(r => setTimeout(r, 12000));

  console.log("\\n4. Checking Post-Evaluation Trust State...");
  const postTrust = await fetch('http://localhost:4052/api/governance/trust').then(r => r.json());
  console.log("Post Trust Ledger:", JSON.stringify(postTrust.agents, null, 2));
  
  const maxTrust = postTrust.agents.find(a => a.agent_id === 'max');
  if (maxTrust && maxTrust.status === 'QUARANTINE') {
      console.log("✅ Zeta successfully pushed Max into QUARANTINE.");
  } else {
      console.log("❌ Failed to quarantine Max.");
  }

  console.log("\\n5. Writing to Max to verify structural tier lockdown (L2, L3, L4, L6 disabled)...");
  const writeResMax = await fetch('http://localhost:4050/api/gmi/shadow/page', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      agentId: "max",
      region: 99,
      slot: 99999,
      payload: { msg: "Quarantine Write Test" }
    })
  }).then(r => r.json());
  
  console.log("Max Write Tiers:", writeResMax.tiersWritten);
  
  const hasPersistence = writeResMax.tiersWritten.some(t => ['L2', 'L3', 'L4', 'L6'].includes(t));
  if (!hasPersistence) {
      console.log("✅ SUCCESS: Max is fully quarantined. Persistence tiers structurally locked.");
  } else {
      console.log("❌ FAILED: Persistence tiers were still written to.");
  }
}

runTrustTest().catch(console.error);
