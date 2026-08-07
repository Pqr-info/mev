async function run() {
  console.log('[Test] Initiating Advanced 6-Tier Memory Allocation...');

  // 1. Write Page (Will cascade to all 6 tiers)
  const postRes = await fetch('http://localhost:4050/api/gmi/shadow/page', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      agentId: 'max',
      region: 7,
      slot: 123456,
      payload: { memoryType: 'shadow', simulatedDma: true, message: 'Tiering Test' }
    })
  });
  const postJson = await postRes.json();
  console.log('POST /shadow/page Response:', postJson);

  // 2. Read Page (Should hit L2 primarily in this simulation since L1 doesn't have an exact offset mapping)
  console.log('\n[Test] Retrieving Page from Architecture...');
  const getRes = await fetch('http://localhost:4050/api/gmi/shadow/page?agentId=max&region=7&slot=123456');
  const getJson = await getRes.json();
  console.log('GET /shadow/page Response:', getJson);

  // 3. Get Tier Status
  console.log('\n[Test] Checking Tier Status...');
  const statusRes = await fetch('http://localhost:4050/api/gmi/shadow/tier-status');
  const statusJson = await statusRes.json();
  console.log('GET /shadow/tier-status Response:', JSON.stringify(statusJson, null, 2));
}

run().catch(console.error);
