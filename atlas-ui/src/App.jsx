import React, { useState } from 'react';
import { Activity, Hexagon, Terminal, LayoutDashboard, MessageSquare, Shield, Lock, AlertTriangle, Zap } from 'lucide-react';
import './index.css';

import AtlasView from './components/AtlasView';
import Marketplace from './components/Marketplace';
import AntigravityChat from './components/AntigravityChat';
import GovernanceView from './components/GovernanceView';
import MEVArbitragePanel from './components/MEVArbitragePanel';
import SecureVaultModal from './components/SecureVaultModal';
import { GovernanceCapabilities } from './engine/GovernanceCapabilities';
import { bindGovernanceToConstitution } from './engine/GovernanceBinding';

// Register the Auditor capabilities with the Governance Constitution
bindGovernanceToConstitution({}, GovernanceCapabilities);

export default function App() {
  const [activeTab, setActiveTab] = useState('atlas');
  const [chatOpen, setChatOpen] = useState(true);
  const [vaultModalOpen, setVaultModalOpen] = useState(false);
  const [isSimulationMode, setIsSimulationMode] = useState(false); // Default: Native Real-World Execution

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh', width: '100vw', overflow: 'hidden', background: 'var(--bg-0)' }}>
      
      {/* MANDATORY RULE: Flashing Red Banner if Simulation Mode Enabled with Human in the Loop */}
      {isSimulationMode && (
        <div style={{
          background: 'linear-gradient(90deg, #dc2626, #991b1b, #dc2626)',
          color: '#ffffff',
          padding: '0.5rem 1rem',
          textAlign: 'center',
          fontWeight: 900,
          fontSize: '0.9rem',
          letterSpacing: '2px',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          gap: '12px',
          boxShadow: '0 0 25px rgba(220, 38, 38, 0.8)',
          zIndex: 99999,
          animation: 'pulse 1.5s infinite'
        }}>
          <AlertTriangle size={20} color="#fff" />
          <span>⚠️ THIS IS A SIMULATION — ACTIVE HUMAN IN THE LOOP AUTHORIZED ⚠️</span>
          <button
            onClick={() => setIsSimulationMode(false)}
            style={{
              background: '#000',
              border: '1px solid #fff',
              color: '#fff',
              padding: '0.2rem 0.6rem',
              borderRadius: '6px',
              cursor: 'pointer',
              fontSize: '0.75rem',
              fontWeight: 800,
              marginLeft: '1rem'
            }}
          >
            RETURN TO NATIVE REAL-WORLD MODE
          </button>
        </div>
      )}

      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Navigation Sidebar */}
        <nav style={{ width: '80px', borderRight: '1px solid var(--border)', display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '1rem 0', gap: '1rem', background: 'var(--bg-1)', zIndex: 50 }}>
          <div style={{ padding: '0.5rem', background: 'rgba(59,130,246,0.1)', borderRadius: '8px', marginBottom: '0.5rem' }}>
            <Activity size={24} color="var(--color-blue)" />
          </div>
          
          <button 
            onClick={() => setActiveTab('atlas')}
            style={{ background: 'transparent', border: 'none', color: activeTab === 'atlas' ? 'var(--color-blue)' : 'var(--text-secondary)', cursor: 'pointer', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px' }}
          >
            <LayoutDashboard size={24} />
            <span style={{ fontSize: '10px' }}>Atlas</span>
          </button>

          <button 
            onClick={() => setActiveTab('governance')}
            style={{ background: 'transparent', border: 'none', color: activeTab === 'governance' ? 'var(--color-blue)' : 'var(--text-secondary)', cursor: 'pointer', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px' }}
          >
            <Shield size={24} />
            <span style={{ fontSize: '10px' }}>Governance</span>
          </button>

          <button 
            onClick={() => setActiveTab('mev')}
            style={{ background: 'transparent', border: 'none', color: activeTab === 'mev' ? '#f59e0b' : 'var(--text-secondary)', cursor: 'pointer', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px' }}
          >
            <Zap size={24} color={activeTab === 'mev' ? '#f59e0b' : 'inherit'} />
            <span style={{ fontSize: '10px', color: activeTab === 'mev' ? '#f59e0b' : 'inherit' }}>MEV</span>
          </button>

          <button 
            onClick={() => setVaultModalOpen(true)}
            style={{ background: 'transparent', border: 'none', color: '#00f3ff', cursor: 'pointer', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px' }}
          >
            <Lock size={24} color="#00f3ff" />
            <span style={{ fontSize: '10px', color: '#00f3ff' }}>Vault</span>
          </button>
          
          <button 
            onClick={() => setActiveTab('marketplace')}
            style={{ background: 'transparent', border: 'none', color: activeTab === 'marketplace' ? 'var(--color-blue)' : 'var(--text-secondary)', cursor: 'pointer', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px' }}
          >
            <Hexagon size={24} />
            <span style={{ fontSize: '10px' }}>Market</span>
          </button>

          <div style={{ marginTop: 'auto', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.5rem' }}>
            <button
              onClick={() => setIsSimulationMode(!isSimulationMode)}
              title="Toggle Human-in-the-Loop Simulation Guard"
              style={{
                background: isSimulationMode ? '#dc2626' : '#1e293b',
                border: `1px solid ${isSimulationMode ? '#ef4444' : '#334155'}`,
                borderRadius: '8px',
                padding: '0.4rem',
                color: isSimulationMode ? '#fff' : '#94a3b8',
                cursor: 'pointer',
                fontSize: '9px',
                fontWeight: 700,
                textAlign: 'center'
              }}
            >
              {isSimulationMode ? 'SIM: ON' : 'NATIVE'}
            </button>

            <button 
              onClick={() => setChatOpen(!chatOpen)}
              style={{ background: 'transparent', border: 'none', color: chatOpen ? 'var(--color-blue)' : 'var(--text-secondary)', cursor: 'pointer', padding: '0.5rem' }}
            >
              <MessageSquare size={20} />
            </button>
          </div>
        </nav>

        {/* Main Content Area (Scrollable Viewport) */}
        <main style={{ flex: 1, position: 'relative', overflowY: 'auto', overflowX: 'hidden', height: '100%' }}>
          {activeTab === 'atlas' && <AtlasView />}
          {activeTab === 'governance' && <GovernanceView />}
          {activeTab === 'mev' && <MEVArbitragePanel />}
          {activeTab === 'marketplace' && <Marketplace />}
          
          <SecureVaultModal 
            isOpen={vaultModalOpen} 
            onClose={() => setVaultModalOpen(false)} 
          />
        </main>

        {/* Side Drawer Chat */}
        {chatOpen && (
          <aside style={{ width: '400px', borderLeft: '1px solid var(--border)', background: 'var(--bg-1)', zIndex: 40 }}>
            <AntigravityChat onClose={() => setChatOpen(false)} />
          </aside>
        )}
      </div>
    </div>
  );
}
