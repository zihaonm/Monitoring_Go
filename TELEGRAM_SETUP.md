# Telegram Notification Setup Guide

This guide will help you set up Telegram notifications for your monitoring service.

## Quick Start

### Step 1: Create a Telegram Bot

1. Open Telegram and search for **@BotFather**
2. Start a chat and send the command: `/newbot`
3. Follow the prompts:
   - Choose a name for your bot (e.g., "My Monitoring Bot")
   - Choose a username (must end in 'bot', e.g., "mymonitoring_bot")
4. BotFather will give you a **Bot Token** - save this! It looks like:
   ```
   123456789:ABCdefGHIjklMNOpqrsTUVwxyz
   ```

### Step 2: Get Your Chat ID

#### Option A: Personal Chat (recommended for testing)

1. Search for **@userinfobot** in Telegram
2. Start a chat and it will immediately show your user ID
3. Your Chat ID is the number shown (e.g., `123456789`)

#### Option B: Group Chat (recommended for teams)

1. Create a new Telegram group
2. Add your bot to the group:
   - Click on group name → Add members → Search for your bot username
3. Send any message in the group
4. Open this URL in your browser (replace `<YOUR_BOT_TOKEN>` with your actual token):
   ```
   https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates
   ```
5. Look for the `"chat":{"id":` field in the JSON response
6. The Chat ID will be negative for groups (e.g., `-1001234567890`)

### Step 3: Configure in Dashboard

1. Open your monitoring dashboard at http://localhost:8080
2. Find the "Telegram Notifications" section at the top
3. Enter your **Bot Token** and **Chat ID**
4. Check the **Enable** checkbox
5. Click **Save Config**
6. Click **Test** to send a test message

If you receive the test message in Telegram, you're all set!

## How It Works

Once configured, the monitoring service will automatically send Telegram messages:

### Down Alert (🔴)
Sent when a monitored service goes down:
```
🔴 Service Down Alert

Service: My API
URL: https://api.example.com
Status: DOWN
Error: connection timeout
Time: 2025-10-22 10:30:45
```

### Recovery Alert (🟢)
Sent when a service comes back online:
```
🟢 Service Recovered

Service: My API
URL: https://api.example.com
Status: UP
Response Time: 125ms
Time: 2025-10-22 10:35:12
```

## Troubleshooting

### Test message fails

1. **Check Bot Token**: Make sure you copied the entire token from BotFather
2. **Check Chat ID**:
   - For personal chats, should be a positive number
   - For groups, should be a negative number starting with `-100`
3. **Bot in Group**: If using a group, make sure the bot is actually added as a member
4. **Bot Permissions**: Ensure the bot hasn't been blocked or removed

### Not receiving down/up alerts

1. **Enable checkbox**: Make sure "Enable" is checked in the dashboard
2. **Save Config**: Click "Save Config" after making changes
3. **Service Status**: Check that services are actually changing status
4. **Check logs**: Look at the server console for any Telegram-related errors

### Getting errors in console

Common errors and solutions:

- **"Unauthorized"**: Bot token is incorrect
- **"Bad Request: chat not found"**: Chat ID is incorrect or bot not in group
- **"Forbidden: bot was blocked by the user"**: You need to unblock the bot in Telegram

## API Usage

You can also configure Telegram via API:

### Get current config
```bash
curl http://localhost:8080/api/telegram/config
```

### Update config
```bash
curl -X PUT http://localhost:8080/api/telegram/config \
  -H "Content-Type: application/json" \
  -d '{
    "bot_token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
    "chat_id": "-1001234567890",
    "enabled": true
  }'
```

### Send test notification
```bash
curl -X POST http://localhost:8080/api/telegram/test
```

## Security Notes

- The bot token is sensitive - don't share it publicly
- When retrieving config via API, the bot token is partially masked for security
- Store credentials securely if using in production
- Consider using environment variables for production deployments

## Tips

- Use a dedicated monitoring group for team notifications
- Give the bot a descriptive name so you know what alerts are from
- You can mute the group but still get desktop notifications
- Multiple people can be in the same group to receive alerts

## Need Help?

- Telegram Bot API Docs: https://core.telegram.org/bots/api
- BotFather Commands: https://core.telegram.org/bots#botfather
