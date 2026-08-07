const express = require('express');
const cors = require('cors');
const Redis = require('ioredis');

const app = express();
app.use(cors());
app.use(express.json());

const redis = new Redis(process.env.VALKEY_ADDR || '127.0.0.1:6379');

// Native Webhook API Endpoint for true context injection
app.post('/api/v1/antigravity/message', async (req, res) => {
  try {
    const { message, sender } = req.body;
    if (!message) return res.status(400).json({ error: 'Message is required' });

    const payload = {
      id: Date.now().toString(),
      content: message,
      sender: sender || 'Telegram-Bot',
      target: 'broadcast',
      contextSize: 'default',
      flags: {}
    };
    await redis.publish('mesh:chat', JSON.stringify(payload));
    res.json({ status: 'ok', detail: 'Injected into Antigravity queue successfully' });
  } catch (error) {
    console.error(`Error publishing message: ${error}`);
    res.status(500).json({ error: 'Failed to inject message into mesh' });
  }
});

// Endpoint to post messages to the chat
app.post('/antigravity/chat', async (req, res) => {
  try {
    const { message, sender_id, target_agent, context_window, inject_memory, permanent, system_critical } = req.body;
    const payload = JSON.stringify({ 
      sender: sender_id || 'UI', 
      message,
      target_agent: target_agent || 'ALL',
      context_window: context_window || 8192,
      inject_memory: !!inject_memory,
      permanent: !!permanent,
      system_critical: !!system_critical
    });
    
    // Publish to pub/sub for real-time listeners (like TED and Max)
    await redis.publish('mesh:chat', payload);
    
    res.json({ status: 'ok' });
  } catch (error) {
    console.error('Error in /antigravity/chat:', error);
    res.status(500).json({ error: 'Internal Server Error' });
  }
});

// Endpoint for UI to poll for new messages directed to it
app.get('/antigravity/poll', async (req, res) => {
  try {
    // The agents will push their replies to 'mesh:chat:ui_inbox'.
    const msg = await redis.rpop('mesh:chat:ui_inbox');
    if (msg) {
      const data = JSON.parse(msg);
      res.json({ reply: `${data.sender}: ${data.message}` });
    } else {
      res.json({});
    }
  } catch (error) {
    console.error('Error in /antigravity/poll:', error);
    res.status(500).json({ error: 'Internal Server Error' });
  }
});

const port = process.env.PORT || 3457;
app.listen(port, '0.0.0.0', () => {
  console.log(`Antigravity Bridge running on 0.0.0.0:${port}`);
});
