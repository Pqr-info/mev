const fetch = globalThis.fetch;

async function runPolicyTest() {
  console.log("=== S27 GOVERNANCE POLICY ORCHESTRATION ===");

  console.log("\\n1. Pushing Global Policy to Zeta: Disable L3 (Redis) and L4 (Swap)...");
  
  const policyRes = await fetch('http://localhost:4052/api/governance/policy', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      type: "MEMORY_RESTRICTION",
      disabledTiers: ["L3", "L4"]
    })
  }).then(r => r.json());

  console.log("Zeta Policy Response:", policyRes);

  console.log("\\n2. Waiting 1 second for policy broadcast to propagate to agents...");
  await new Promise(r => setTimeout(r, 1000));

  console.log("\\n3. Running Parity Verification (Writing data to Max and Ted)...");
  
  // We'll write data to Max and see which tiers it hits.
  const writeResMax = await fetch('http://localhost:4050/api/gmi/shadow/page', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      agentId: "max",
      region: 42,
      slot: 12345,
      payload: { msg: "Testing Policy Orchestration" }
    })
  }).then(r => r.json());
  
  console.log("Max Write Tiers:", writeResMax.tiersWritten);

  // Read it back
  const readResMax = await fetch('http://localhost:4050/api/gmi/shadow/page?agentId=max&region=42&slot=12345').then(r => r.json());
  console.log("Max Read Response:", readResMax);

  console.log("\\n4. Checking Telemetry for Policy Enforcement...");
  
  const maxTelemetry = await fetch('http://localhost:4050/api/gmi/shadow/tier-status').then(r => r.json());
  console.log("Max L3 Status:", maxTelemetry.l3_redis.state);
  console.log("Max L4 Status:", maxTelemetry.l4_swap.state);
  
  if (maxTelemetry.l3_redis.state === 'disabled_by_policy' && maxTelemetry.l4_swap.state === 'disabled_by_policy') {
      console.log("\\n✅ SUCCESS: Zeta successfully orchestrated global memory policy down to Max!");
  } else {
      console.log("\\n❌ FAILED: Policy not enforced at the edge.");
  }
}

runPolicyTest().catch(console.error);
