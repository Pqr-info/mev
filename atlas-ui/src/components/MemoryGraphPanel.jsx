import React, { useState, useEffect } from 'react';
import { Database, Zap, Cpu, RefreshCw, Sparkles, Layers, RotateCcw } from 'lucide-react';
import { FuzzyMemoryGraphEngine } from '../engine/FuzzyMemoryGraphEngine';
import { ContextLoadController } from '../engine/ContextLoadController';
import { ContextStateTracker } from '../engine/ContextStateTracker';

export default function MemoryGraphPanel({ onClose }) {
  const [tickets, setTickets] = useState([]);
  const [selectedSlot, setSelectedSlot] = useState(19); // Default to Ticket 19 (TSRE)
  const [loading, setLoading] = useState(true);

  // Priming Mode state: 'COLD_LOAD' | 'HOTLOAD'
  const [loadMode, setLoadMode] = useState('COLD_LOAD');
  const [activeVolley, setActiveVolley] = useState(1);

  const fetchTickets = async () => {
    try {
      setLoading(true);
      const res = await fetch('http://localhost:4052/api/memorygraph/tickets');
      const data = await res.json();
      if (data.ok) {
        // Calculate fuzzy scores using FuzzyMemoryGraphEngine
        const scored = data.tickets.map(t => ({
          ...t,
          fuzzy_score: FuzzyMemoryGraphEngine.computeFuzzyScore(t)
        }));
        setTickets(scored);
      }
    } catch (e) {
      console.error('Failed to fetch MemoryGraph tickets:', e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTickets();
  }, []);

  const handleInteract = async (slotIndex) => {
    try {
      const res = await fetch('http://localhost:4052/api/memorygraph/interact', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ slot_index: slotIndex })
      });
      const data = await res.json();
      if (data.ok) {
        await fetchTickets();
      }
    } catch (e) {
      console.error('Failed to interact with MemoryGraph ticket:', e);
    }
  };

  const handleResetSession = () => {
    ContextStateTracker.resetSession();
    setLoadMode('COLD_LOAD');
    setActiveVolley(1);
  };

  const selectedTicket = tickets.find(t => t.slot_index === selectedSlot);
  const adaptivePayload = selectedTicket ? FuzzyMemoryGraphEngine.formatAdaptivePayload(selectedTicket) : null;
  const recallMeta = selectedTicket ? FuzzyMemoryGraphEngine.getRecallLevel(selectedTicket.fuzzy_score) : null;

  // Generate Priming Payload based on active mode & volley
  const nowCube = { timestamp: Date.now(), status: 'LIVE_DELTAS_ACTIVE', active_node: 'node-sovereign-27' };
  const primingPayload = loadMode === 'COLD_LOAD' 
    ? ContextLoadController.generateColdLoadVolley(tickets, activeVolley, nowCube)
    : ContextLoadController.generateHotloadPayload(tickets, nowCube);

  const getSlotColor = (score) => {
    if (!score) return 'rgba(255,255,255,0.05)';
    if (score >= 0.80) return 'rgba(16, 185, 129, 0.25)'; // Green
    if (score >= 0.40) return 'rgba(59, 130, 246, 0.25)'; // Blue
    return 'rgba(107, 114, 128, 0.15)'; // Gray
  };

  const getSlotBorder = (score, isSelected) => {
    if (isSelected) return '2px solid #a855f7';
    if (score >= 0.80) return '1px solid #10b981';
    if (score >= 0.40) return '1px solid #3b82f6';
    return '1px solid var(--border)';
  };

  return (
    <div className="memorygraph-panel" style={{ display: 'flex', flexDirection: 'column', height: '100%', color: '#f5f5f7' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', borderBottom: '1px solid var(--border)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#a855f7' }}>
          <Database size={20} />
          <span style={{ fontWeight: 700, fontSize: '1rem' }}>49-Position Relational MemoryGraph</span>
          <span className="tag" style={{ background: 'rgba(168, 85, 247, 0.2)', color: '#a855f7', border: '1px solid #a855f7', fontSize: '0.7rem', padding: '0.1rem 0.4rem', borderRadius: '4px' }}>CHUNKED PRIMING</span>
        </div>
        <button onClick={onClose} style={{ background: 'transparent', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', fontSize: '1.2rem' }}>×</button>
      </div>

      {/* Priming Mode Controls Bar */}
      <div style={{ padding: '0.75rem 1rem', background: 'rgba(0,0,0,0.3)', borderBottom: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <span style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Loading Strategy:</span>
          <button
            onClick={() => setLoadMode('COLD_LOAD')}
            style={{
              padding: '0.3rem 0.6rem',
              borderRadius: '4px',
              fontSize: '0.75rem',
              fontWeight: 600,
              cursor: 'pointer',
              background: loadMode === 'COLD_LOAD' ? 'rgba(59, 130, 246, 0.25)' : 'transparent',
              color: loadMode === 'COLD_LOAD' ? '#3b82f6' : 'var(--text-secondary)',
              border: `1px solid ${loadMode === 'COLD_LOAD' ? '#3b82f6' : 'var(--border)'}`
            }}>
            Cold Load (7 Volleys)
          </button>
          <button
            onClick={() => setLoadMode('HOTLOAD')}
            style={{
              padding: '0.3rem 0.6rem',
              borderRadius: '4px',
              fontSize: '0.75rem',
              fontWeight: 600,
              cursor: 'pointer',
              background: loadMode === 'HOTLOAD' ? 'rgba(16, 185, 129, 0.25)' : 'transparent',
              color: loadMode === 'HOTLOAD' ? '#10b981' : 'var(--text-secondary)',
              border: `1px solid ${loadMode === 'HOTLOAD' ? '#10b981' : 'var(--border)'}`
            }}>
            Hotload (Delta Volley)
          </button>
        </div>

        {/* 7-Volley Stepper Buttons when in COLD_LOAD mode */}
        {loadMode === 'COLD_LOAD' && (
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginRight: '0.25rem' }}>Volley:</span>
            {[1, 2, 3, 4, 5, 6, 7].map(v => (
              <button
                key={v}
                onClick={() => setActiveVolley(v)}
                style={{
                  width: '24px',
                  height: '24px',
                  borderRadius: '4px',
                  fontSize: '0.7rem',
                  fontWeight: 700,
                  cursor: 'pointer',
                  background: activeVolley === v ? '#a855f7' : 'rgba(255,255,255,0.05)',
                  color: activeVolley === v ? '#fff' : 'var(--text-secondary)',
                  border: activeVolley === v ? '1px solid #a855f7' : '1px solid var(--border)'
                }}>
                {v}
              </button>
            ))}
          </div>
        )}

        <button onClick={handleResetSession} style={{ background: 'transparent', border: '1px solid var(--border)', color: 'var(--text-secondary)', padding: '0.2rem 0.5rem', borderRadius: '4px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.3rem', fontSize: '0.75rem' }}>
          <RotateCcw size={12} /> Reset Context
        </button>
      </div>

      {/* Main Container: Left 7x7 Grid, Right Chunk / Volley Inspector */}
      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Left: 49 Grid Slots (7x7) */}
        <div style={{ width: '45%', padding: '1rem', borderRight: '1px solid var(--border)', overflowY: 'auto' }}>
          <div style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.75rem', display: 'flex', justifyContent: 'space-between' }}>
            <span>49 Ticket Matrix (7 x 7 Slots)</span>
            <span style={{ fontSize: '0.75rem' }}>Selected: Slot #{selectedSlot}</span>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: '0.4rem' }}>
            {tickets.map(t => {
              const isSel = t.slot_index === selectedSlot;
              // Highlight slots in the active Volley chunk
              const isVolleySlot = loadMode === 'COLD_LOAD' &&
                t.slot_index >= (activeVolley - 1) * 7 + 1 &&
                t.slot_index <= activeVolley * 7;

              return (
                <div
                  key={t.slot_index}
                  onClick={() => setSelectedSlot(t.slot_index)}
                  style={{
                    aspectRatio: '1',
                    background: isVolleySlot ? 'rgba(168, 85, 247, 0.3)' : getSlotColor(t.fuzzy_score),
                    border: getSlotBorder(t.fuzzy_score, isSel),
                    borderRadius: '6px',
                    padding: '0.3rem',
                    cursor: 'pointer',
                    display: 'flex',
                    flexDirection: 'column',
                    justify: 'space-between',
                    transition: 'all 0.15s ease-in-out',
                    transform: isSel ? 'scale(1.05)' : 'none'
                  }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.65rem', fontWeight: 700, color: isSel ? '#a855f7' : 'var(--text-secondary)' }}>
                    <span>#{t.slot_index}</span>
                    {t.auditor_flags > 0 && <span style={{ color: '#ef4444' }}>!</span>}
                  </div>
                  <div style={{ fontSize: '0.75rem', fontWeight: 700, textAlign: 'center', color: t.fuzzy_score >= 0.8 ? '#10b981' : t.fuzzy_score >= 0.4 ? '#3b82f6' : '#9ca3af' }}>
                    {t.fuzzy_score}
                  </div>
                </div>
              );
            })}
          </div>

          {/* Color Legend */}
          <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1rem', fontSize: '0.7rem', color: 'var(--text-secondary)', justifyContent: 'center' }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: '0.2rem' }}><span style={{ width: 8, height: 8, borderRadius: 2, background: 'rgba(16, 185, 129, 0.4)', border: '1px solid #10b981' }}></span> Level 3 ($\ge 0.8$)</span>
            <span style={{ display: 'flex', alignItems: 'center', gap: '0.2rem' }}><span style={{ width: 8, height: 8, borderRadius: 2, background: 'rgba(59, 130, 246, 0.4)', border: '1px solid #3b82f6' }}></span> Level 2 ($\ge 0.4$)</span>
            <span style={{ display: 'flex', alignItems: 'center', gap: '0.2rem' }}><span style={{ width: 8, height: 8, borderRadius: 2, background: 'rgba(107, 114, 128, 0.2)', border: '1px solid #6b7280' }}></span> Level 1</span>
          </div>
        </div>

        {/* Right: Volley & Priming Payload Inspector */}
        <div style={{ flex: 1, padding: '1rem', overflowY: 'auto', background: 'rgba(0,0,0,0.1)' }}>
          <div style={{ marginBottom: '1rem' }}>
            <div style={{ fontSize: '0.75rem', color: loadMode === 'COLD_LOAD' ? '#3b82f6' : '#10b981', fontWeight: 700, letterSpacing: '0.5px' }}>
              {loadMode === 'COLD_LOAD' ? `PRIMING VOLLEY ${activeVolley} OF 7 (Slots #${(activeVolley - 1) * 7 + 1} - #${activeVolley * 7})` : 'HOTLOAD DELTA VOLLEY 1 OF 1'}
            </div>
            <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
              Hash: <code style={{ color: '#a855f7' }}>{primingPayload.state_hash}</code>
            </div>
          </div>

          {/* Formatted Header */}
          <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>Header & De-Duplication Directives:</div>
          <pre style={{ background: '#09090b', padding: '0.6rem', borderRadius: '4px', border: '1px solid var(--border)', fontSize: '0.7rem', color: '#60a5fa', marginBottom: '1rem', whiteSpace: 'pre-wrap' }}>
            {primingPayload.header}
          </pre>

          {/* Ticket Inspector inside active volley */}
          {selectedTicket && (
            <div style={{ marginTop: '0.5rem', borderTop: '1px solid var(--border)', paddingTop: '0.75rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                <span style={{ fontSize: '0.8rem', fontWeight: 700, color: '#f5f5f7' }}>Slot #{selectedTicket.slot_index}: {selectedTicket.title}</span>
                <button
                  onClick={() => handleInteract(selectedTicket.slot_index)}
                  style={{ background: 'rgba(168, 85, 247, 0.2)', color: '#a855f7', border: '1px solid #a855f7', padding: '0.25rem 0.5rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.3rem' }}>
                  <Sparkles size={12} /> Touch (Boost Score)
                </button>
              </div>

              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>Formatted Ticket Payload (Adaptive Recall):</div>
              <pre style={{ background: '#0a0a0c', padding: '0.6rem', borderRadius: '4px', border: '1px solid var(--border)', fontSize: '0.7rem', color: '#34d399', overflowX: 'auto', whiteSpace: 'pre-wrap' }}>
                {JSON.stringify(adaptivePayload, null, 2)}
              </pre>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
