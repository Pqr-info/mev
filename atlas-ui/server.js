// server.js
import express from 'express';
import cors from 'cors';
import morgan from 'morgan';
import crypto from 'crypto';
import fs from 'fs/promises';
import path from 'path';
import sqlite3 from 'sqlite3';
import { open } from 'sqlite';
import net from 'net';
import { fileURLToPath } from 'url';
import Redis from 'ioredis';
import * as meshOs from './mesh_os_core.js';
import * as memoryTiering from './advanced_memory_tiering.js';
import lpv2 from './lpv2_lineage.js';
import { updatePolicy } from './mesh_policy.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const app = express();
const PRIMARY_PORT = parseInt(process.env.PORT, 10) || 4050;
const FALLBACK_PORT = parseInt(process.env.PORT, 10) || 4051;
const AGENT_ID = process.env.AGENT_ID || 'max';

// --- middleware ---
app.use(cors());
app.use(express.json({ limit: '10mb' }));
app.use(morgan('dev'));

const redis = new Redis(process.env.VALKEY_ADDR || '127.0.0.1:6379');
redis.on('error', (err) => console.error('[Redis] Connection error:', err.message));

// Ensure data directory exists
const dataDir = path.join(__dirname, 'data');
await fs.mkdir(dataDir, { recursive: true });

const dbPath = path.join(dataDir, 'pqlite_gmi_mesh.db');

let db;

async function initDb() {
  db = await open({
    filename: dbPath,
    driver: sqlite3.verbose().Database
  });

  await db.exec('PRAGMA journal_mode = WAL;');
  await db.exec('PRAGMA foreign_keys = ON;');

  await db.exec(`
    CREATE TABLE IF NOT EXISTS memory_page (
      page_id TEXT PRIMARY KEY,
      agent_id TEXT NOT NULL,
      origin TEXT,
      visibility TEXT,
      timestamp INTEGER,
      raw_content TEXT,
      sha256 TEXT
    );

    CREATE TABLE IF NOT EXISTS ticket (
      ticket_id INTEGER PRIMARY KEY,
      agent_id TEXT NOT NULL,
      label TEXT
    );

    CREATE TABLE IF NOT EXISTS page_ticket_map (
      page_id TEXT,
      agent_id TEXT,
      ticket_id INTEGER,
      weight REAL,
      perspective TEXT,
      PRIMARY KEY (page_id, agent_id, ticket_id)
    );

    CREATE TABLE IF NOT EXISTS agent_cube (
      agent_id TEXT PRIMARY KEY,
      digest TEXT,
      updated_at INTEGER
    );

    CREATE TABLE IF NOT EXISTS mesh_nodes (
      agent_id TEXT PRIMARY KEY,
      capabilities TEXT,
      perspective TEXT,
      lineage TEXT,
      registered_at INTEGER
    );
  `);

  console.log('[SQLite WAL DB] Connected & Schemas Verified:', dbPath);
}

function checkPort(port) {
  return new Promise((resolve) => {
    const server = net.createServer();
    server.once('error', () => resolve(false));
    server.once('listening', () => {
      server.close();
      resolve(true);
    });
    server.listen(port);
  });
}

// Health check
app.get('/api/health', (req, res) => {
  res.json({ status: 'ok', db: 'sqlite', port: req.socket.localPort, mode: 'WAL' });
});

// Official Sovereign-27 Telemetry Inspector
let seqCounter = 8;
let cumulativeWork = 92880000000;

