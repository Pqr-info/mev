/**
 * Dolphin Whitelabel Client - Rich Mock Dataset
 * High-fidelity community data for offline demo & rapid preview
 */

export const INITIAL_STORIES = [
  {
    id: 's_add',
    isAdd: true,
    name: 'Your Pod',
    avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=250&q=80'
  },
  {
    id: 's1',
    name: 'Sarah Jenkins',
    avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=250&q=80',
    unseen: true,
    slides: [
      { type: 'image', url: 'https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=600&q=80', caption: 'Sunrise ocean breeze 🌊' }
    ]
  },
  {
    id: 's2',
    name: 'Alex Rivera',
    avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=250&q=80',
    unseen: true,
    slides: [
      { type: 'image', url: 'https://images.unsplash.com/photo-1518770660439-4636190af475?auto=format&fit=crop&w=600&q=80', caption: 'Building new Android AI stack 🚀' }
    ]
  },
  {
    id: 's3',
    name: 'Elena Rostova',
    avatar: 'https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&w=250&q=80',
    unseen: false,
    slides: [
      { type: 'image', url: 'https://images.unsplash.com/photo-1470071459604-3b5ec3a7fe05?auto=format&fit=crop&w=600&q=80', caption: 'Mist over mountain peaks 🏔️' }
    ]
  },
  {
    id: 's4',
    name: 'Dolphin Tech',
    avatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&w=250&q=80',
    unseen: true,
    slides: [
      { type: 'image', url: 'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&w=600&q=80', caption: 'UNA Dolphin v2.0 REST API Ready!' }
    ]
  }
];

export const INITIAL_POSTS = [
  {
    id: 'p101',
    author: {
      id: 'u1',
      name: 'Dr. Marina Vance',
      username: '@marina_vance',
      avatar: 'https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?auto=format&fit=crop&w=250&q=80',
      verified: true
    },
    content: '🐬 Exciting update! Our Dolphin Whitelabel Android App client is now operational with full custom theme switching, real-time community feed, and serverless offline store!',
    media: 'https://images.unsplash.com/photo-1518770660439-4636190af475?auto=format&fit=crop&w=800&q=80',
    timestamp: '2 hours ago',
    location: 'Pacific Innovation Hub',
    reactions: {
      dolphin: 42,
      like: 89,
      userReacted: { dolphin: true, like: false }
    },
    comments: [
      {
        id: 'c1',
        user: 'Alex Rivera',
        avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=250&q=80',
        text: 'The dynamic theme customizer works smoothly! Love the ocean gradient color tokens.',
        timestamp: '1h ago'
      }
    ]
  },
  {
    id: 'p102',
    author: {
      id: 'u2',
      name: 'Sarah Jenkins',
      username: '@sjenkins',
      avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=250&q=80',
      verified: false
    },
    content: 'Building proximity-aware Android applications for UNA/Dolphin. Geofenced pods and offline local state persistence unlock true decentralization. 📍',
    media: null,
    timestamp: '4 hours ago',
    location: 'San Francisco, CA',
    reactions: {
      dolphin: 18,
      like: 34,
      userReacted: { dolphin: false, like: true }
    },
    comments: []
  }
];

export const INITIAL_GROUPS = [
  {
    id: 'g1',
    name: 'Dolphin Core Developers',
    category: 'Engineering & Tech',
    cover: 'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&w=600&q=80',
    membersCount: 1420,
    description: 'Official developer community building whitelabel mobile clients for Dolphin & UNA platform.'
  },
  {
    id: 'g2',
    name: 'Oceanic AI & Mesh Network',
    category: 'Research & Innovation',
    cover: 'https://images.unsplash.com/photo-1451187580459-43490279c0fa?auto=format&fit=crop&w=600&q=80',
    membersCount: 890,
    description: 'Decentralized peer-to-peer mobile networking and edge intelligence.'
  }
];

