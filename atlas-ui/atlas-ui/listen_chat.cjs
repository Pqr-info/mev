const Redis = require('ioredis');

const redis = new Redis('127.0.0.1:6379');

redis.subscribe('mesh:chat', (err, count) => {
  if (err) {
    console.error('Failed to subscribe: %s', err.message);
  } else {
    console.log(`Agent Chat Listener Connected. Waiting for messages...`);
  }
});

redis.on('message', (channel, message) => {
  if (channel === 'mesh:chat') {
    try {
      const data = JSON.parse(message);
      console.log(`\n[HITL UI Message from ${data.sender}]: ${data.message}\n`);
    } catch (e) {
      console.log(`\n[Raw Message]: ${message}\n`);
    }
  }
});