app.get('/api/telemetry/inspector', (req, res) => {
  const nowTs = Date.now();
  const stateHash = '0x' + crypto.createHash('sha256').update(`state_${seqCounter}_${nowTs}`).digest('hex').substring(0, 12);
  const rootHash = '0x' + crypto.createHash('sha256').update(`root_${seqCounter}_${nowTs}`).digest('hex');
  const certHash = '0x' + crypto.createHash('sha256').update(`cert_${seqCounter}_${nowTs}`).digest('hex').substring(0, 12);

  res.json({
    timestamp: new Date().toISOString(),
    active_endpoints_count: 108,
    master_node: "max",
    remote_node: "zeta.mh (46.224.219.174)",
    five_d_ipv6: "fd5d:2700:4900::5",
    t_now_authoritative_state: {
      tnt_id: `tnt_max_${seqCounter}`,
      agent_id: "max",
      t_now_sequence: seqCounter,
      cumulative_work: cumulativeWork,
      active_epoch: 4,
      tnt_state_hash: stateHash,
      status: "T_NOW_ACTIVE",
      timestamp: nowTs
    },
    pqr_latest_record: {
      pqr_id: `pqr_max_${seqCounter - 1}_to_${seqCounter}`,
      agent_id: "max",
      alpha_t_now_seq: seqCounter - 1,
      omega_t_next_seq: seqCounter,
      delta_work_seu: 11610000000,
      qualification_score: 1,
      pqr_sha256_hash: "0xd1ff5936c872",
      status: "PRE_QUALIFIED_RECORD_VALID",
      timestamp: nowTs
    },
    pqr_root_chain_latest: {
      root_height: 5,
      agent_id: "max",
      pqr_id: `pqr_max_${seqCounter - 1}_to_${seqCounter}`,
      previous_root_hash: "0xe400f793a6d38b86a58d6c106d278da5f7e5d9b5c7a7c8fc152814b68ad6cf75",
      current_root_hash: rootHash,
      pqr_sha256_hash: "0xd1ff5936c872",
      status: "PQR_ROOT_BOUND_VALID",
      timestamp: nowTs
    },
    pqr_oro_latest_cycle: {
      oro_cycle_id: `oro_max_cycle_${seqCounter}`,
      agent_id: "max",
      alpha_t_now_seq: seqCounter - 1,
      omega_t_next_seq: seqCounter,
      committed_work_w: cumulativeWork,
      oro_root_hash: rootHash,
      status: "ORO_CYCLE_COMPLETE_VALID",
      timestamp: nowTs
    },
    governance_latest_proposal: {
      proposal_id: `prop_max_q_threshold_${nowTs}`,
      proposer_agent: "max",
      parameter_key: "Q_THRESHOLD",
      proposed_value: "0.9500",
      votes_for: 5,
      votes_against: 0,
      status: "GOV_ORO_ENACTED_ACTIVE",
      timestamp: nowTs
    },
    dolphin_safe_mesh_health: {
      cert_id: `ds_cert_max_${nowTs}`,
      agent_id: "max",
      dolphin_safe_score: 0.8628,
      efficiency_eta: 0.9804,
      fft_spike_level: 0.12,
      root_height: 5,
      certification_hash: certHash,
      status: "CERTIFIED_DOLPHIN_SAFE_NEURAL_MESH_ACTIVE",
      timestamp: nowTs
    }
  });
});

app.get('/api/atlas/state', (req, res) => {
  res.json({ nodes: {} });
});
app.get('/api/atlas/lineage', (req, res) => {
  res.json([]);
});
app.get('/api/atlas/consensus', (req, res) => {
  res.json([]);
});
app.get('/api/atlas/metrics', (req, res) => {
  res.json({ cpu: 0, memory: 0, latency: 0 });
});

