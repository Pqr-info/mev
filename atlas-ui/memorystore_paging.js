import Redis from 'ioredis';

const redis = new Redis(process.env.VALKEY_ADDR || '127.0.0.1:6379', {
  enableOfflineQueue: false,
  maxRetriesPerRequest: 1
});
redis.on('error', (err) => console.error('[Shadow Paging Redis] Connection error:', err.message));

// The 8Mx2M topology implies 8 regions/partitions and 2M slots.
// We will use lazy allocation for memory efficiency. 
// A typical key: mesh:shadow:page:{region}:{slot}:{agentId}
const MAX_REGION = 8;
const MAX_SLOT = 2000000;

/**
 * Validates the bucket coordinates against the 8Mx2M bounds
 */
function validateBucket(region, slot) {
  const r = parseInt(region, 10);
  const s = parseInt(slot, 10);
  
  if (isNaN(r) || isNaN(s)) return false;
  if (r < 0 || r >= MAX_REGION) return false;
  if (s < 0 || s >= MAX_SLOT) return false;
  
  return true;
}

/**
 * Set a shadow memory page
 * @param {string} agentId - The agent ID
 * @param {number} region - The memory region (0 to 7)
 * @param {number} slot - The memory slot (0 to 1999999)
 * @param {object} payload - The memory data payload
 */
export async function setShadowPage(agentId, region, slot, payload) {
  if (!validateBucket(region, slot)) {
    throw new Error(`Invalid bucket bounds for 8Mx2M topology. Region must be 0-7, slot must be 0-1999999.`);
  }

  const key = `mesh:shadow:page:${region}:${slot}:${agentId}`;
  
  // Storing the payload as a JSON string in a Redis String key
  // You could also use Redis Hashes, but simple SET/GET is faster for raw page dumps
  await redis.set(key, JSON.stringify({
    timestamp: Date.now(),
    agentId,
    region,
    slot,
    data: payload
  }));

  return { ok: true, key };
}

/**
 * Get a shadow memory page
 * @param {string} agentId - The agent ID
 * @param {number} region - The memory region (0 to 7)
 * @param {number} slot - The memory slot (0 to 1999999)
 */
export async function getShadowPage(agentId, region, slot) {
  if (!validateBucket(region, slot)) {
    throw new Error(`Invalid bucket bounds for 8Mx2M topology. Region must be 0-7, slot must be 0-1999999.`);
  }

  const key = `mesh:shadow:page:${region}:${slot}:${agentId}`;
  const data = await redis.get(key);
  
  if (!data) return null;
  return JSON.parse(data);
}
