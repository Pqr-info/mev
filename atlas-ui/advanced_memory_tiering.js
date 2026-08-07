import fs from 'fs/promises';
import path from 'path';
import crypto from 'crypto';
import Redis from 'ioredis';
import zlib from 'zlib';
import { promisify } from 'util';
import { ActivePolicy } from './mesh_policy.js';

const gzip = promisify(zlib.gzip);
const gunzip = promisify(zlib.gunzip);

const redis = new Redis(process.env.VALKEY_ADDR || '127.0.0.1:6379', {
  enableOfflineQueue: false,
  maxRetriesPerRequest: 1
});

redis.on('error', (err) => console.error('[L3 Redis] Connection error:', err.message));

const DATA_DIR = path.join(process.cwd(), 'data');
const DIRS = {
  L2: path.join(DATA_DIR, 'ssd'),
  L4: path.join(DATA_DIR, 'swap'),
  L5: path.join(DATA_DIR, 'local'),
  L6: path.join(DATA_DIR, 'cloud')
};

// Ensure directories exist asynchronously
(async () => {
  for (const dir of Object.values(DIRS)) {
    await fs.mkdir(dir, { recursive: true });
  }
})();

// L1 DMA Buffer (16MB per agent)
// Structure: Map<agentId, { buffer: Buffer, lastAccessed: number }>
const L1_CACHE = new Map();
const L1_SIZE = 16 * 1024 * 1024; // 16MB

function getL1Buffer(agentId) {
  if (!L1_CACHE.has(agentId)) {
    L1_CACHE.set(agentId, {
      buffer: Buffer.alloc(L1_SIZE),
      lastAccessed: Date.now()
    });
  }
  const entry = L1_CACHE.get(agentId);
  entry.lastAccessed = Date.now();
  return entry.buffer;
}

export async function writeShadowPage(agentId, region, slot, data) {
  const payloadStr = JSON.stringify(data);
  const payloadBuffer = Buffer.from(payloadStr, 'utf8');
  const tiersWritten = [];
  const isQuarantined = (ActivePolicy.quarantinedAgents || []).includes(agentId);
  
  // 1. L1: Active MemoryStore DMA (16MB per agent)
  if (!ActivePolicy.disabledTiers.includes('L1')) {
    const l1 = getL1Buffer(agentId);
    if (payloadBuffer.length <= L1_SIZE) {
      payloadBuffer.copy(l1, 0); // Simulate DMA load
      tiersWritten.push('L1');
    }
  }

  // 2. L2: SSD Zerocopy
  if (!ActivePolicy.disabledTiers.includes('L2') && !isQuarantined) {
    const l2File = path.join(DIRS.L2, `${agentId}_${region}_${slot}.img`);
    await fs.writeFile(l2File, payloadBuffer);
    tiersWritten.push('L2');
  }

  // 3. L3: rqlite-Redis
  if (!ActivePolicy.disabledTiers.includes('L3') && !isQuarantined) {
    const l3Key = `mesh:l3:${agentId}:${region}:${slot}`;
    try {
      await redis.set(l3Key, payloadStr);
      tiersWritten.push('L3');
    } catch (e) {
      // Graceful fail if L3 offline
    }
  }

  // 4. L4: Swap / Pagefile
  if (!ActivePolicy.disabledTiers.includes('L4') && !isQuarantined) {
    const compressed = await gzip(payloadBuffer);
    const l4File = path.join(DIRS.L4, `${agentId}_${region}_${slot}.gz`);
    await fs.writeFile(l4File, compressed);
    tiersWritten.push('L4');
  }

  // 5. L5: LocalStorage
  if (!ActivePolicy.disabledTiers.includes('L5')) {
    const l5File = path.join(DIRS.L5, `${agentId}_${region}_${slot}.json`);
    await fs.writeFile(l5File, JSON.stringify({ data: payloadStr, tier: 'L5' }));
    tiersWritten.push('L5');
  }

  // 6. L6: NFS / Cloud (Now a Cryptographic Consensus Spine)
  if (!ActivePolicy.disabledTiers.includes('L6') && !isQuarantined) {
    const l6Dir = DIRS.L6;
    const spineFile = path.join(l6Dir, `${agentId}_spine.json`);
    
    // Load local L6 ledger state
    let height = 1;
    let prev_root = '0000000000000000000000000000000000000000000000000000000000000000';
    try {
      const spineData = await fs.readFile(spineFile, 'utf8');
      const spine = JSON.parse(spineData);
      height = spine.height + 1;
      prev_root = spine.current_root;
    } catch (e) {
      // Genesis sequence
    }

    const timestamp = Date.now();
    // Use lineage_id if available from data, else generate one
    const lineage_id = data.lineage_id || `lpv2-${timestamp}-${agentId}`;
    
    // Compute deterministic hash of the new state payload
    const dataHash = crypto.createHash('sha256').update(`${agentId}:${region}:${slot}:${payloadStr}:${timestamp}`).digest('hex');
    
    // Compute the new Merkle root spanning history + new state
    const root = crypto.createHash('sha256').update(`${prev_root}:${dataHash}`).digest('hex');

    // Save payload to L6
    const l6File = path.join(DIRS.L6, `${agentId}_${region}_${slot}.cloud`);
    await fs.writeFile(l6File, payloadBuffer);
    
    // Update local spine ledger
    await fs.writeFile(spineFile, JSON.stringify({ height, current_root: root }));

    // Emit to Zeta L7 Auditor (STATE_COMMIT event)
    try {
      const topoRes = await fetch('http://localhost:4052/api/mesh/topology');
      if (topoRes.ok) {
        // Just directly call Zeta on 4052 as the auditor endpoint is there
        await fetch('http://localhost:4052/api/mesh/auditor/commit', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            agent_id: agentId,
            lineage_id,
            height,
            prev_root,
            root,
            version: data.version || 1,
            timestamp
          })
        });
      }
    } catch (e) {
      console.warn(`[L6 Auditor] Failed to emit STATE_COMMIT to Zeta:`, e.message);
    }
    
    tiersWritten.push('L6');
  }

  return { ok: true, agentId, tiersWritten };
}