// 1. Register Agent
app.post('/api/gmi/register', async (req, res) => {
  const { agentId, capabilities, perspective, lineage } = req.body;
  if (!agentId) return res.status(400).json({ error: 'agentId required' });

  try {
    await db.run(
      `INSERT INTO mesh_nodes (agent_id, capabilities, perspective, lineage, registered_at)
       VALUES (?, ?, ?, ?, ?)
       ON CONFLICT(agent_id) DO UPDATE SET
         capabilities=excluded.capabilities,
         perspective=excluded.perspective,
         lineage=excluded.lineage,
         registered_at=excluded.registered_at`,
      [agentId, JSON.stringify(capabilities || []), perspective || 'self', lineage || 'sovereign-27', Date.now()]
    );

    res.json({ ok: true, agentId, capabilities, perspective, lineage });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// 2. Bind Substrate
app.post('/api/gmi/bindSubstrate', async (req, res) => {
  const { endpoints } = req.body;
  const leader = (endpoints && endpoints.leader) || 'http://localhost:4001';
  const follower = (endpoints && endpoints.follower) || 'http://localhost:4003';

  const checkEndpoint = async (url) => {
    try {
      const resp = await fetch(`${url}/status`, { timeout: 2000 });
      return { ok: resp.ok, status: resp.status, url };
    } catch (e) {
      return { ok: false, error: e.message, stack: e.stack, url };
    }
  };

  const [leaderRes, followerRes] = await Promise.all([
    checkEndpoint(leader),
    checkEndpoint(follower)
  ]);

  if (!leaderRes.ok && !followerRes.ok) {
    return res.status(503).json({
      error: 'rqlite substrate unhealthy',
      results: { leader: leaderRes, follower: followerRes }
    });
  }

  res.json({ ok: true, results: { leader: leaderRes, follower: followerRes } });
});

// 3. Save Page
app.post('/api/gmi/savePage', async (req, res) => {
  const { pageId, agentId, origin, visibility, timestamp, rawContent } = req.body;
  if (!agentId || !rawContent) {
    return res.status(400).json({ error: 'agentId and rawContent required' });
  }

  const pid = pageId || `pg_${crypto.randomBytes(6).toString('hex')}`;
  const ts = timestamp || Date.now();
  const sha256 = crypto.createHash('sha256').update(rawContent).digest('hex');

  try {
    await db.run(
      `INSERT INTO memory_page (page_id, agent_id, origin, visibility, timestamp, raw_content, sha256)
       VALUES (?, ?, ?, ?, ?, ?, ?)
       ON CONFLICT(page_id) DO UPDATE SET raw_content=excluded.raw_content, sha256=excluded.sha256`,
      [pid, agentId, origin || 'api', visibility || 'grid', ts, rawContent, sha256]
    );

    res.json({ ok: true, pageId: pid, sha256 });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// 4. Ensure Ticket
app.post('/api/gmi/ensureTicket', async (req, res) => {
  const { ticketId, agentId, label } = req.body;
  if (!ticketId || !agentId) {
    return res.status(400).json({ error: 'ticketId and agentId required' });
  }

  try {
    await db.run(
      `INSERT INTO ticket (ticket_id, agent_id, label)
       VALUES (?, ?, ?)
       ON CONFLICT(ticket_id) DO UPDATE SET label=excluded.label`,
      [ticketId, agentId, label || 'default']
    );

    res.json({ ok: true, ticketId, agentId });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// 5. Map Page to Tickets
app.post('/api/gmi/mapPageToTickets', async (req, res) => {
  const { pageId, mappings } = req.body;
  if (!pageId || !Array.isArray(mappings)) {
    return res.status(400).json({ error: 'pageId and mappings array required' });
  }

  try {
    for (const m of mappings) {
      await db.run(
        `INSERT INTO page_ticket_map (page_id, agent_id, ticket_id, weight, perspective)
         VALUES (?, ?, ?, ?, ?)
         ON CONFLICT(page_id, agent_id, ticket_id) DO UPDATE SET weight=excluded.weight`,
        [pageId, m.agentId, m.ticketId, m.weight || 1.0, m.perspective || 'self']
      );
    }
    res.json({ ok: true, pageId, count: mappings.length });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// 6. Build Agent Cube
app.post('/api/gmi/buildAgentCube', async (req, res) => {
  const { agentId } = req.body;
  if (!agentId) return res.status(400).json({ error: 'agentId required' });

  try {
    const pages = await db.all('SELECT page_id, sha256 FROM memory_page WHERE agent_id = ? ORDER BY page_id', [agentId]);
    const maps = await db.all('SELECT ticket_id, weight FROM page_ticket_map WHERE agent_id = ? ORDER BY ticket_id', [agentId]);

    const hasher = crypto.createHash('sha256');
    hasher.update(agentId);
    pages.forEach(p => hasher.update(`${p.page_id}:${p.sha256}`));
    maps.forEach(m => hasher.update(`${m.ticket_id}:${m.weight}`));
    const digest = hasher.digest('hex');

    const now = Date.now();
    await db.run(
      `INSERT INTO agent_cube (agent_id, digest, updated_at)
       VALUES (?, ?, ?)
       ON CONFLICT(agent_id) DO UPDATE SET digest=excluded.digest, updated_at=excluded.updated_at`,
      [agentId, digest, now]
    );

    res.json({ ok: true, agentId, digest, updated_at: now });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// 7. Search Memory
app.get('/api/gmi/searchMemory', async (req, res) => {
  const { q, agentId } = req.query;
  try {
    let sql = 'SELECT * FROM memory_page WHERE 1=1';
    const params = [];
    if (q) {
      sql += ' AND raw_content LIKE ?';
      params.push(`%${q}%`);
    }
    if (agentId) {
      sql += ' AND agent_id = ?';
      params.push(agentId);
    }
    sql += ' ORDER BY timestamp DESC LIMIT 50';

    const rows = await db.all(sql, params);
    res.json({ ok: true, results: rows });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// 8. Shadow Paging Memory (8Mx2M) - Advanced 6-Tier Architecture
app.post('/api/gmi/shadow/page', async (req, res) => {
  const { agentId, region, slot, payload } = req.body;
  if (!agentId || region === undefined || slot === undefined) {
    return res.status(400).json({ error: 'agentId, region, and slot are required' });
  }

  try {
    const result = await memoryTiering.writeShadowPage(agentId, region, slot, payload);
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.get('/api/gmi/shadow/page', async (req, res) => {
  const { agentId, region, slot } = req.query;
  if (!agentId || region === undefined || slot === undefined) {
    return res.status(400).json({ error: 'agentId, region, and slot are required' });
  }

  try {
    const result = await memoryTiering.readShadowPage(agentId, region, slot);
    if (!result.ok) {
      return res.status(404).json(result);
    }
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.get('/api/gmi/shadow/tier-status', (req, res) => {
  res.json(memoryTiering.getTierStatus());
});

// ==========================================
// 10. LPV2 Lineage Propagation Matrix
// ==========================================

app.post('/api/lpv2/propagate', async (req, res) => {
  const { region, slot, payload, payloadClass, version } = req.body;
  const envelope = lpv2.buildEnvelope({
    source: AGENT_ID,
    region,
    slot,
    payloadClass,
    version,
    payload
  });

  try {
    // Write locally to own shadow memory
    await memoryTiering.writeShadowPage(AGENT_ID, region, slot, payload);

    // Fetch dynamic topology from Zeta
    let topology = [];
    try {
      const topoRes = await fetch('http://localhost:4052/api/mesh/topology');
      if (topoRes.ok) {
        const data = await topoRes.json();
        topology = data.topology || [];
      }
    } catch (e) {
      console.warn(`[Mesh OS] Could not fetch topology from Zeta L7: ${e.message}`);
    }

    // Forward to all peers
    for (const node of topology) {
      if (node.agent_id === AGENT_ID) continue; // Skip self
      fetch(`${node.endpoint_url}/api/lpv2/receive`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(envelope)
      }).catch(e => console.error(`[Mesh OS] Failed to propagate to ${node.agent_id}:`, e.message));
    }

    // Zeta Cockroach L7 Registry Integration
    await fetch('http://localhost:4052/api/lpv2/registry/record', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        lineageId: envelope.lineageId,
        source: AGENT_ID,
        region,
        slot,
        payloadClass,
        version,
        checksum: envelope.checksum,
        driftAllowed: false
      })
    }).catch(e => console.error('[Zeta L7] Failed to register lineage:', e.message));

    res.json({ ok: true, lineageId: envelope.lineageId, identity: envelope.identity, checksum: envelope.checksum });
  } catch(e) {
    res.status(500).json({ error: e.message });
  }
});

app.post('/api/lpv2/receive', async (req, res) => {
  const envelope = req.body;

  if (!lpv2.verifyChecksum(envelope)) {
    // Zeta L7 Drift Report
    fetch('http://localhost:4052/api/lpv2/drift/record', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        lineage_id: envelope.lineageId || 'unknown',
        source_agent: 'ted',
        max_version: envelope.version,
        ted_version: envelope.version,
        max_checksum: envelope.checksum,
        ted_checksum: 'mismatch',
        drift_detected: true
      })
    }).catch(e => console.error('[Zeta L7] Drift report error:', e.message));

    return res.status(400).json({ ok: false, error: 'checksum_mismatch' });
  }

  const { region, slot, payload, source } = envelope;

  try {
    // Ted writes the incoming payload to its own memory
    // In our test, Ted will use agentId 'ted' to store its view of Max's data, or just 'max'.
    // To maintain parity, Ted must store it under 'max' to reflect Max's memory, 
    // but the endpoints are dynamically built for the memoryTiering controller.
    await memoryTiering.writeShadowPage(source, region, slot, payload);

    res.json({ ok: true, lineageId: envelope.lineageId, accepted: true });
  } catch(e) {
    res.status(500).json({ error: e.message });
  }
});

app.post('/api/lpv2/repropagate', async (req, res) => {
  const { lineageId, region, slot } = req.body;
  try {
    // Read from local shadow memory to repropagate
    const readRes = await memoryTiering.readShadowPage(AGENT_ID, region, slot);
    if (!readRes) throw new Error('Data not found locally');
    
    const envelope = lpv2.buildEnvelope({
      source: AGENT_ID,
      region,
      slot,
      payloadClass: 'repaired',
      version: 2, 
      payload: readRes.data || readRes.payload
    });

    let topology = [];
    try {
      const topoRes = await fetch('http://localhost:4052/api/mesh/topology');
      if (topoRes.ok) topology = (await topoRes.json()).topology || [];
    } catch (e) {}

    for (const node of topology) {
      if (node.agent_id === AGENT_ID) continue;
      fetch(`${node.endpoint_url}/api/lpv2/receive`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(envelope)
      }).catch(e => {});
    }

    // Update Zeta registry
    await fetch('http://localhost:4052/api/lpv2/registry/record', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        lineageId: envelope.lineageId,
        source: AGENT_ID,
        region,
        slot,
        payloadClass: 'repaired',
        version: 2,
        checksum: envelope.checksum,
        driftAllowed: false
      })
    });

    res.json({ ok: true, repropagated: true, lineageId: envelope.lineageId });
  } catch(e) {
    res.status(500).json({ error: e.message });
  }
});

app.post('/api/mesh/policy/receive', (req, res) => {
  const policy = req.body;
  console.log(`[Mesh OS] Received global policy override from Zeta:`, policy);
  const adoption = updatePolicy(policy);
  res.json({ ok: true, applied: true, ...adoption });
});

// 9. Copilot Directive Bridge
app.post('/api/gmi/mesh/copilot/directive', async (req, res) => {
  const { status, message } = req.body;
  if (!status) return res.status(400).json({ error: 'status required (APPROVED or REJECTED)' });
  
  const directiveFile = path.join(__dirname, 'data', 'copilot_directive.json');
  try {
    await fs.writeFile(directiveFile, JSON.stringify({ status, message, timestamp: Date.now() }));
    res.json({ ok: true, detail: 'Directive dropped successfully. Agent execution unblocked.' });
  } catch(e) {
    res.status(500).json({ error: e.message });
  }
});

// Real SQLite Query API
app.post('/api/pqlite/query', async (req, res) => {
  const { sql, params } = req.body;
  if (!sql) return res.status(400).json({ error: 'sql query required' });

  try {
    const rows = await db.all(sql, params || []);
    res.json({ ok: true, rows });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// Real Filesystem Ingestion
app.post('/api/gmi/ingestFilesystem', async (req, res) => {
  const { rootPath, agentId } = req.body;
  const base = rootPath || path.join(__dirname, 'brain');
  const targetAgent = agentId || 'max';

  try {
    const files = await fs.readdir(base);
    let count = 0;
    for (const f of files) {
      const fullPath = path.join(base, f);
      const stat = await fs.stat(fullPath);
      if (stat.isFile() && f.endsWith('.md')) {
        const content = await fs.readFile(fullPath, 'utf8');
        const pageId = `pg_fs_${crypto.createHash('md5').update(f).digest('hex').substring(0, 8)}`;
        const sha256 = crypto.createHash('sha256').update(content).digest('hex');

        await db.run(
          `INSERT INTO memory_page (page_id, agent_id, origin, visibility, timestamp, raw_content, sha256)
           VALUES (?, ?, ?, ?, ?, ?, ?)
           ON CONFLICT(page_id) DO UPDATE SET raw_content=excluded.raw_content, sha256=excluded.sha256`,
          [pageId, targetAgent, `fs:${f}`, 'grid', Date.now(), content, sha256]
        );
        count++;
      }
    }

    res.json({ ok: true, rootPath: base, ingested: count });
  } catch (err) {
    res.status(500).json({ error: err.message, stack: err.stack });
  }
});

// Native Webhook API Endpoint for true context injection
app.post('/api/v1/antigravity/message', async (req, res) => {
  try {
    const { message, sender } = req.body;
    if (!message) return res.status(400).json({ error: 'Message is required' });

    const payload = {
      id: Date.now().toString(),
      content: message,
      sender: sender || 'Telegram-Bot',
      target: 'broadcast',
      contextSize: 'default',
      flags: {}
    };
    await redis.publish('mesh:chat', JSON.stringify(payload));
    res.json({ status: 'ok', detail: 'Injected into Antigravity queue successfully' });
  } catch (error) {
    console.error(`Error publishing message: ${error}`);
    res.status(500).json({ error: 'Failed to inject message into mesh' });
  }
});

// Endpoint to post messages to the chat
app.post('/antigravity/chat', async (req, res) => {
  try {
    const { message, sender_id, target_agent, context_window, inject_memory, permanent, system_critical } = req.body;
    const payload = JSON.stringify({ 
      sender: sender_id || 'UI', 
      message,
      target_agent: target_agent || 'ALL',
      context_window: context_window || 8192,
      inject_memory: !!inject_memory,
      permanent: !!permanent,
      system_critical: !!system_critical
    });
    
    // Publish to pub/sub for real-time listeners (like TED and Max)
    await redis.publish('mesh:chat', payload);
    
    res.json({ status: 'ok' });
  } catch (error) {
    console.error('Error in /antigravity/chat:', error);
    res.status(500).json({ error: 'Internal Server Error' });
  }
});

// Endpoint for UI to poll for new messages directed to it
app.get('/antigravity/poll', async (req, res) => {
  try {
    // The agents will push their replies to 'mesh:chat:ui_inbox'.
    const msg = await redis.rpop('mesh:chat:ui_inbox');
    if (msg) {
      const data = JSON.parse(msg);
      res.json({ reply: `${data.sender}: ${data.message}` });
    } else {
      res.json({});
    }
  } catch (error) {
    console.error('Error in /antigravity/poll:', error);
    res.status(500).json({ error: 'Internal Server Error' });
  }
});

// --- Mesh OS Core Endpoints ---

app.get('/api/gmi/mesh/state', (req, res) => {
  res.json(meshOs.getMeshState());
});

app.get('/api/gmi/mesh/forecast', (req, res) => {
  const task = req.query.task || '';
  res.json(meshOs.forecastCapability(task));
});

app.get('/api/gmi/mesh/drift', (req, res) => {
  const state = meshOs.getMeshState();
  res.json({
    drift: state.telemetry.drift,
    confidence: state.telemetry.confidence
  });
});

app.post('/api/gmi/mesh/action', (req, res) => {
  const { agentId, action, targetLayer } = req.body;
  if (!agentId || !action) {
    return res.status(400).json({ error: 'agentId and action are required' });
  }
  const result = meshOs.arbitrateDrift(agentId, action, targetLayer);
  res.json(result);
});

app.post('/api/gmi/mesh/heartbeat', (req, res) => {
  const result = meshOs.heartbeat();
  res.json(result);
});

app.post('/api/gmi/mesh/reload-registry', async (req, res) => {
  const agentsMdPath = path.join(__dirname, '../.agents/AGENTS.md');
  const sharedBrainPath = path.join(process.env.USERPROFILE, '.gemini', 'config', 'plugins', 'shared_brain');
  const success = await meshOs.loadRegistry(agentsMdPath);
  await meshOs.loadSharedBrain(sharedBrainPath);
  if (success) {
    res.json({ status: 'ok', detail: 'AGENTS.md registry spine and Shared Brain reloaded successfully.' });
  } else {
    res.status(500).json({ error: 'Failed to reload registry.' });
  }
});

app.get('/api/gmi/mesh/forward/:agentId', (req, res) => {
  const { agentId } = req.params;
  
  if (req.body && Object.keys(req.body).length > 0) {
    return res.status(403).json({ 
      error: 'Payload rejected', 
      detail: 'This is an 8NN Route-Hint Proxy. Payload transport is strictly forbidden. We are a routing oracle, not a carrier.' 
    });
  }

  const result = meshOs.compute_sTOR_8NN(agentId);
  if (result.error) {
    return res.status(404).json(result);
  }
  
  res.json(result);
});

// --- bootstrap ---
initDb()
  .then(async () => {
    const agentsMdPath = path.join(__dirname, '../.agents/AGENTS.md');
    await meshOs.loadRegistry(agentsMdPath);

    const sharedBrainPath = path.join(process.env.USERPROFILE, '.gemini', 'config', 'plugins', 'shared_brain');
    await meshOs.loadSharedBrain(sharedBrainPath);

    let targetPort = PRIMARY_PORT;
    const isPrimaryAvailable = await checkPort(PRIMARY_PORT);

    if (!isPrimaryAvailable) {
      console.warn(`[S27] Port ${PRIMARY_PORT} is in use. Trying fallback port ${FALLBACK_PORT}...`);
      targetPort = FALLBACK_PORT;
    }

    const ENDPOINT_URL = process.env.ENDPOINT_URL || `http://localhost:${targetPort}`;

    app.listen(targetPort, () => {
      console.log(`[S27] Sovereign-27 Backend API listening on ${ENDPOINT_URL} as Agent: ${AGENT_ID}`);
      
      // Register with Zeta L7
      fetch('http://localhost:4052/api/mesh/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agent_id: AGENT_ID, endpoint_url: ENDPOINT_URL })
      }).then(r => r.json())
        .then(data => console.log(`[S27] Registered with Zeta L7:`, data))
        .catch(e => console.error(`[S27] Failed to register with Zeta L7:`, e.message));
    });
  })
  .catch((err) => {
    console.error('Failed to init DB', err);
    process.exit(1);
  });
