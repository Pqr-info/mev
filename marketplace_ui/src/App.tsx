import { useEffect, useState } from 'react';
import { Hexagon, ShoppingCart, Loader2, CheckCircle } from 'lucide-react';
import './App.css';
import { fetchAssets, buyAsset } from './api';
import type { ArtAsset } from './api';

const CURRENT_USER_ADDRESS = '0x5D_MY_ADDRESS_999';

function App() {
  const [assets, setAssets] = useState<ArtAsset[]>([]);
  const [loading, setLoading] = useState(true);
  const [buyingId, setBuyingId] = useState<string | null>(null);
  const [notification, setNotification] = useState<string | null>(null);

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

  const handleBuy = async (asset: ArtAsset) => {
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

  const showNotification = (msg: string) => {
    setNotification(msg);
    setTimeout(() => setNotification(null), 4000);
  };

  return (
    <div className="app-container">
      <header className="header">
        <div className="logo">
          <Hexagon size={28} color="var(--accent-color)" />
          ImageFX <span>Marketplace</span>
        </div>
        <div className="user-identity">
          <span className="label">5D Identity:</span>
          <span className="address">{CURRENT_USER_ADDRESS}</span>
        </div>
      </header>

      <main className="main-content">
        <section className="hero">
          <h1>Discover Sovereign Art</h1>
          <p>
            The canonical 5D-ASP marketplace for trading digital assets. 
            All state transitions are securely zetafolded.
          </p>
        </section>

        {loading ? (
          <div className="loader">
            <Loader2 size={48} className="spin" />
            <p>Syncing state from :9085...</p>
          </div>
        ) : (
          <div className="grid">
            {assets.map((asset) => (
              <div key={asset.id} className="card">
                <img src={asset.imageUrl} alt={asset.title} className="card-image" loading="lazy" />
                <div className="card-content">
                  <div className="card-header">
                    <div>
                      <h3 className="card-title">{asset.title}</h3>
                      <p className="card-artist">by {asset.artist}</p>
                    </div>
                    <span className="card-price">{asset.price} ZET</span>
                  </div>
                  
                  <div className="card-footer">
                    <span className="owner-badge">
                      {asset.ownerAddress === CURRENT_USER_ADDRESS ? 'Owned by You' : `Owner: ${asset.ownerAddress.substring(0, 10)}...`}
                    </span>
                    
                    <button 
                      className={`btn ${asset.ownerAddress === CURRENT_USER_ADDRESS ? 'btn-success' : 'btn-primary'}`}
                      disabled={buyingId === asset.id || asset.ownerAddress === CURRENT_USER_ADDRESS}
                      onClick={() => handleBuy(asset)}
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

      <footer className="footer">
        <p>&copy; 2026 ImageFX &bull; Powered by 5D-ASP Sovereign-27 Network</p>
      </footer>

      {notification && (
        <div className="notification">
          <CheckCircle size={20} color="var(--success-color)" />
          {notification}
        </div>
      )}
    </div>
  );
}

export default App;
