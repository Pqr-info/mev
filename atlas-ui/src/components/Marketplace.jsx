import React, { useEffect, useState } from 'react';
import { Hexagon, ShoppingCart, Loader2, CheckCircle } from 'lucide-react';
import { fetchAssets, buyAsset } from './api';

const CURRENT_USER_ADDRESS = '0x5D_MY_ADDRESS_999';

export default function Marketplace() {
  const [assets, setAssets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [buyingId, setBuyingId] = useState(null);
  const [notification, setNotification] = useState(null);

  useEffect(() => {
    loadAssets();
  }, []);

  const loadAssets = async () => {
    setLoading(true);
    try {
      const data = await fetchAssets();
      setAssets(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleBuy = async (asset) => {
    if (asset.ownerAddress === CURRENT_USER_ADDRESS) return;
    
    setBuyingId(asset.id);
    const success = await buyAsset(asset.id, CURRENT_USER_ADDRESS);
    
    if (success) {
      // Optimistic update
      setAssets(assets.map(a => 
        a.id === asset.id ? { ...a, ownerAddress: CURRENT_USER_ADDRESS } : a
      ));
      
      showNotification(`Successfully purchased ${asset.title}! State logged to 5D-ASP.`);
    } else {
      showNotification('Purchase failed.');
    }
    
    setBuyingId(null);
  };

  const showNotification = (msg) => {
    setNotification(msg);
    setTimeout(() => setNotification(null), 4000);
  };

  return (
    <div style={{ padding: '2rem', height: '100%', overflowY: 'auto' }}>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '1.5rem', fontWeight: 700 }}>
          <Hexagon size={28} color="var(--color-blue)" />
          ImageFX <span style={{ fontWeight: 300 }}>Marketplace</span>
        </div>
        <div style={{ background: 'rgba(255,255,255,0.05)', padding: '0.5rem 1rem', borderRadius: '8px' }}>
          <span style={{ color: 'var(--text-secondary)', marginRight: '0.5rem' }}>5D Identity:</span>
          <span style={{ fontFamily: 'monospace' }}>{CURRENT_USER_ADDRESS}</span>
        </div>
      </header>

      <main>
        <section style={{ marginBottom: '3rem', textAlign: 'center' }}>
          <h1 style={{ fontSize: '2.5rem', marginBottom: '1rem' }}>Discover Sovereign Art</h1>
          <p style={{ color: 'var(--text-secondary)', maxWidth: '600px', margin: '0 auto' }}>
            The canonical 5D-ASP marketplace for trading digital assets. 
            All state transitions are securely zetafolded.
          </p>
        </section>

        {loading ? (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
            <Loader2 size={48} className="spin" style={{ marginBottom: '1rem' }} />
            <p>Syncing state from :9085...</p>
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '2rem' }}>
            {assets.map((asset) => (
              <div key={asset.id} className="glass-card" style={{ overflow: 'hidden', padding: 0 }}>
                <img 
                  src={asset.imageUrl} 
                  alt={asset.title} 
                  style={{ width: '100%', height: '250px', objectFit: 'cover' }} 
                  loading="lazy" 
                />
                <div style={{ padding: '1.5rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
                    <div>
                      <h3 style={{ fontSize: '1.25rem', marginBottom: '0.25rem' }}>{asset.title}</h3>
                      <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>by {asset.artist}</p>
                    </div>
                    <span style={{ fontSize: '1.25rem', fontWeight: 600, color: 'var(--color-blue)' }}>{asset.price} ZET</span>
                  </div>
                  
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1.5rem' }}>
                    <span style={{ fontSize: '0.75rem', background: 'rgba(255,255,255,0.1)', padding: '0.25rem 0.5rem', borderRadius: '4px' }}>
                      {asset.ownerAddress === CURRENT_USER_ADDRESS ? 'Owned by You' : `Owner: ${asset.ownerAddress.substring(0, 10)}...`}
                    </span>
                    
                    <button 
                      disabled={buyingId === asset.id || asset.ownerAddress === CURRENT_USER_ADDRESS}
                      onClick={() => handleBuy(asset)}
                      style={{
                        background: asset.ownerAddress === CURRENT_USER_ADDRESS ? 'rgba(16, 185, 129, 0.2)' : 'var(--color-blue)',
                        color: asset.ownerAddress === CURRENT_USER_ADDRESS ? 'var(--color-green)' : '#fff',
                        border: asset.ownerAddress === CURRENT_USER_ADDRESS ? '1px solid var(--color-green)' : 'none',
                        padding: '0.5rem 1rem',
                        borderRadius: '6px',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        cursor: asset.ownerAddress === CURRENT_USER_ADDRESS ? 'default' : 'pointer',
                        opacity: buyingId === asset.id ? 0.7 : 1
                      }}
                    >
                      {buyingId === asset.id ? (
                        <><Loader2 size={16} className="spin" /> Processing...</>
                      ) : asset.ownerAddress === CURRENT_USER_ADDRESS ? (
                        <><CheckCircle size={16} /> Owned</>
                      ) : (
                        <><ShoppingCart size={16} /> Buy Now</>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      {notification && (
        <div style={{
          position: 'fixed', bottom: '2rem', left: '50%', transform: 'translateX(-50%)',
          background: 'rgba(16, 185, 129, 0.2)', border: '1px solid var(--color-green)',
          color: '#fff', padding: '1rem 2rem', borderRadius: '8px', display: 'flex', alignItems: 'center', gap: '0.5rem',
          backdropFilter: 'blur(10px)', zIndex: 100, boxShadow: '0 4px 12px rgba(0,0,0,0.5)'
        }}>
          <CheckCircle size={20} color="var(--color-green)" />
          {notification}
        </div>
      )}
    </div>
  );
}