export async function readShadowPage(agentId, region, slot) {
  // Cascading tier read simulation (L1 -> L6)
  
  // L1 & L2 (Primary fast paths)
  if (!ActivePolicy.disabledTiers.includes('L2')) {
    const l2File = path.join(DIRS.L2, `${agentId}_${region}_${slot}.img`);
    try {
      const data = await fs.readFile(l2File, 'utf8');
      if (!ActivePolicy.disabledTiers.includes('L1')) {
        getL1Buffer(agentId); // Touch L1 DMA
      }
      return { ok: true, tier: 'L2', data: JSON.parse(data) };
    } catch (e) {}
  }

  // L3 (Redis)
  if (!ActivePolicy.disabledTiers.includes('L3')) {
    const l3Key = `mesh:l3:${agentId}:${region}:${slot}`;
    try {
      const data = await redis.get(l3Key);
      if (data) return { ok: true, tier: 'L3', data: JSON.parse(data) };
    } catch(e) {}
  }

  // L4 (Swap)
  if (!ActivePolicy.disabledTiers.includes('L4')) {
    const l4File = path.join(DIRS.L4, `${agentId}_${region}_${slot}.gz`);
    try {
      const comp = await fs.readFile(l4File);
      const raw = await gunzip(comp);
      return { ok: true, tier: 'L4', data: JSON.parse(raw.toString('utf8')) };
    } catch(e) {}
  }

  // L5 (Local)
  if (!ActivePolicy.disabledTiers.includes('L5')) {
    const l5File = path.join(DIRS.L5, `${agentId}_${region}_${slot}.json`);
    try {
      const data = await fs.readFile(l5File, 'utf8');
      const parsed = JSON.parse(data);
      return { ok: true, tier: 'L5', data: JSON.parse(parsed.data) };
    } catch(e) {}
  }
  
  // L6 (Cloud)
  if (!ActivePolicy.disabledTiers.includes('L6')) {
    const l6File = path.join(DIRS.L6, `${agentId}_${region}_${slot}.cloud`);
    try {
      const data = await fs.readFile(l6File, 'utf8');
      return { ok: true, tier: 'L6', data: JSON.parse(data) };
    } catch(e) {}
  }

  return { ok: false, error: 'Page fault across all 6 tiers (or tiers disabled by policy)' };
}

export function getTierStatus() {
  const l1Usage = L1_CACHE.size * 16; // MB
  return {
    policy_enforcement: {
      policy_id: ActivePolicy.policy_id,
      version: ActivePolicy.version,
      appliedAt: ActivePolicy.appliedAt,
      quarantinedAgents: ActivePolicy.quarantinedAgents || []
    },
    l1_dma: { active_agents: L1_CACHE.size, allocated_mb: l1Usage, state: ActivePolicy.disabledTiers.includes('L1') ? 'disabled_by_policy' : 'active' },
    l2_ssd: { state: ActivePolicy.disabledTiers.includes('L2') ? 'disabled_by_policy' : 'active', mount: DIRS.L2 },
    l3_redis: { state: ActivePolicy.disabledTiers.includes('L3') ? 'disabled_by_policy' : (redis.status === 'ready' ? 'connected' : 'offline') },
    l4_swap: { state: ActivePolicy.disabledTiers.includes('L4') ? 'disabled_by_policy' : 'active', mount: DIRS.L4 },
    l5_local: { state: ActivePolicy.disabledTiers.includes('L5') ? 'disabled_by_policy' : 'active', mount: DIRS.L5 },
    l6_cloud: { state: ActivePolicy.disabledTiers.includes('L6') ? 'disabled_by_policy' : 'active', mount: DIRS.L6 }
  };
}
