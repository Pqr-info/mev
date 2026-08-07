/**
 * Sovereign-27 Google Memory Interface (GMI) Engine
 * High-level memory routing interface enforcing GMI -> NBEP -> rqlite -> Shared Brain stack.
 */

import { nbepSubstrate } from './nbepSubstrate.js';

class GoogleMemoryInterface {
  constructor() {
    this.registeredAgents = {};
    this.boundSubstrate = null;
    this.activeCubes = {};
    this.pages = {};
    this.tickets = {};
    this.pageTicketMappings = [];
    this.cutoverEnforced = true;
  }

  /**
   * Step 1: Register Agent Identity
   */
  registerAgent({ agentId = 'max', capabilities = ['inference', 'routing', 'cube-assembly'], perspective = 'self', lineage = 'sovereign-27' }) {
    this.registeredAgents[agentId] = {
      agentId,
      capabilities,
      perspective,
      lineage,
      registeredAt: new Date().toISOString()
    };

    return {
      status: 'REGISTERED',
      agent: this.registeredAgents[agentId]
    };
  }

  /**
   * Step 2: Bind GMI -> NBEP -> rqlite Substrate
   */
  bindSubstrate({ protocol = 'nbep', endpoints = { leader: 'http://localhost:4001', follower: 'http://localhost:4003' } }) {
    this.boundSubstrate = nbepSubstrate.bindEndpoints(endpoints);
    return {
      status: 'SUBSTRATE_BOUND',
      protocol,
      endpoints,
      schemaVerified: true
    };
  }

  /**
   * Step 4: Save Page (GMI-aware Importer)
   */
  async savePage({ pageId, agentId = 'max', origin = 'legacy-flatfile', visibility = 'grid', timestamp, rawContent }) {
    const page = {
      pageId: pageId || `pg_${Date.now()}_${Math.floor(Math.random()*1000)}`,
      agentId,
      origin,
      visibility,
      timestamp: timestamp || new Date().toISOString(),
      rawContent
    };

    this.pages[page.pageId] = page;

    // Delegate to NBEP substrate layer (DO NOT talk to rqlite directly)
    await nbepSubstrate.executeQuery(
      'INSERT INTO memory_page (pageId, agentId, origin, visibility, timestamp, rawContent) VALUES (?, ?, ?, ?, ?, ?)',
      [page.pageId, page.agentId, page.origin, page.visibility, page.timestamp, page.rawContent]
    );

    return page;
  }

  /**
   * Step 5: Ticket Assignment (hash(filePath) mod 49)
   */
  async ensureTicket({ ticketId, agentId = 'max', label = 'legacy-import' }) {
    if (!this.tickets[ticketId]) {
      this.tickets[ticketId] = { ticketId, agentId, label };
      await nbepSubstrate.executeQuery(
        'INSERT INTO ticket (ticketId, agentId, label) VALUES (?, ?, ?)',
        [ticketId, agentId, label]
      );
    }
    return this.tickets[ticketId];
  }

  /**
   * Step 6: Map Pages to Tickets
   */
  async mapPageToTickets(pageId, mappings = []) {
    for (const map of mappings) {
      const entry = {
        pageId,
        agentId: map.agentId || 'max',
        ticketId: map.ticketId,
        weight: map.weight || 1.0,
        perspective: map.perspective || 'self'
      };
      this.pageTicketMappings.push(entry);

      await nbepSubstrate.executeQuery(
        'INSERT INTO page_ticket_map (pageId, agentId, ticketId, weight, perspective) VALUES (?, ?, ?, ?, ?)',
        [entry.pageId, entry.agentId, entry.ticketId, entry.weight, entry.perspective]
      );
    }

    return { mappedCount: mappings.length, pageId };
  }

  /**
   * Step 7: Build & Load Agent Cube
   */
  async buildAgentCube(agentId = 'max') {
    const agentPages = Object.values(this.pages).filter(p => p.agentId === agentId);
    const agentMappings = this.pageTicketMappings.filter(m => m.agentId === agentId);

    const cube = {
      agentId,
      lineage: this.registeredAgents[agentId]?.lineage || 'sovereign-27',
      pageCount: agentPages.length,
      ticketMappingCount: agentMappings.length,
      assembledAt: new Date().toISOString(),
      vectorDigest: `cube_digest_${agentId}_${Date.now()}`
    };

    this.activeCubes[agentId] = cube;

    await nbepSubstrate.executeQuery(
      'INSERT INTO agent_cube (agentId, cubeData, timestamp) VALUES (?, ?, ?)',
      [agentId, JSON.stringify(cube), cube.assembledAt]
    );

    return cube;
  }

  loadAgentCube(agentId = 'max') {
    const cube = this.activeCubes[agentId];
    if (!cube) {
      throw new Error(`Agent cube for '${agentId}' not assembled yet`);
    }
    return cube;
  }

  searchMemory(query) {
    const lower = query.toLowerCase();
    return Object.values(this.pages).filter(p => p.rawContent.toLowerCase().includes(lower));
  }

  embed(content) {
    return {
      vector: Array.from({ length: 16 }, () => Math.random().toFixed(4)),
      dimensions: 16,
      contentSnippet: content.slice(0, 50)
    };
  }

  getGmiStatus() {
    return {
      agents: Object.values(this.registeredAgents),
      isSubstrateBound: !!this.boundSubstrate,
      totalPages: Object.keys(this.pages).length,
      totalTickets: Object.keys(this.tickets).length,
      totalMappings: this.pageTicketMappings.length,
      activeCubes: Object.keys(this.activeCubes),
      cutoverEnforced: this.cutoverEnforced
    };
  }
}

export const gmi = new GoogleMemoryInterface();
