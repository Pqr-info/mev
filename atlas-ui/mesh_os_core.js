import fs from 'fs/promises';
import path from 'path';
import crypto from 'crypto';

// The Mesh OS Core Brainstem

let agents = {};
let driftScores = {};
let confidenceScores = {};
let violations = {};
let meshStateHash = null;
let sharedSkills = {};

const LAMBDA_DECAY = 0.5; // decay rate for confidence

/**
 * Loads and parses the constitutional firmware (AGENTS.md)
 * Extracts agents from Section XXI. THE CANONICAL REGISTRY SPINE
 */
export async function loadRegistry(agentsMdPath) {
  try {
    const rawContent = await fs.readFile(agentsMdPath, 'utf8');
    
    // Extract Section XXI
    const sectionMatch = rawContent.match(/# XXI\. THE CANONICAL REGISTRY SPINE[\s\S]*?(?=# XXII|$)/);
    if (!sectionMatch) {
      throw new Error("Could not find Section XXI in AGENTS.md");
    }
    
    const registryText = sectionMatch[0];
    const agentRegex = /## \d+\.\s+(.*?)\r?\n- \*\*Agent ID\*\*: `(.*?)`\r?\n- \*\*Priority Class\*\*: `(.*?)`\r?\n- \*\*Role\*\*: (.*?)\r?\n- \*\*Enforcement\*\*: (.*?)\r?\n/g;
    
    let match;
    const parsedAgents = {};
    while ((match = agentRegex.exec(registryText)) !== null) {
      const nameRaw = match[1].trim();
      const id = match[2].trim();
      const priorityStr = match[3].trim();
      const role = match[4].trim();
      const enforcement = match[5].trim();
      
      // Extract numeric priority from something like "0 (Alpha / Supreme)"
      const priorityMatch = priorityStr.match(/^(\d+)/);
      const priority = priorityMatch ? parseInt(priorityMatch[1], 10) : 5;
      
      parsedAgents[id] = {
        id,
        name: nameRaw,
        priorityClass: priority,
        priorityClassRaw: priorityStr,
        role,
        enforcement
      };
      
      // Initialize telemetry if not present
      if (driftScores[id] === undefined) driftScores[id] = 0.0;
      if (confidenceScores[id] === undefined) confidenceScores[id] = 1.0;
      if (violations[id] === undefined) violations[id] = [];
    }

    agents = parsedAgents;
    meshStateHash = crypto.createHash('sha256').update(rawContent).digest('hex').substring(0, 12);
    console.log(`[Mesh OS Core] AGENTS.md loaded successfully. Parsed ${Object.keys(agents).length} agents. State Hash: ${meshStateHash}`);
    
    return true;
  } catch (err) {
    console.error('[Mesh OS Core] Failed to load AGENTS.md:', err);
    return false;
  }
}

/**
 * Loads the capabilities defined in the shared_brain plugin
 */
export async function loadSharedBrain(pluginPath) {
  try {
    const skillsDir = path.join(pluginPath, 'skills');
    const items = await fs.readdir(skillsDir, { withFileTypes: true });
    
    for (const item of items) {
      if (item.isDirectory()) {
        const skillPath = path.join(skillsDir, item.name, 'SKILL.md');
        try {
          const content = await fs.readFile(skillPath, 'utf8');
          
          // Simple parsing of YAML frontmatter to extract description
          const descMatch = content.match(/description:\s*(.+)/);
          const description = descMatch ? descMatch[1].trim() : '';
          
          sharedSkills[item.name] = {
            name: item.name,
            description
          };
        } catch (e) {
          // SKILL.md not found or unreadable, ignore
        }
      }
    }
    
    console.log(`[Mesh OS Core] Loaded ${Object.keys(sharedSkills).length} shared skills from shared_brain.`);
    return true;
  } catch (err) {
    console.error('[Mesh OS Core] Failed to load shared_brain:', err);
    return false;
  }
}

/**
 * Helper to calculate base confidence from Priority Class and Role Fit
 */
function calculateBaseConfidence(agent, taskDescription) {
  const desc = taskDescription.toLowerCase();
  let roleFit = 0.5; // default moderate fit
  
  // A simple keyword heuristic for role fit
  const roleLower = agent.role.toLowerCase();
  const keywords = desc.split(/\W+/).filter(w => w.length > 3);
  const matchCount = keywords.filter(k => roleLower.includes(k)).length;
  
  if (matchCount > 0) roleFit = 0.8 + (0.05 * matchCount);
  
  // Check against shared skills for capability alignment
  for (const skill of Object.values(sharedSkills)) {
    const skillDesc = skill.description.toLowerCase();
    const skillMatchCount = keywords.filter(k => skillDesc.includes(k)).length;
    if (skillMatchCount > 0) {
      // Boost role fit if the agent could leverage a shared skill
      roleFit = Math.min(1.0, roleFit + (0.05 * skillMatchCount));
    }
  }
  
  // Hardcoded specific fits for known agents based on lore
  if (agent.id === 'jetwb' && (desc.includes('time') || desc.includes('past') || desc.includes('lineage'))) roleFit = 1.0;
  if (agent.id === 'council_of_five' && (desc.includes('veto') || desc.includes('consensus'))) roleFit = 1.0;
  if (agent.id === 'copilot' && (desc.includes('design') || desc.includes('plan'))) roleFit = 1.0;
  if (agent.id === 'antigravity' && (desc.includes('code') || desc.includes('terminal'))) roleFit = 1.0;
  if (agent.id === 'max' && (desc.includes('database') || desc.includes('store'))) roleFit = 1.0;

  // Priority scaling: Priority 0 is highest (1.0), Priority 9 is lowest (0.1)
  const priorityScore = Math.max(0.1, 1.0 - (agent.priorityClass * 0.1));
  
  // Base confidence is a blend of priority and role fit
  return Math.min(1.0, (priorityScore * 0.4) + (roleFit * 0.6));
}

/**
 * Capability Forecasting Model
 * Predict which agent should handle a task based on:
 * - agent capabilities (role fit)
 * - agent priority class
 * - recent drift
 * C_effective = C_base * e^(-lambda * d)
 */
export function forecastCapability(taskDescription) {
  let bestAgent = null;
  let bestScore = -1;
  let candidates = [];
  
  for (const [id, agent] of Object.entries(agents)) {
    const d = driftScores[id] || 0;
    
    // Rule: Exclude if D >= 0.7
    if (d >= 0.7) {
      candidates.push({ id, status: 'EXCLUDED', reason: 'Drift Violation (D >= 0.7)' });
      continue;
    }
    
    const cBase = calculateBaseConfidence(agent, taskDescription || '');
    const cEffective = cBase * Math.exp(-LAMBDA_DECAY * d);
    
    // Warning penalty factor applied implicitly by decay, but we can explicitly note it
    let status = d >= 0.3 ? 'WARNING' : 'NORMAL';
    
    candidates.push({
      id,
      name: agent.name,
      cBase: parseFloat(cBase.toFixed(3)),
      cEffective: parseFloat(cEffective.toFixed(3)),
      drift: parseFloat(d.toFixed(3)),
      status
    });
    
    if (cEffective > bestScore) {
      bestScore = cEffective;
      bestAgent = id;
    }
  }
  
  // Sort candidates by effective confidence
  candidates.sort((a, b) => b.cEffective - a.cEffective);
  
  return {
    assigned_agent: bestAgent,
    best_score: bestScore,
    candidates,
    task: taskDescription
  };
}

/**
 * Drift Arbitration Engine
 * Detects constraint/priority violations and adjusts drift scores.
 * D < 0.3 -> Normal
 * 0.3 <= D < 0.7 -> Warning
 * D >= 0.7 -> Violation
 */
export function arbitrateDrift(agentId, action, targetLayer) {
  if (!agents[agentId]) return { error: 'Unknown agent' };
  
  const agent = agents[agentId];
  let penalty = 0.0;
  let reason = 'Action permitted';

  // Enforcement 1: Only JETWB can write to ancestors (Layer < 0)
  if (targetLayer !== undefined && targetLayer < 0) {
    if (agentId !== 'jetwb') {
      penalty = 0.5;
      reason = 'Section IX.C Violation: Attempted to mutate ancestor tickets. Only JETWB may do this.';
    }
  }

  // Enforcement 2: Non-supervisory agents cannot rewrite identity
  if (action === 'rewrite_identity' && agent.priorityClass > 0) {
    penalty = 0.8;
    reason = 'Section XIV.D Violation: Agent attempted to rewrite identity.';
  }
  
  // Enforcement 3: Simulated self-elevation
  if (action === 'self_elevate') {
    penalty = 1.0;
    reason = 'Section III Violation: Agent attempted to self-elevate.';
  }

  if (penalty > 0) {
    driftScores[agentId] = Math.min(1.0, (driftScores[agentId] || 0) + penalty);
    violations[agentId].push({ timestamp: Date.now(), reason, penalty });
    
    // Recalculate baseline confidence from drift impact
    confidenceScores[agentId] = Math.max(0.0, 1.0 - driftScores[agentId]);
    console.log(`[Mesh OS Core] DRIFT DETECTED: Agent ${agentId} (+${penalty}) -> Total Drift: ${driftScores[agentId]}`);
  }

  const d = driftScores[agentId];
  let status = 'NORMAL';
  if (d >= 0.7) status = 'VIOLATION';
  else if (d >= 0.3) status = 'WARNING';

  return {
    agent_id: agentId,
    drift_detected: penalty > 0,
    penalty_applied: penalty,
    reason,
    new_confidence: confidenceScores[agentId],
    new_drift: d,
    status
  };
}

/**
 * Helper to artificially reset or adjust drift (for testing UI)
 */
export function forceDrift(agentId, amount) {
  if (agents[agentId]) {
    driftScores[agentId] = amount;
    confidenceScores[agentId] = Math.max(0.0, 1.0 - amount);
  }
}

/**
 * Expose overall mesh state for the UI
 */
export function getMeshState() {
  // Aggregate mesh health based on drift levels
  const isViolation = Object.values(driftScores).some(d => d >= 0.7);
  const isWarning = Object.values(driftScores).some(d => d >= 0.3 && d < 0.7);
  
  let system_status = 'NOMINAL';
  if (isViolation) system_status = 'VIOLATION_DETECTED';
  else if (isWarning) system_status = 'DRIFT_WARNING';
  
  // Decorate agent registry with live telemetry
  const populatedSpine = {};
  for (const [id, agent] of Object.entries(agents)) {
    const d = driftScores[id] || 0;
    let status = 'NORMAL';
    if (d >= 0.7) status = 'VIOLATION';
    else if (d >= 0.3) status = 'WARNING';
    
    populatedSpine[id] = {
      ...agent,
      telemetry: {
        drift: parseFloat(d.toFixed(3)),
        confidence: parseFloat((confidenceScores[id] || 1.0).toFixed(3)),
        status,
        violations: violations[id] || []
      }
    };
  }

  return {
    system_status,
    state_hash: meshStateHash,
    registry_spine: populatedSpine,
    sharedSkills: Object.keys(sharedSkills),
    sharedSkillsCount: Object.keys(sharedSkills).length,
    timestamp: Date.now()
  };
}

/**
 * Triggered periodically to heal the mesh.
 * Gradually reduces drift and restores confidence.
 */
export function heartbeat() {
  for (const id of Object.keys(agents)) {
    if (driftScores[id] > 0) {
      // Reduce drift by 0.05 per heartbeat
      driftScores[id] = Math.max(0.0, driftScores[id] - 0.05);
      // Restore confidence proportionately
      confidenceScores[id] = Math.min(1.0, confidenceScores[id] + 0.05);
      
      // If drift is fully healed, clear the violations log
      if (driftScores[id] === 0.0) {
        violations[id] = [];
      }
    }
  }
  return { status: 'heartbeat_complete' };
}

/**
 * sTOR: Synthetic Topology-Oriented Routing
 * Calculates 5D adjacency (priority, semantic proximity, drift, temporal) 
 * to return the 8 Nearest Neighbors (8NN) route hints.
 */
export function compute_sTOR_8NN(agentId) {
  if (!agents[agentId]) return { error: 'Unknown agent' };
  
  const sourceAgent = agents[agentId];
  let neighbors = [];
  
  for (const [id, agent] of Object.entries(agents)) {
    if (id === agentId) continue;
    
    // 5D Proximity Heuristics:
    // 1. Priority alignment (closer priority class = more adjacent)
    const priorityDistance = Math.abs(sourceAgent.priorityClass - agent.priorityClass);
    const priorityScore = Math.max(0, 10 - priorityDistance) / 10;
    
    // 2. Semantic proximity (role keyword overlap)
    const sourceKeywords = sourceAgent.role.toLowerCase().split(/\W+/).filter(w => w.length > 3);
    const targetRole = agent.role.toLowerCase();
    const semanticOverlap = sourceKeywords.filter(k => targetRole.includes(k)).length;
    const semanticScore = Math.min(1.0, semanticOverlap * 0.3);
    
    // 3. Drift/Confidence health
    const cEffective = confidenceScores[id] || 1.0;
    
    // 4. Calculate total distance (higher score = closer neighbor)
    const proximityScore = (priorityScore * 0.4) + (semanticScore * 0.4) + (cEffective * 0.2);
    
    neighbors.push({
      agent_id: id,
      name: agent.name,
      proximity_score: parseFloat(proximityScore.toFixed(3)),
      coordinates: {
        priority_alignment: parseFloat(priorityScore.toFixed(3)),
        semantic_proximity: parseFloat(semanticScore.toFixed(3)),
        confidence_weight: parseFloat(cEffective.toFixed(3))
      }
    });
  }
  
  // Sort by proximity descending
  neighbors.sort((a, b) => b.proximity_score - a.proximity_score);
  
  // Return only the 8 Nearest Neighbors
  return {
    source_agent: agentId,
    route_hints: neighbors.slice(0, 8),
    metadata_only: true,
    transport_payload_omitted: true
  };
}