export const INITIAL_CHATS = [
  {
    id: 'chat_1',
    user: {
      name: 'Sarah Jenkins',
      avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=250&q=80',
      online: true
    },
    lastMessage: 'Hey! Did you check out the new Whitelabel Studio modal?',
    time: '10:42 AM',
    unread: 2,
    messages: [
      { id: 'm1', sender: 'them', text: 'Hey! Did you check out the new Whitelabel Studio modal?', time: '10:42 AM' }
    ]
  },
  {
    id: 'chat_2',
    user: {
      name: 'Alex Rivera',
      avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=250&q=80',
      online: false
    },
    lastMessage: 'Let me know when the APK build is uploaded.',
    time: 'Yesterday',
    unread: 0,
    messages: [
      { id: 'm10', sender: 'them', text: 'Let me know when the APK build is uploaded.', time: '09:00 AM' }
    ]
  },
  {
    id: 'chat_3',
    user: {
      name: 'Dr. Marina Vance',
      avatar: 'https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?auto=format&fit=crop&w=250&q=80',
      online: true
    },
    lastMessage: '🐬 Shared the new Whitelabel preview link with the team!',
    time: '3h ago',
    unread: 0,
    messages: [
      { id: 'm20', sender: 'them', text: '🐬 Shared the new Whitelabel preview link with the team!', time: '07:30 AM' }
    ]
  },
  {
    id: 'chat_gemma_4_e4b',
    modelId: 'google/gemma-4-e4b',
    user: {
      name: 'Local Gemma 4-e4b AI',
      avatar: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=250&q=80',
      online: true
    },
    lastMessage: '⚡ Ready with 49-Ticket Context Matrix & Agentic RAG!',
    time: 'Live',
    unread: 1,
    messages: [
      { id: 'm_gemma1', sender: 'them', text: 'Hello Sovereign-27 Node! I am Gemma 4-e4b running locally with full 49-ticket context matrix and RAG memory integration.', time: 'Live' }
    ]
  },
  {
    id: 'chat_qwen3_30b',
    modelId: 'qwen/qwen3-coder-30b',
    user: {
      name: 'Local Qwen 3-Coder-30b AI',
      avatar: 'https://images.unsplash.com/photo-1620712943543-bcc4688e7485?auto=format&fit=crop&w=250&q=80',
      online: true
    },
    lastMessage: '🧠 Ready for code generation & Substrate 27 queries.',
    time: 'Live',
    unread: 0,
    messages: [
      { id: 'm_qwen1', sender: 'them', text: 'Greetings! Qwen 3-Coder-30b is online on LM Studio. Ask me anything about your Rust, Go, or Substrate 27 stack.', time: 'Live' }
    ]
  },
  {
    id: 'chat_midi_engine',
    isMidi: true,
    user: {
      name: 'MIDI State Machine Engine',
      avatar: 'https://images.unsplash.com/photo-1511671782779-c97d3d27a1d4?auto=format&fit=crop&w=250&q=80',
      online: true
    },
    lastMessage: '🎹 16-Channel In-Memory Engine & CockroachDB Sync (zeta.mh)',
    time: 'Live',
    unread: 0,
    messages: [
      { id: 'm_midi1', sender: 'them', text: 'MIDI State Machine active: 120.0 BPM, 16 Channels, SQLite WAL + CockroachDB Bridge on zeta.mh.', time: 'Live' }
    ]
  },
  {
    id: 'chat_marcus',
    user: {
      name: 'Marcus Chen',
      avatar: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=250&q=80',
      online: true
    },
    lastMessage: 'Sovereign-27 Proximity Mesh connection established ⚡',
    time: 'Just now',
    unread: 0,
    messages: [
      { id: 'm_marcus_1', sender: 'them', text: 'Greetings 5D Node! Connected via local proximity radar! (488m away) 📍', time: 'Just now' }
    ]
  }
];


export const INITIAL_NOTIFICATIONS = [
  {
    id: 'n1',
    user: 'Sarah Jenkins',
    avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=250&q=80',
    type: 'dolphin',
    text: 'splashed 🐬 your post: "Dolphin Whitelabel Android App"',
    time: '10m ago',
    read: false
  },
  {
    id: 'n2',
    user: 'Alex Rivera',
    avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=250&q=80',
    type: 'comment',
    text: 'commented: "This dark glassmorphism looks incredibly sleek!"',
    time: '25m ago',
    read: false
  },
  {
    id: 'n3',
    user: 'Dolphin Core Developers',
    avatar: 'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&w=250&q=80',
    type: 'group',
    text: 'approved your request to join Dolphin Core Developers',
    time: '2h ago',
    read: true
  }
];

// Alias exports for offlineStore module compatibility
export const SEED_POSTS = INITIAL_POSTS;
export const SEED_STORIES = INITIAL_STORIES;
export const SEED_CHATS = INITIAL_CHATS;
export const SEED_GROUPS = INITIAL_GROUPS;
export const SEED_NOTIFICATIONS = INITIAL_NOTIFICATIONS;
