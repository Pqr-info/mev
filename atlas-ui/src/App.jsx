import React, { useState } from 'react';
import { Activity, Hexagon, Terminal, LayoutDashboard, MessageSquare } from 'lucide-react';
import './index.css';

import AtlasView from './components/AtlasView';
import Marketplace from './components/Marketplace';
import AntigravityChat from './components/AntigravityChat';

export default function App() {
  const [activeTab, setActiveTab] = useState('atlas');
  const [chatOpen, setChatOpen] = useState(true);

  return (
    <div style={{ display: 'flex', height: '100vh', width: '100vw', overflow: 'hidden', background: 'var(--bg-0)' }}>
      {/* Navigation Sidebar */}
      <nav style={{ width: '80px', borderRight: '1px solid var(--border)', display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '1rem 0', gap: '1rem', background: 'var(--bg-1)', zIndex: 50 }}>
        <div style={{ padding: '0.5rem', background: 'rgba(59,130,246,0.1)', borderRadius: '8px', marginBottom: '1rem' }}>
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
          onClick={() => setActiveTab('marketplace')}
          style={{ background: 'transparent', border: 'none', color: activeTab === 'marketplace' ? 'var(--color-blue)' : 'var(--text-secondary)', cursor: 'pointer', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px' }}
        >
          <Hexagon size={24} />
          <span style={{ fontSize: '10px' }}>Market</span>
        </button>

        <button 
          onClick={() => setActiveTab('portal')}
          style={{ background: 'transparent', border: 'none', color: activeTab === 'portal' ? 'var(--color-blue)' : 'var(--text-secondary)', cursor: 'pointer', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px' }}
        >
          <Terminal size={24} />
          <span style={{ fontSize: '10px' }}>Portal</span>
        </button>

        <div style={{ flex: 1 }} />
        
        <button 
          onClick={() => setChatOpen(!chatOpen)}
          style={{ background: 'transparent', border: 'none', color: chatOpen ? 'var(--color-green)' : 'var(--text-secondary)', cursor: 'pointer', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px' }}
        >
          <MessageSquare size={24} />
          <span style={{ fontSize: '10px' }}>Chat</span>
        </button>
      </nav>

      {/* Main Content Area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {activeTab === 'atlas' && <AtlasView />}
        {activeTab === 'marketplace' && <Marketplace />}
        {activeTab === 'portal' && (
          <iframe 
            src="/portal.html" 
            style={{ width: '100%', height: '100%', border: 'none' }}
            title="Portal View"
          />
        )}
      </div>

      {/* Global Chat UI */}
      {chatOpen && (
        <div style={{ width: '350px', borderLeft: '1px solid var(--border)', background: 'var(--bg-1)', display: 'flex', flexDirection: 'column', zIndex: 40 }}>
          <AntigravityChat />
        </div>
      )}
    </div>
  );
}
