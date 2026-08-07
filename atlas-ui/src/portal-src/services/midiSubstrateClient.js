/**
 * Substrate 27 + CockroachDB In-Memory MIDI State Machine Client
 * High-performance, ultra-low latency in-memory MIDI engine with SQLite WAL & CockroachDB long-term persistence.
 */

export class MidiStateMachine {
  constructor() {
    this.sessionId = `midi_sess_${Date.now()}`;
    this.bpm = 120.0;
    this.currentTick = 0;
    this.state = 'IDLE'; // IDLE, RECORDING, PLAYBACK, PATTERN_MUTATED, QUANTIZED
    this.channels = Array.from({ length: 16 }, () => ({ notes: {}, cc: {}, program: 0, volume: 100 }));
    this.eventBuffer = [];
  }

  setBpm(bpm) {
    this.bpm = parseFloat(bpm);
    this.eventBuffer.push({ type: 'BPM_CHANGE', value: bpm, tick: this.currentTick });
  }

  noteOn(channel, note, velocity) {
    if (channel < 0 || channel > 15 || note < 0 || note > 127) return;
    this.channels[channel].notes[note] = velocity;
    const evt = {
      type: 'NOTE_ON',
      channel,
      note,
      velocity,
      tick: this.currentTick,
      timestampMs: Date.now()
    };
    this.eventBuffer.push(evt);
    this.state = 'RECORDING';
    return evt;
  }

  noteOff(channel, note) {
    if (this.channels[channel]?.notes[note] !== undefined) {
      delete this.channels[channel].notes[note];
    }
    const evt = {
      type: 'NOTE_OFF',
      channel,
      note,
      velocity: 0,
      tick: this.currentTick,
      timestampMs: Date.now()
    };
    this.eventBuffer.push(evt);
    return evt;
  }

  controlChange(channel, ccNumber, value) {
    this.channels[channel].cc[ccNumber] = value;
    const evt = {
      type: 'CONTROL_CHANGE',
      channel,
      ccNumber,
      value,
      tick: this.currentTick,
      timestampMs: Date.now()
    };
    this.eventBuffer.push(evt);
    return evt;
  }

  advanceTicks(ticks = 24) {
    this.currentTick += ticks;
  }

  exportSnapshot() {
    return {
      sessionId: this.sessionId,
      bpm: this.bpm,
      currentTick: this.currentTick,
      state: this.state,
      channels: this.channels,
      eventCount: this.eventBuffer.length
    };
  }
}

export class Substrate27MidiClient {
  constructor(apiBase = 'http://localhost:4000', cockroachHost = '46.224.219.174') {
    this.apiBase = apiBase;
    this.cockroachHost = cockroachHost;
    this.stateMachine = new MidiStateMachine();
  }

  async persistMidiSnapshot(agentId = 'max') {
    const snapshot = this.stateMachine.exportSnapshot();
    const rawContent = JSON.stringify(snapshot, null, 2);

    try {
      const resp = await fetch(`${this.apiBase}/api/gmi/savePage`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agentId,
          origin: 'midi:state_machine',
          visibility: 'grid',
          rawContent
        })
      });
      const data = await resp.json();

      return {
        ok: true,
        pageId: data.pageId,
        sha256: data.sha256,
        snapshot
      };
    } catch (e) {
      return {
        ok: false,
        error: e.message,
        snapshot
      };
    }
  }
}

export const midiSubstrateClient = new Substrate27MidiClient();
