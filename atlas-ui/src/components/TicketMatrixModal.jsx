import React, { useState, useEffect } from 'react';
import { X, Search, FileText, CheckCircle2, Clock, ExternalLink, ShieldCheck } from 'lucide-react';

/**
 * TicketMatrixModal.jsx — Sovereign-27 In-App Ticket Matrix Viewer
 * 
 * Remediation for Browser File-Protocol Link Blocker (Ticket S27-TKT-0023):
 * 1. Fetches live tickets from /api/tickets (Zeta Master Compute Port 4052).
 * 2. Displays interactive, searchable 22+ Ticket Sequence Matrix directly inside Atlas UI.
 * 3. Provides clickable HTTP REST link (/api/tickets) for external browser view.
 */

export default function TicketMatrixModal({ isOpen, onClose }) {
  const [tickets, setTickets] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [loading, setLoading] = useState(true);
  const [selectedTicket, setSelectedTicket] = useState(null);

  useEffect(() => {
    if (!isOpen) return;

    const fetchTickets = async () => {
      setLoading(true);
      try {
        const res = await fetch('/api/tickets');
        if (res.ok) {
          const data = await res.json();
          if (data.ok && Array.isArray(data.tickets)) {
            setTickets(data.tickets);
            if (data.tickets.length > 0) {
              setSelectedTicket(data.tickets[data.tickets.length - 1]); // Select latest by default
            }
          }
        }
      } catch (e) {
        console.error('[Ticket Matrix Modal] Failed to fetch tickets:', e.message);
      } finally {
        setLoading(false);
      }
    };

    fetchTickets();
  }, [isOpen]);

  if (!isOpen) return null;

  const filteredTickets = tickets.filter(t => 
    t.ticket_code.toLowerCase().includes(searchTerm.toLowerCase()) ||
    t.title.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      background: 'rgba(5, 7, 15, 0.85)',
      backdropFilter: 'blur(12px)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 99990,
      padding: '1.5rem'
    }}>
      <div style={{
        background: 'linear-gradient(135deg, #0f172a, #1e293b)',
        border: '1px solid rgba(56, 189, 248, 0.4)',
        borderRadius: '16px',
        width: '100%',
        maxWidth: '1100px',
        maxHeight: '85vh',
        display: 'flex',
        flexDirection: 'column',
        boxShadow: '0 25px 50px -12px rgba(0, 243, 255, 0.25)',
        overflow: 'hidden'
      }}>
        {/* Header Bar */}
        <div style={{
          padding: '1.25rem 1.5rem',
          borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          background: 'rgba(15, 23, 42, 0.8)'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <ShieldCheck size={24} color="#38bdf8" />
            <div>
              <h2 style={{ fontSize: '1.15rem', fontWeight: 800, color: '#f8fafc', margin: 0 }}>
                Sovereign-27 Self-Healing Multi-Ticket Matrix
              </h2>
              <p style={{ fontSize: '0.75rem', color: '#94a3b8', margin: 0 }}>
                Persistent Sequence ID Engine ({tickets.length} Active Tickets • Next: S27-TKT-00{tickets.length + 1})
              </p>
            </div>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <a
              href="/api/tickets"
              target="_blank"
              rel="noopener noreferrer"
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '6px',
                padding: '0.4rem 0.8rem',
                borderRadius: '8px',
                background: 'rgba(56, 189, 248, 0.15)',
                border: '1px solid #38bdf8',
                color: '#38bdf8',
                fontSize: '0.75rem',
                fontWeight: 700,
                textDecoration: 'none'
              }}
            >
              <span>Open Raw REST API</span>
              <ExternalLink size={12} />
            </a>

            <button
              onClick={onClose}
              style={{
                background: 'transparent',
                border: 'none',
                color: '#94a3b8',
                cursor: 'pointer',
                padding: '0.2rem',
                display: 'flex',
                alignItems: 'center'
              }}
            >
              <X size={24} />
            </button>
          </div>
        </div>

        {/* Modal Body */}
        <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
          {/* Left Column: Ticket List */}
          <div style={{
            width: '360px',
            borderRight: '1px solid rgba(255, 255, 255, 0.1)',
            display: 'flex',
            flexDirection: 'column',
            background: 'rgba(15, 23, 42, 0.5)'
          }}>
            <div style={{ padding: '0.75rem', borderBottom: '1px solid rgba(255, 255, 255, 0.05)' }}>
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                background: '#0f172a',
                border: '1px solid #334155',
                borderRadius: '8px',
                padding: '0.4rem 0.75rem'
              }}>
                <Search size={16} color="#64748b" />
                <input
                  type="text"
                  placeholder="Filter sequence tickets..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  style={{
                    background: 'transparent',
                    border: 'none',
                    color: '#fff',
                    outline: 'none',
                    width: '100%',
                    fontSize: '0.8rem'
                  }}
                />
              </div>
            </div>

            <div style={{ flex: 1, overflowY: 'auto', padding: '0.5rem' }}>
              {loading ? (
                <div style={{ color: '#94a3b8', padding: '1rem', textAlign: 'center', fontSize: '0.85rem' }}>
                  Loading Ticket Matrix...
                </div>
              ) : filteredTickets.map(t => {
                const isSelected = selectedTicket?.ticket_code === t.ticket_code;
                return (
                  <div
                    key={t.ticket_code}
                    onClick={() => setSelectedTicket(t)}
                    style={{
                      padding: '0.65rem 0.85rem',
                      borderRadius: '8px',
                      marginBottom: '0.4rem',
                      cursor: 'pointer',
                      background: isSelected ? 'rgba(56, 189, 248, 0.15)' : 'transparent',
                      border: `1px solid ${isSelected ? '#38bdf8' : 'transparent'}`,
                      transition: 'all 0.15s ease'
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.2rem' }}>
                      <span style={{ fontSize: '0.75rem', fontWeight: 800, color: '#38bdf8' }}>{t.ticket_code}</span>
                      <span style={{ fontSize: '0.65rem', padding: '0.1rem 0.4rem', borderRadius: '4px', background: 'rgba(16, 185, 129, 0.2)', color: '#10b981', fontWeight: 700 }}>
                        {t.status}
                      </span>
                    </div>
                    <div style={{ fontSize: '0.8rem', color: '#e2e8f0', fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {t.title}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Right Column: 5-Step Details Panel */}
          <div style={{ flex: 1, padding: '1.5rem', overflowY: 'auto', background: '#0b1120' }}>
            {selectedTicket ? (
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', borderBottom: '1px solid rgba(255, 255, 255, 0.1)', paddingBottom: '0.75rem' }}>
                  <div>
                    <span style={{ fontSize: '0.85rem', fontWeight: 800, color: '#38bdf8', letterSpacing: '1px' }}>
                      TICKET {selectedTicket.ticket_code} (Seq #{selectedTicket.seq_num})
                    </span>
                    <h3 style={{ fontSize: '1.25rem', fontWeight: 700, color: '#fff', marginTop: '0.2rem' }}>
                      {selectedTicket.title}
                    </h3>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '6px', padding: '0.3rem 0.6rem', background: 'rgba(16, 185, 129, 0.2)', border: '1px solid #10b981', borderRadius: '6px', color: '#10b981', fontSize: '0.75rem', fontWeight: 800 }}>
                    <CheckCircle2 size={14} />
                    <span>{selectedTicket.status}</span>
                  </div>
                </div>

                {/* 5-Step Matrix breakdown */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  <div style={{ background: 'rgba(15, 23, 42, 0.8)', border: '1px solid rgba(255, 255, 255, 0.08)', borderRadius: '10px', padding: '1rem' }}>
                    <h4 style={{ fontSize: '0.8rem', fontWeight: 800, color: '#f59e0b', textTransform: 'uppercase', marginBottom: '0.4rem' }}>
                      1. WHAT (Defect / Feature Signature)
                    </h4>
                    <p style={{ fontSize: '0.85rem', color: '#cbd5e1', lineHeight: '1.5', margin: 0 }}>
                      {selectedTicket.step_what}
                    </p>
                  </div>

                  <div style={{ background: 'rgba(15, 23, 42, 0.8)', border: '1px solid rgba(255, 255, 255, 0.08)', borderRadius: '10px', padding: '1rem' }}>
                    <h4 style={{ fontSize: '0.8rem', fontWeight: 800, color: '#38bdf8', textTransform: 'uppercase', marginBottom: '0.4rem' }}>
                      2. HOW (Architectural Remediation Executed)
                    </h4>
                    <p style={{ fontSize: '0.85rem', color: '#cbd5e1', lineHeight: '1.5', margin: 0 }}>
                      {selectedTicket.step_how}
                    </p>
                  </div>

                  <div style={{ background: 'rgba(15, 23, 42, 0.8)', border: '1px solid rgba(255, 255, 255, 0.08)', borderRadius: '10px', padding: '1rem' }}>
                    <h4 style={{ fontSize: '0.8rem', fontWeight: 800, color: '#ef4444', textTransform: 'uppercase', marginBottom: '0.4rem' }}>
                      3. WHY UNCAUGHT (Root-Cause Vulnerability Analysis)
                    </h4>
                    <p style={{ fontSize: '0.85rem', color: '#cbd5e1', lineHeight: '1.5', margin: 0 }}>
                      {selectedTicket.step_why_uncaught}
                    </p>
                  </div>

                  <div style={{ background: 'rgba(15, 23, 42, 0.8)', border: '1px solid rgba(255, 255, 255, 0.08)', borderRadius: '10px', padding: '1rem' }}>
                    <h4 style={{ fontSize: '0.8rem', fontWeight: 800, color: '#10b981', textTransform: 'uppercase', marginBottom: '0.4rem' }}>
                      4. HOW DO WE CATCH THIS AUTOMATICALLY & HEAL IT NEXT TIME
                    </h4>
                    <p style={{ fontSize: '0.85rem', color: '#cbd5e1', lineHeight: '1.5', margin: 0 }}>
                      {selectedTicket.step_auto_catch_heal}
                    </p>
                  </div>

                  <div style={{ background: 'rgba(15, 23, 42, 0.8)', border: '1px solid rgba(255, 255, 255, 0.08)', borderRadius: '10px', padding: '1rem' }}>
                    <h4 style={{ fontSize: '0.8rem', fontWeight: 800, color: '#a855f7', textTransform: 'uppercase', marginBottom: '0.4rem' }}>
                      5. DOCUMENTATION UPDATE
                    </h4>
                    <p style={{ fontSize: '0.85rem', color: '#cbd5e1', lineHeight: '1.5', margin: 0 }}>
                      {selectedTicket.step_doc_update}
                    </p>
                  </div>
                </div>
              </div>
            ) : (
              <div style={{ color: '#64748b', textAlign: 'center', padding: '3rem' }}>
                Select a ticket from the left sequence to inspect 5-step details.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
