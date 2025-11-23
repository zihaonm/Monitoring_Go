# Data Persistence Guide

Your monitoring tool now automatically saves all data to a JSON file!

## What Gets Saved

All your data is stored in `monitoring_data.json`:

1. **Services**:
   - Service ID, name, and URL
   - Check intervals and timeouts
   - Current status (up/down/unknown)
   - Last check time
   - Response times
   - Error messages
   - Creation timestamps

2. **Telegram Configuration**:
   - Bot token (saved securely)
   - Chat ID
   - Enabled/disabled status

## How It Works

### Auto-Save
Data is automatically saved whenever:
- ✅ You add a new service
- ✅ You update a service
- ✅ You delete a service
- ✅ Health checks update service status
- ✅ You change Telegram configuration

### Auto-Load
When you start the application:
- All services are restored with their last known status
- Telegram configuration is restored
- Monitoring continues from where it left off

## File Location

The data file is created in the same directory as your application:
```
/Users/zihao/pt/Monitoring/monitoring_data.json
```

## Example Data Structure

```json
{
  "services": {
    "uuid-here": {
      "id": "uuid-here",
      "name": "Google",
      "url": "https://google.com",
      "check_interval": 60,
      "timeout": 10,
      "status": "up",
      "last_check": "2025-10-22T10:30:00Z",
      "last_uptime": "2025-10-22T10:30:00Z",
      "last_downtime": "0001-01-01T00:00:00Z",
      "response_time": 124,
      "error_message": "",
      "created_at": "2025-10-22T09:00:00Z"
    }
  },
  "telegram_config": {
    "bot_token": "123456:ABC-DEF...",
    "chat_id": "-1001234567890",
    "enabled": true
  }
}
```

## Common Operations

### Backup Your Data

Simply copy the file:
```bash
cp monitoring_data.json monitoring_data.backup.json
```

Or with timestamp:
```bash
cp monitoring_data.json monitoring_data.$(date +%Y%m%d_%H%M%S).json
```

### Restore from Backup

1. Stop the application (Ctrl+C)
2. Replace the file:
   ```bash
   cp monitoring_data.backup.json monitoring_data.json
   ```
3. Restart the application:
   ```bash
   go run main.go
   ```

### Reset Everything

Delete the data file and restart:
```bash
rm monitoring_data.json
go run main.go
```

The application will start fresh with no services or configuration.

### Migrate to Another Server

1. Copy `monitoring_data.json` from old server
2. Place it in the application directory on new server
3. Start the application
4. All services and configuration will be restored

## Best Practices

### Regular Backups

Set up a cron job for automatic backups:
```bash
# Add to crontab (every hour)
0 * * * * cp /path/to/monitoring_data.json /path/to/backups/monitoring_data.$(date +\%Y\%m\%d_\%H).json
```

### Version Control

Add to `.gitignore` (already done):
```
monitoring_data.json
```

Never commit sensitive data like Telegram tokens to git!

### Secure Storage

The data file contains:
- ✅ Service URLs and names (usually safe)
- ⚠️ Telegram bot token (sensitive!)

Protect this file:
```bash
chmod 600 monitoring_data.json  # Only owner can read/write
```

### Data Location for Production

You can change the data file location by modifying `main.go`:
```go
// Change from:
dataFile := "monitoring_data.json"

// To:
dataFile := "/var/lib/monitoring/data.json"
```

## Troubleshooting

### File Permission Errors

If you get permission errors:
```bash
chmod 644 monitoring_data.json
```

### Corrupted File

If the file gets corrupted:
1. Stop the application
2. Restore from backup or delete the file
3. Restart the application

### Large File Size

The JSON file is human-readable but can grow large. To keep it manageable:
- Regularly remove old/unused services
- Consider archiving old data periodically

### File Not Found on Startup

This is normal! If the file doesn't exist:
- Application starts with empty data
- File is created on first save
- No action needed

## Technical Details

### Thread Safety

The persistence layer is thread-safe:
- Concurrent reads are allowed
- Writes are serialized with mutex locks
- No data races or corruption

### Save Performance

Saves are done asynchronously:
- Non-blocking - doesn't slow down API responses
- Runs in background goroutine
- Minimal performance impact

### Error Handling

If save fails:
- Error is logged to console
- Application continues running
- Data remains in memory
- Next save will retry

## Future Enhancements

Consider migrating to:
- **SQLite**: Better for 100+ services
- **PostgreSQL**: Enterprise-grade, multi-instance
- **Redis**: Fast, supports clustering
- **Cloud Storage**: S3, Google Cloud Storage for backups

But for most users, JSON file persistence is perfect!
