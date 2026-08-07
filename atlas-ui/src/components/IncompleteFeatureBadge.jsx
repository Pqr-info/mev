import React, { useState } from 'react';
import { Clock, ExternalLink, ArrowRight, CheckCircle2, ListFilter } from 'lucide-react';
import TicketMatrixModal from './TicketMatrixModal';

/**
 * IncompleteFeatureBadge.jsx — Sovereign-27 Completion Tracking Badge
 * 
 * Mandate:
 * If a feature or module is under active hyperdevelopment (< 100% complete):
 * 1. Renders explicit ~X% completion progress bar & badge.
 * 2. Provides direct clickable link to Queued Tickets Matrix (In-App Modal & /api/tickets REST).
 * 3. Displays status of the NEXT ticket required for 100% completion.
 */

export default function IncompleteFeatureBadge({
  featureName = 'System Module',
  percentComplete = 85,
  nextTicketId = '007',
  nextTicketTitle = 'High-Throughput DMA Cache Optimization',
  nextTicketStatus = 'IN_PROGRESS'
}) {
  const [isMatrixModalOpen, setIsMatrixModalOpen] = useState(false);
  const isComplete = percentComplete >= 100;

  if (isComplete) {
    return (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', padding: '0.25rem 0.6rem', background: 'rgba(16, 185, 129, 0.15)', border: '1px solid #10b981', borderRadius: '6px', color: '#10b981', fontSize: '0.75rem', fontWeight: 700 }}>
        <CheckCircle2 size={14} />
        <span>100% COMPLETE</span>
      </div>
    );
  }

  return (
    <>
      <div style={{
        background: 'linear-gradient(135deg, rgba(15, 23, 42, 0.95), rgba(30, 41, 59, 0.95))',
        border: '1px solid rgba(245, 158, 11, 0.4)',
        borderRadius: '12px',
        padding: '0.75rem 1rem',
        marginBottom: '1rem',
        boxShadow: '0 0 20px rgba(245, 158, 11, 0.15)',
        fontFamily: 'sans-serif'
      }}>
        {/* Top Header Row */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Clock size={16} color="#f59e0b" />
            <span style={{ fontSize: '0.85rem', fontWeight: 800, color: '#f59e0b', letterSpacing: '0.5px' }}>
              {featureName}: ~{percentComplete}% COMPLETE
            </span>
          </div>
          
          <button
            onClick={() => setIsMatrixModalOpen(true)}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
              color: '#38bdf8',
              background: 'rgba(56, 189, 248, 0.15)',
              border: '1px solid #38bdf8',
              borderRadius: '6px',
              padding: '0.3rem 0.6rem',
              fontSize: '0.75rem',
              fontWeight: 700,
              cursor: 'pointer'
            }}
          >
            <ListFilter size={14} />
            <span>Queued Tickets Matrix</span>
          </button>
        </div>

        {/* Visual Progress Bar */}
        <div style={{ width: '100%', height: '6px', background: '#0f172a', borderRadius: '3px', overflow: 'hidden', marginBottom: '0.6rem', border: '1px solid rgba(245, 158, 11, 0.2)' }}>
          <div style={{
            width: `${percentComplete}%`,
            height: '100%',
            background: 'linear-gradient(90deg, #f59e0b, #eab308, #38bdf8)',
            borderRadius: '3px',
            transition: 'width 0.5s ease-in-out'
          }} />
        </div>

        {/* Next Ticket Status Mapping Row */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.75rem', color: '#94a3b8' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ color: '#f8fafc', fontWeight: 600 }}>Next Ticket #{nextTicketId}:</span>
            <span style={{ color: '#cbd5e1' }}>{nextTicketTitle}</span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <span style={{
              padding: '0.15rem 0.4rem',
              borderRadius: '4px',
              fontSize: '0.65rem',
              fontWeight: 800,
              background: nextTicketStatus === 'IN_PROGRESS' ? 'rgba(56, 189, 248, 0.2)' : 'rgba(245, 158, 11, 0.2)',
              color: nextTicketStatus === 'IN_PROGRESS' ? '#38bdf8' : '#f59e0b',
              border: `1px solid ${nextTicketStatus === 'IN_PROGRESS' ? '#38bdf8' : '#f59e0b'}`
            }}>
              {nextTicketStatus}
            </span>
            <ArrowRight size={12} color="#94a3b8" />
          </div>
        </div>
      </div>

      {/* Interactive Ticket Matrix Modal */}
      <TicketMatrixModal
        isOpen={isMatrixModalOpen}
        onClose={() => setIsMatrixModalOpen(false)}
      />
    </>
  );
}
