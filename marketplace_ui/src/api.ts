export interface ArtAsset {
  id: string;
  title: string;
  artist: string;
  price: number;
  imageUrl: string;
  ownerAddress: string;
  description: string;
}

export interface StateSnapshot {
  timestamp: string;
  action: 'BUY' | 'SELL' | 'MINT' | 'TRANSFER';
  assetId: string;
  fromAddress?: string;
  toAddress: string;
  metadata?: any;
}

// Mock initial data
let mockAssets: ArtAsset[] = [
  {
    id: 'art-001',
    title: 'Neon Zenith',
    artist: 'CyberPunk_5D',
    price: 15.5,
    imageUrl: 'https://images.unsplash.com/photo-1614850523459-c2f4c699c52e?auto=format&fit=crop&q=80&w=800',
    ownerAddress: '0x5D_A1B2C3D4E5',
    description: 'A mesmerizing glimpse into the neon-lit future of sovereign nodes.'
  },
  {
    id: 'art-002',
    title: 'Abstract Zeta',
    artist: 'NodeWeaver',
    price: 8.0,
    imageUrl: 'https://images.unsplash.com/photo-1549490349-8643362247b5?auto=format&fit=crop&q=80&w=800',
    ownerAddress: '0x5D_F9E8D7C6B5',
    description: 'Geometric abstract representations of the Zetafolded graph.'
  },
  {
    id: 'art-003',
    title: 'Ethereal Bound',
    artist: 'QuantumCanvas',
    price: 42.0,
    imageUrl: 'https://images.unsplash.com/photo-1574169208507-84376144848b?auto=format&fit=crop&q=80&w=800',
    ownerAddress: '0x5D_X1Y2Z3W4V5',
    description: 'Surrealist expression of identity within the 5D-ASP framework.'
  },
  {
    id: 'art-004',
    title: 'Digital Genesis',
    artist: 'Origin_00',
    price: 100.0,
    imageUrl: 'https://images.unsplash.com/photo-1563089145-599997674d42?auto=format&fit=crop&q=80&w=800',
    ownerAddress: '0x5D_G7H8I9J0K1',
    description: 'The birth of a new artistic era on the Sovereign-27 network.'
  }
];

export const fetchAssets = async (): Promise<ArtAsset[]> => {
  return new Promise((resolve) => setTimeout(() => resolve(mockAssets), 800));
};

export const submitStateSnapshot = async (payload: StateSnapshot): Promise<void> => {
  console.log(`[5D-ASP Logging] Sending StateSnapshot to :9085...`, payload);
  
  // Simulate network request to :9085 backend
  return new Promise((resolve) => {
    setTimeout(() => {
      // Mocking the fetch
      console.log(`[5D-ASP Logging] Successfully logged state for asset ${payload.assetId}`);
      resolve();
    }, 500);
  });
};

export const buyAsset = async (assetId: string, buyerAddress: string): Promise<boolean> => {
  const asset = mockAssets.find(a => a.id === assetId);
  if (!asset) return false;
  
  const snapshot: StateSnapshot = {
    timestamp: new Date().toISOString(),
    action: 'BUY',
    assetId: asset.id,
    fromAddress: asset.ownerAddress,
    toAddress: buyerAddress,
    metadata: { price: asset.price }
  };
  
  try {
    await submitStateSnapshot(snapshot);
    // Update local mock state
    asset.ownerAddress = buyerAddress;
    return true;
  } catch (e) {
    console.error("Failed to update state on :9085 backend", e);
    return false;
  }
};
