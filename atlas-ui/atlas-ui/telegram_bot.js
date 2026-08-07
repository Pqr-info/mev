import 'dotenv/config';
import TelegramBot from 'node-telegram-bot-api';
import { exec } from 'child_process';

const token = process.env.TELEGRAM_BOT_TOKEN || 'YOUR_TELEGRAM_BOT_TOKEN_HERE';

if (token === 'YOUR_TELEGRAM_BOT_TOKEN_HERE') {
  console.error("CRITICAL ERROR: You must provide a valid TELEGRAM_BOT_TOKEN environment variable.");
  process.exit(1);
}

const CONVERSATION_ID = 'c38354eb-749f-46a1-b32e-3c1d1de35c82';

const bot = new TelegramBot(token, { polling: true });

console.log('Antigravity Telegram Bot is running...');

bot.on('message', (msg) => {
  const chatId = msg.chat.id;
  const text = msg.text;

  if (!text) return;

  console.log(`[Telegram] Received message from ${msg.from.username || msg.from.first_name}: ${text}`);

  const payload = {
    message: text,
    sender: msg.from.username || msg.from.first_name
  };

  bot.sendMessage(chatId, `⏳ Injecting command into Antigravity Mesh...`);

  fetch('http://127.0.0.1:8196/REST/2.0/chat/antigravity', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })
  .then(res => res.json())
  .then(data => {
    console.log(`[Telegram] Injected successfully.`, data);
    bot.sendMessage(chatId, `✅ Command injected. TED is processing your request.`);
  })
  .catch(error => {
    console.error(`Error executing API: ${error}`);
    bot.sendMessage(chatId, `❌ Error injecting command: ${error.message}`);
  });
});
