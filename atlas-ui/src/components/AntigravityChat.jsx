import React, { useState, useRef, useEffect } from 'react';
import { Terminal, Send, Bot, User, Cpu, Database, ShieldAlert, Zap, Globe, Layers } from 'lucide-react';

export default function AntigravityChat() {
  const [messages, setMessages] = useState([
    { role: 'system', content: 'Antigravity Multi-LLM Command Center Initialized.' },
    { role: 'agent', content: 'Ready for mesh commands. You can route prompts to specific agents, control context size, and flag injections.' }
  ]);
  const [input, setInput] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  
  // Advanced Controls State
  const [targetAgent, setTargetAgent] = useState('ALL');
  const [contextWindow, setContextWindow] = useState(8192);
  const [injectMemory, setInjectMemory] = useState(false);
  const [isPermanent, setIsPermanent] = useState(false);
  const [isSystemCritical, setIsSystemCritical] = useState(false);

  const messagesEndRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    const pollInterval = setInterval(async () => {
      try {
        const response = await fetch('/antigravity/poll');
        if (response.ok) {
          const data = await response.json();
          if (data.reply) {
            setMessages(prev => [...prev, { role: 'agent', content: data.reply }]);
          }
        }
      } catch (err) {
        // Ignore poll errors
      }
    }, 2000);
    return () => clearInterval(pollInterval);
  }, []);

  const handleSend = async (e) => {
    e.preventDefault();
    if (!input.trim()) return;

    const userMessage = input.trim();
    setMessages(prev => [...prev, { role: 'user', content: userMessage, metadata: { target: targetAgent } }]);
    setInput('');
    setIsTyping(true);

    try {
      const response = await fetch('/antigravity/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          message: userMessage, 
          sender_id: 'MAX-Dashboard',
          target_agent: targetAgent,
          context_window: contextWindow,
          inject_memory: injectMemory,
          permanent: isPermanent,
          system_critical: isSystemCritical
        })
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const data = await response.json();
      if (data.reply) {
        setMessages(prev => [...prev, { role: 'agent', content: data.reply }]);
      }
    } catch (error) {
      console.warn('Backend not yet connected:', error);
      setTimeout(() => {
        setMessages(prev => [...prev, { role: 'agent', content: `[Backend disconnected] Echo: ${userMessage}` }]);
      }, 500);
    } finally {
      setIsTyping(false);
    }
  };

  return (
    <div className="glass-card side-panel" style={{ flex: 1, display: 'flex', flexDirection: 'column', height: '100%', borderRadius: 0, border: 'none' }}>
      <div className="card-title" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Terminal size={20} /> Command Center
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.75rem', color: 'var(--color-green)' }}>
          <Globe size={14} /> LIVE (0.0.0.0)
        </div>
      </div>

      {/* Advanced Controls Header */}
      <div style={{ background: 'rgba(0,0,0,0.3)', padding: '0.75rem', borderBottom: '1px solid rgba(255,255,255,0.05)', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        
        {/* Target & Context */}
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', background: 'rgba(255,255,255,0.05)', padding: '0.25rem 0.5rem', borderRadius: '4px', flex: 1 }}>
            <Cpu size={14} color="var(--color-blue)" />
            <select 
              value={targetAgent} 
              onChange={e => setTargetAgent(e.target.value)}
              style={{ background: 'transparent', border: 'none', color: 'var(--text-primary)', fontSize: '0.8rem', outline: 'none', width: '100%' }}
            >
              <option value="ALL">BROADCAST ALL</option>
              <option value="TED">TED (192.168.12.110)</option>
              <option value="MAX">MAX (192.168.12.204)</option>
              <option value="ORCHESTRATOR">ORCHESTRATOR SWARM</option>
            </select>
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Layers size={14} color="var(--text-secondary)" />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '2px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.7rem', color: 'var(--text-secondary)' }}>
              <span>Context Size:</span>
              <span>{contextWindow} tokens</span>
            </div>
            <input 
              type="range" 
              min="4096" max="128000" step="4096"
              value={contextWindow}
              onChange={e => setContextWindow(Number(e.target.value))}
              style={{ width: '100%', accentColor: 'var(--color-blue)' }}
            />
          </div>
        </div>

        {/* Memory Toggles */}
        <div style={{ display: 'flex', gap: '0.25rem', flexWrap: 'wrap' }}>
          <button 
            type="button"
            onClick={() => setInjectMemory(!injectMemory)}
            style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.7rem', padding: '0.25rem 0.5rem', borderRadius: '4px', background: injectMemory ? 'rgba(16, 185, 129, 0.2)' : 'rgba(255,255,255,0.05)', border: `1px solid ${injectMemory ? 'var(--color-green)' : 'transparent'}`, color: injectMemory ? 'var(--color-green)' : 'var(--text-secondary)', cursor: 'pointer' }}
          >
            <Database size={12} /> Inject Memory
          </button>
          
          <button 
            type="button"
            disabled={!injectMemory}
            onClick={() => setIsPermanent(!isPermanent)}
            style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.7rem', padding: '0.25rem 0.5rem', borderRadius: '4px', background: isPermanent ? 'rgba(192, 132, 252, 0.2)' : 'rgba(255,255,255,0.05)', border: `1px solid ${isPermanent ? 'var(--color-purple)' : 'transparent'}`, color: isPermanent ? 'var(--color-purple)' : 'var(--text-secondary)', cursor: injectMemory ? 'pointer' : 'not-allowed', opacity: injectMemory ? 1 : 0.4 }}
          >
            <Zap size={12} /> Permanent
          </button>

          <button 
            type="button"
            onClick={() => setIsSystemCritical(!isSystemCritical)}
            style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.7rem', padding: '0.25rem 0.5rem', borderRadius: '4px', background: isSystemCritical ? 'rgba(239, 68, 68, 0.2)' : 'rgba(255,255,255,0.05)', border: `1px solid ${isSystemCritical ? 'var(--color-red)' : 'transparent'}`, color: isSystemCritical ? 'var(--color-red)' : 'var(--text-secondary)', cursor: 'pointer' }}
          >
            <ShieldAlert size={12} /> System Critical
          </button>
        </div>
      </div>
      
      <div 
        style={{ 
          flex: 1, 
          overflowY: 'auto', 
          display: 'flex', 
          flexDirection: 'column', 
          gap: '0.75rem',
          padding: '0.75rem',
        }}
      >
        {messages.map((msg, idx) => (
          <div key={idx} style={{ 
            display: 'flex', 
            gap: '0.5rem', 
            alignItems: 'flex-start',
            flexDirection: msg.role === 'user' ? 'row-reverse' : 'row'
          }}>
            <div style={{
              background: 'rgba(255,255,255,0.05)', 
              padding: '0.4rem', 
              borderRadius: '6px'
            }}>
              {msg.role === 'user' ? <User size={14} color="var(--color-blue)" /> : 
               msg.role === 'system' ? <Terminal size={14} color="var(--color-yellow)" /> :
               <Bot size={14} color="var(--color-green)" />}
            </div>
            <div style={{
              background: msg.role === 'user' ? 'rgba(0, 163, 255, 0.1)' : 'rgba(255,255,255,0.03)',
              border: msg.role === 'user' ? '1px solid rgba(0, 163, 255, 0.2)' : '1px solid rgba(255,255,255,0.05)',
              padding: '0.5rem 0.75rem',
              borderRadius: '8px',
              maxWidth: '85%',
              fontSize: '0.85rem',
              lineHeight: '1.4',
              color: msg.role === 'system' ? 'var(--color-yellow)' : 'var(--text-primary)',
              fontFamily: msg.role === 'system' ? 'monospace' : 'inherit'
            }}>
              {msg.metadata?.target && msg.metadata.target !== 'ALL' && (
                <div style={{ fontSize: '0.65rem', color: 'var(--color-blue)', marginBottom: '4px', textTransform: 'uppercase', fontWeight: 'bold' }}>
                  [{msg.metadata.target}]
                </div>
              )}
              {msg.content}
            </div>
          </div>
        ))}
        {isTyping && (
          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', opacity: 0.5, fontSize: '0.8rem' }}>
            <Bot size={14} color="var(--color-green)" />
            <span>Agent is processing...</span>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <form onSubmit={handleSend} style={{ display: 'flex', gap: '0.5rem', padding: '1rem', borderTop: '1px solid rgba(255,255,255,0.05)', background: 'rgba(0,0,0,0.2)' }}>
        <input 
          type="text" 
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Send command to swarm..." 
          style={{
            flex: 1,
            background: 'rgba(255,255,255,0.05)',
            border: '1px solid rgba(255,255,255,0.1)',
            borderRadius: '6px',
            padding: '0.5rem 0.75rem',
            color: 'var(--text-primary)',
            outline: 'none'
          }}
        />
        <button 
          type="submit"
          disabled={!input.trim() || isTyping}
          style={{
            background: 'var(--color-blue)',
            border: 'none',
            borderRadius: '6px',
            padding: '0.5rem 1rem',
            color: '#fff',
            cursor: input.trim() && !isTyping ? 'pointer' : 'not-allowed',
            opacity: input.trim() && !isTyping ? 1 : 0.5,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}
        >
          <Send size={16} />
        </button>
      </form>
    </div>
  );
}
