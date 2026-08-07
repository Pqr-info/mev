const fetch = globalThis.fetch; // Node 18+

async function propagateMatrix() {
  console.log("=== S27 LPV2 MATRIX PROPAGATION WEAVER ===\\n");
  
  const MAX_ENDPOINT = 'http://localhost:4050/api/lpv2/propagate/from-max';
  
  // Matrix Definition: Regions 10-19, Slots 100000 - 100099
  // To prevent console flooding, we will do a smaller test matrix of 5 slots across 2 regions
  const regions = [10, 11];
  const slots = [100000, 100001, 100002, 100003, 100004];
  
  let successCount = 0;
  
  for (const region of regions) {
    for (const slot of slots) {
      const payload = `[S27CN-LPV2-PAYLOAD] MATRIX-WEAVE-R${region}-S${slot}`;
      
      const reqBody = {
        region,
        slot,
        payload,
        payloadClass: 'intent',
        version: 1
      };
      
      try {
        const res = await fetch(MAX_ENDPOINT, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(reqBody)
        });
        
        if (!res.ok) {
            const errBody = await res.text();
            console.error(`[X] R${region}:S${slot} | Propagation Failed | Status: ${res.status} | ${errBody}`);
            continue;
        }
        
        const data = await res.json();
        console.log(`[OK] R${region}:S${slot} | ID: ${data.identity.padEnd(12)} | Lineage: ${data.lineageId} | SHA256: ${data.checksum.substring(0,8)}...`);
        successCount++;
        
      } catch (err) {
        console.error(`[X] R${region}:S${slot} | Network Error | ${err.message}`);
      }
      
      // Artificial delay to prevent overwhelming the local DMA cache simulation
      await new Promise(resolve => setTimeout(resolve, 50));
    }
  }
  
  console.log(`\\n=== WEAVING COMPLETE ===`);
  console.log(`Successfully propagated ${successCount} out of ${regions.length * slots.length} matrix slots across the mesh.`);
}

propagateMatrix().catch(console.error);
