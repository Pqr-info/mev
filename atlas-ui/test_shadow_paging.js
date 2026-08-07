async function run() {
  console.log('[Test] Setting shadow page...');
  const postRes = await fetch('http://localhost:4050/api/gmi/shadow/page', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      agentId: 'max',
      region: 7,
      slot: 1999999,
      payload: { memoryType: 'shadow', test: true }
    })
  });
  const postJson = await postRes.json();
  console.log('POST Response:', postJson);

  console.log('[Test] Getting shadow page...');
  const getRes = await fetch('http://localhost:4050/api/gmi/shadow/page?agentId=max&region=7&slot=1999999');
  const getJson = await getRes.json();
  console.log('GET Response:', getJson);
}

run().catch(console.error);
