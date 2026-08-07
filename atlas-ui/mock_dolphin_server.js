import express from 'express';
import cors from 'cors';
import morgan from 'morgan';

// We import the mock data from the frontend to keep the same state base
import { 
  INITIAL_POSTS, 
  INITIAL_STORIES, 
  INITIAL_CHATS, 
  INITIAL_GROUPS, 
  INITIAL_NOTIFICATIONS 
} from './src/portal-src/services/mockData.js';

const app = express();
const PORT = 4000;

app.use(cors());
app.use(express.json());
app.use(morgan('dev'));

// In-memory data store for the mock server
const store = {
  posts: [...INITIAL_POSTS],
  stories: [...INITIAL_STORIES],
  chats: [...INITIAL_CHATS],
  groups: [...INITIAL_GROUPS],
  notifications: [...INITIAL_NOTIFICATIONS],
};

app.post('/', (req, res) => {
  const queryR = req.query.r;
  const params = req.body || {};

  console.log(`[Dolphin Mock] Received request for: ${queryR}`);

  switch (queryR) {
    case 'bx_posts/get_posts':
      return res.json({ status: 'success', data: store.posts });
      
    case 'bx_posts/entity_add':
      const newPost = {
        id: `p_${Date.now()}`,
        author: {
          id: 'u_mock',
          name: 'Sovereign Node',
          username: '@sovereign_node',
          avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=250&q=80',
          verified: true
        },
        time: 'Just now',
        locationTag: '📍 Server-side Mock',
        content: params.content || '',
        media: params.media || null,
        mediaType: params.media ? 'image' : null,
        stats: { likes: 0, dolphins: 0, comments: 0, shares: 0 },
        userReacted: { like: false, dolphin: false },
        comments: []
      };
      store.posts.unshift(newPost);
      return res.json({ status: 'success', data: newPost });

    case 'bx_posts/react':
      const post = store.posts.find(p => p.id === params.postId);
      if (post) {
        if (params.type === 'dolphin') {
          post.userReacted.dolphin = !post.userReacted.dolphin;
          post.stats.dolphins += post.userReacted.dolphin ? 1 : -1;
        } else {
          post.userReacted.like = !post.userReacted.like;
          post.stats.likes += post.userReacted.like ? 1 : -1;
        }
      }
      return res.json({ status: 'success', data: post });

    case 'bx_comments/add':
      const targetPost = store.posts.find(p => p.id === params.postId);
      if (targetPost) {
        const newComment = {
          id: `c_${Date.now()}`,
          author: 'Sovereign Node',
          avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=250&q=80',
          text: params.text,
          time: 'Just now'
        };
        targetPost.comments.push(newComment);
        targetPost.stats.comments = targetPost.comments.length;
      }
      return res.json({ status: 'success', data: targetPost });

    case 'bx_persons/get_stories':
      return res.json({ status: 'success', data: store.stories });

    case 'bx_messenger/get_chats':
      return res.json({ status: 'success', data: store.chats });

    case 'bx_messenger/send_message':
      const chat = store.chats.find(c => c.id === params.chatId);
      if (chat) {
        const newMsg = {
          id: `m_${Date.now()}`,
          sender: 'me',
          text: params.text,
          time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        };
        chat.messages.push(newMsg);
        chat.lastMessage = params.text;
        chat.time = 'Just now';
      }
      return res.json({ status: 'success', data: chat });

    case 'bx_groups/get_groups':
      return res.json({ status: 'success', data: store.groups });

    case 'bx_notifications/get_all':
      return res.json({ status: 'success', data: store.notifications });

    default:
      return res.json({ status: 'error', message: `Unknown module/method: ${queryR}` });
  }
});

// GET fallback in case some endpoints use GET
app.get('/', (req, res) => {
  const queryR = req.query.r;
  console.log(`[Dolphin Mock] Received GET request for: ${queryR}`);
  // Most UNA endpoints are POST, but we handle GET by routing it similarly for get_* methods
  if (queryR === 'bx_posts/get_posts') return res.json({ status: 'success', data: store.posts });
  if (queryR === 'bx_persons/get_stories') return res.json({ status: 'success', data: store.stories });
  if (queryR === 'bx_messenger/get_chats') return res.json({ status: 'success', data: store.chats });
  if (queryR === 'bx_groups/get_groups') return res.json({ status: 'success', data: store.groups });
  if (queryR === 'bx_notifications/get_all') return res.json({ status: 'success', data: store.notifications });
  
  res.json({ status: 'error', message: 'Use POST for mutations, or unknown GET endpoint.' });
});

// Add a standard health endpoint
app.get('/api/health', (req, res) => {
  res.json({ status: 'ok', server: 'Mock Dolphin/UNA Backend', mode: 'LIVE' });
});

app.listen(PORT, () => {
  console.log(`[Dolphin Mock] E2E Server listening on port ${PORT}`);
  console.log(`[Dolphin Mock] Portal connection Mode: 'LIVE' should point to http://localhost:${PORT}`);
});
