import React, { useState, useEffect } from 'react';
import { FileText, Search, ArrowUpRight, CheckCircle2, Shield, Activity, Layers, ExternalLink, Filter } from 'lucide-react';

/**
 * ColorCodedTransactionLedger.jsx — Sovereign-27 Multi-Environment Color-Coded Transaction Ledger
 * 
 * Requirements:
 * 1. Full color-coded transaction table for ALL transactions across LIVE (Green), TEST (Blue), and DEV (Amber).
 * 2. Interactive filtering by Environment (ALL, 🟢 LIVE, 🔵 TEST, 🟡 DEV) & live search.
 * 3. Displays TX Hash, Type, From -> To, Amount, Asset, Gas Fee, Timestamp, and Status.
 */

export default function ColorCodedTransactionLedger() {
  const [transactions, setTransactions] = useState([]);
  const [activeFilter, setActiveFilter] = useState('ALL');
  const [searchTerm, setSearchTerm] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchTransactions = async () => {
      try {
        const res = await fetch('/api/ledger/transactions');
        if (res.ok) {
          const data = await res.json();
          if (data.ok && Array.isArray(data.transactions)) {
            setTransactions(data.transactions);
          }
        }
      } catch (e) {
        console.error('[Color-Coded Ledger] Fetch failed:', e.message);
      } finally {
        setLoading(false);
      }
    };

    fetchTransactions();
    const interval = setInterval(fetchTransactions, 5000);
    return () => clearInterval(interval);
  }, []);

  const filteredTxs = transactions.filter(tx => {
    const matchesEnv = activeFilter === 'ALL' || tx.env === activeFilter;
    const matchesSearch = 
      tx.tx_hash.toLowerCase().includes(searchTerm.toLowerCase()) ||
      tx.from_addr.toLowerCase().includes(searchTerm.toLowerCase()) ||
      tx.to_addr.toLowerCase().includes(searchTerm.toLowerCase()) ||
      tx.type.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesEnv && matchesSearch;
  });

  const getEnvStyle = (env) => {
    switch (env) {
      case 'LIVE':
        return {
          bg: 'rgba(16, 185, 129, 0.12)',
          border: '1px solid #10b981',
          color: '#10b981',
          label: '🟢 LIVE MAINNET',
          dot: '#10b981'
        };
      case 'TEST':
        return {
          bg: 'rgba(56, 189, 248, 0.12)',
          border: '1px solid #38bdf8',
          color: '#38bdf8',
          label: '🔵 TEST SEPOLIA',
          dot: '#38bdf8'
        };
      case 'DEV':
        return {
          bg: 'rgba(245, 158, 11, 0.12)',
          border: '1px solid #f59e0b',
          color: '#f59e0b',
          label: '🟡 DEV SANDBOX',
          dot: '#f59e0b'
        };
      default:
        return {
          bg: 'rgba(148, 163, 184, 0.12)',
          border: '1px solid #94a3b8',
          color: '#94a3b8',
          label: env,
          dot: '#94a3b8'
        };
    }
  };

  return (
    <div style={{
      background: 'linear-gradient(135deg, rgba(15, 23, 42, 0.95), rgba(30, 41, 59, 0.95))',
      border: '1px solid rgba(255, 255, 255, 0.1)',
      borderRadius: '16px',
      padding: '1.5rem',
      marginBottom: '1.5rem',
      boxShadow: '0 15px 35px rgba(0, 0, 0, 0.45)',
      fontFamily: 'sans-serif'
    }}>
      {/* Header Bar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.2rem', borderBottom: '1px solid rgba(255, 255, 255, 0.08)', paddingBottom: '0.85rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <FileText size={24} color="#38bdf8" />
          <div>
            <h3 style={{ fontSize: '1.15rem', fontWeight: 800, color: '#f8fafc', margin: 0, letterSpacing: '0.5px' }}>
              Full Color-Coded Swarm Transaction Ledger
            </h3>
            <p style={{ fontSize: '0.75rem', color: '#94a3b8', margin: 0 }}>
              Live Real-World, Testnet, and Sandbox Execution Telemetry ({filteredTxs.length} Transactions)
            </p>
          </div>
        </div>

        {/* Environment Filter Tabs & Search */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          {/* Search Box */}
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            background: '#0f172a',
            border: '1px solid #334155',
            borderRadius: '8px',
            padding: '0.35rem 0.75rem'
          }}>
            <Search size={14} color="#64748b" />
            <input
              type="text"
              placeholder="Search TX hash, address, type..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              style={{
                background: 'transparent',
                border: 'none',
                color: '#fff',
                outline: 'none',
                fontSize: '0.75rem',
                width: '180px'
              }}
            />
          </div>

          {/* Category Filter Pills */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '4px', background: '#0f172a', padding: '0.25rem', borderRadius: '8px', border: '1px solid #334155' }}>
            {['ALL', 'LIVE', 'TEST', 'DEV'].map(env => (
              <button
                key={env}
                onClick={() => setActiveFilter(env)}
                style={{
                  padding: '0.3rem 0.65rem',
                  borderRadius: '6px',
                  border: 'none',
                  fontSize: '0.7rem',
                  fontWeight: 800,
                  cursor: 'pointer',
                  background: activeFilter === env ? (
                    env === 'LIVE' ? '#10b981' : env === 'TEST' ? '#38bdf8' : env === 'DEV' ? '#f59e0b' : '#3b82f6'
                  ) : 'transparent',
                  color: activeFilter === env ? '#0f172a' : '#94a3b8',
                  transition: 'all 0.15s ease'
                }}
              >
                {env === 'ALL' ? 'ALL TXS' : env === 'LIVE' ? '🟢 LIVE' : env === 'TEST' ? '🔵 TEST' : '🟡 DEV'}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Transaction Ledger Table */}
      <div style={{ overflowX: 'auto', borderRadius: '10px', border: '1px solid rgba(255, 255, 255, 0.08)' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.8rem' }}>
          <thead>
            <tr style={{ background: '#0f172a', color: '#94a3b8', borderBottom: '1px solid rgba(255, 255, 255, 0.1)' }}>
              <th style={{ padding: '0.75rem 1rem' }}>ENVIRONMENT</th>
              <th style={{ padding: '0.75rem 1rem' }}>TRANSACTION HASH</th>
              <th style={{ padding: '0.75rem 1rem' }}>TYPE</th>
              <th style={{ padding: '0.75rem 1rem' }}>FROM → TO</th>
              <th style={{ padding: '0.75rem 1rem', textAlign: 'right' }}>AMOUNT</th>
              <th style={{ padding: '0.75rem 1rem', textAlign: 'right' }}>GAS FEE</th>
              <th style={{ padding: '0.75rem 1rem' }}>STATUS</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={7} style={{ padding: '2rem', textAlign: 'center', color: '#94a3b8' }}>
                  Fetching Color-Coded Swarm Ledger...
                </td>
              </tr>
            ) : filteredTxs.length === 0 ? (
              <tr>
                <td colSpan={7} style={{ padding: '2rem', textAlign: 'center', color: '#64748b' }}>
                  No transactions match current search / filter criteria.
                </td>
              </tr>
            ) : filteredTxs.map((tx, idx) => {
              const style = getEnvStyle(tx.env);
              return (
                <tr
                  key={tx.id || idx}
                  style={{
                    borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
                    background: idx % 2 === 0 ? 'rgba(15, 23, 42, 0.3)' : 'transparent',
                    transition: 'background 0.15s ease'
                  }}
                >
                  {/* Environment Tag */}
                  <td style={{ padding: '0.75rem 1rem' }}>
                    <span style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '5px',
                      padding: '0.2rem 0.5rem',
                      borderRadius: '6px',
                      background: style.bg,
                      border: style.border,
                      color: style.color,
                      fontSize: '0.65rem',
                      fontWeight: 800
                    }}>
                      {style.label}
                    </span>
                  </td>

                  {/* Hash */}
                  <td style={{ padding: '0.75rem 1rem', fontFamily: 'monospace', color: '#e2e8f0' }}>
                    <span title={tx.tx_hash}>
                      {tx.tx_hash.slice(0, 10)}...{tx.tx_hash.slice(-8)}
                    </span>
                  </td>

                  {/* Type */}
                  <td style={{ padding: '0.75rem 1rem', fontWeight: 700, color: '#f8fafc' }}>
                    {tx.type}
                  </td>

                  {/* From -> To */}
                  <td style={{ padding: '0.75rem 1rem', fontFamily: 'monospace', color: '#94a3b8', fontSize: '0.75rem' }}>
                    <span>{tx.from_addr.slice(0, 8)}</span>
                    <span style={{ color: '#38bdf8', margin: '0 4px' }}>→</span>
                    <span>{tx.to_addr.slice(0, 8)}</span>
                  </td>

                  {/* Amount */}
                  <td style={{ padding: '0.75rem 1rem', textAlign: 'right', fontWeight: 800, color: style.color }}>
                    {tx.amount} {tx.asset}
                  </td>

                  {/* Gas */}
                  <td style={{ padding: '0.75rem 1rem', textAlign: 'right', color: '#64748b', fontSize: '0.75rem' }}>
                    {tx.gas_fee} ETH
                  </td>

                  {/* Status */}
                  <td style={{ padding: '0.75rem 1rem' }}>
                    <span style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '4px',
                      padding: '0.15rem 0.45rem',
                      borderRadius: '4px',
                      background: 'rgba(16, 185, 129, 0.15)',
                      color: '#10b981',
                      fontSize: '0.65rem',
                      fontWeight: 800
                    }}>
                      <CheckCircle2 size={10} />
                      {tx.status}
                    </span>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
