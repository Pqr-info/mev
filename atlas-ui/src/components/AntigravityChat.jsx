import React, { useState, useRef, useEffect } from 'react';
import { Terminal, Send, Bot, User } from 'lucide-react';

export default function AntigravityChat() {
  const [messages, setMessages] = useState([
    { role: 'system', content: 'Antigravity gRPC Interface Initialized.' },
    { role: 'agent', content: 'Hello! I am connected to the Organ Atlas dashboard. How can I help you manage the mesh?' }
  ]);
  const [input, setInput] = useState('');
  const [isTyping, setIsTyping] = useState(false);
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
    setMessages(prev => [...prev, { role: 'user', content: userMessage }]);
    setInput('');
    setIsTyping(true);

    try {
      // Stub HTTP POST request. This will route to the Go backend bridge eventually.
      const response = await fetch('/antigravity/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: userMessage, sender_id: 'MAX-Dashboard' })
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const data = await response.json();
      setMessages(prev => [...prev, { role: 'agent', content: data.reply || 'Message received.' }]);
    } catch (error) {
      console.warn('Backend not yet connected:', error);
      // Mock response for UI testing until backend is built
      setTimeout(() => {
        setMessages(prev => [...prev, { role: 'agent', content: `[Backend disconnected] Echo: ${userMessage}` }]);
      }, 500);
    } finally {
      setIsTyping(false);
    }
  };

  return (
    <div className="glass-card side-panel" style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
      <div className="card-title">
        <Terminal size={20} /> Antigravity Link
      </div>
      
      <div 
        style={{ 
          flex: 1, 
          overflowY: 'auto', 
          display: 'flex', 
          flexDirection: 'column', 
          gap: '0.75rem',
          padding: '0.5rem',
          maxHeight: '300px'
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
              {msg.content}
            </div>
          </div>
        ))}
        {isTyping && (
          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', opacity: 0.5, fontSize: '0.8rem' }}>
            <Bot size={14} color="var(--color-green)" />
            <span>Agent is typing...</span>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <form onSubmit={handleSend} style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem', borderTop: '1px solid rgba(255,255,255,0.05)', paddingTop: '1rem' }}>
        <input 
          type="text" 
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Send message to Antigravity..." 
          style={{
            flex: 1,
            background: 'rgba(0,0,0,0.2)',
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
            padding: '0.5rem',
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
