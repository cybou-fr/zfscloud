# Snapshot Automation Quick Reference

## Automatic Snapshot Schedule

| Type | Frequency | Default Retention | Example Name |
|------|-----------|-------------------|--------------|
| Hourly | Every hour | 24 (1 day) | `tank/users/alice@hourly-20250610-140000` |
| Daily | Midnight | 7 (1 week) | `tank/users/alice@daily-20250610-000000` |
| Weekly | Sunday midnight | 4 (1 month) | `tank/users/alice@weekly-20250607-000000` |
| Monthly | 1st at midnight | 12 (1 year) | `tank/users/alice@monthly-20250601-000000` |

## Configuration

Edit `server.yaml`:

```yaml
snapshots:
  schedule:
    hourly: 24   # Number of hourly snapshots to keep
    daily: 7     # Number of daily snapshots to keep
    weekly: 4    # Number of weekly snapshots to keep
    monthly: 12  # Number of monthly snapshots to keep
```

Set to 0 to disable that snapshot type.

## Manual Snapshot Creation

### Using Admin RPC (from admin client):
```bash
# Create manual snapshot for user
grpcurl -d '{"admin_token":"your-token","user_id":"alice","snapshot_type":"manual"}' \
  localhost:50051 zfscloud.v1.AdminService/CreateSnapshot
```

### From Go code:
```go
req := &proto.CreateSnapshotRequest{
    AdminToken:   os.Getenv("ADMIN_TOKEN"),
    UserId:       "alice",
    SnapshotType: "manual",
}
resp, err := adminClient.CreateSnapshot(ctx, req)
```

## Snapshot Lifecycle

### Creation
1. Scheduler triggers at configured interval
2. For each active user:
   - Generate snapshot name
   - Create ZFS snapshot
   - Record in database
   - Log action

### Retention Enforcement
1. After creating snapshots of a type
2. For each user:
   - List all snapshots of that type
   - Sort by creation time
   - Keep N most recent (N = retention count)
   - Delete older snapshots
   - Update database

## Monitoring

### Check Scheduler Status
Look for log messages in server output:
```
Starting snapshot scheduler...
Snapshot scheduler started successfully
Creating hourly snapshots for all users...
Created hourly snapshot for user alice: tank/users/alice@hourly-20250610-140000 (1024000 bytes)
Completed hourly snapshots: 5 succeeded, 0 failed
Enforcing retention policy for hourly snapshots...
Deleted 2 old hourly snapshots
```

### Database Queries

**List all snapshots:**
```sql
SELECT * FROM snapshots ORDER BY created_at DESC;
```

**Count snapshots by user:**
```sql
SELECT user_id, COUNT(*) as snapshot_count 
FROM snapshots 
GROUP BY user_id;
```

**Snapshot audit trail:**
```sql
SELECT * FROM action_log 
WHERE action IN ('snapshot_created', 'snapshot_deleted', 'snapshot_created_manual')
ORDER BY timestamp DESC;
```

## ZFS Commands

### List snapshots:
```bash
zfs list -t snapshot
```

### List snapshots for specific user:
```bash
zfs list -t snapshot | grep "users/alice@"
```

### Check snapshot space usage:
```bash
zfs list -t snapshot -o name,used,refer
```

### Manual snapshot creation (for testing):
```bash
zfs snapshot tank/users/alice@manual-$(date +%Y%m%d-%H%M%S)
```

### Manual snapshot deletion:
```bash
zfs destroy tank/users/alice@hourly-20250610-140000
```

## Troubleshooting

### Snapshots not being created

1. **Check scheduler is running:**
   ```bash
   # Look for "Snapshot scheduler started successfully" in server logs
   tail -f server.log | grep snapshot
   ```

2. **Check retention policy:**
   ```bash
   # Make sure retention counts are > 0 in server.yaml
   cat server.yaml | grep -A 5 "snapshots:"
   ```

3. **Check ZFS permissions:**
   ```bash
   # Server needs permission to create snapshots
   zfs allow -u serveruser snapshot,mount,destroy tank/users
   ```

### Too many snapshots

1. **Reduce retention counts** in `server.yaml`
2. **Restart server** to apply new policy
3. **Wait for next retention enforcement** (happens after each snapshot creation)

### Snapshots taking too much space

1. **Check snapshot space usage:**
   ```bash
   zfs list -t snapshot -o name,used,refer | head -20
   ```

2. **Reduce retention policy** for snapshot types consuming most space

3. **ZFS compression** helps reduce snapshot overhead:
   ```bash
   zfs get compression tank/users
   zfs set compression=lz4 tank/users  # If not already enabled
   ```

## Performance Impact

### Snapshot Creation:
- **Time:** < 1 second per user (ZFS copy-on-write)
- **CPU:** Minimal, sequential per user
- **I/O:** Metadata writes only
- **Memory:** < 1MB per goroutine

### Retention Cleanup:
- **Time:** ~1-2 seconds per user (database + ZFS ops)
- **CPU:** Light, mostly I/O wait
- **Disk:** Blocks released asynchronously by ZFS

### Overhead:
- 4 goroutines (one per schedule type)
- Negligible when idle (waiting on timers)
- No impact on backup/restore operations

## Best Practices

1. **Start conservatively:**
   - Begin with default retention (24/7/4/12)
   - Monitor disk usage for first week
   - Adjust based on growth rate

2. **Monitor disk usage:**
   ```bash
   # Weekly check
   zfs list -t snapshot -o space
   ```

3. **Test restore procedures:**
   - Periodically verify snapshots are restorable
   - Document restore process

4. **Audit logs:**
   - Review action_log regularly
   - Set up alerts for snapshot failures

5. **Backup retention policy:**
   - Match business requirements
   - Consider compliance needs
   - Balance storage costs vs. retention needs

## Integration with Backup Client

The automatic snapshots work independently of client backups:

- **Client backups** → Files uploaded to user dataset
- **Automatic snapshots** → Point-in-time copies of user dataset
- **Combined benefit:** Can restore from any snapshot, including past backups

### Example Workflow:
1. Client uploads files daily (Phase 6 automatic backup)
2. Server creates daily snapshot at midnight
3. Can restore to any daily snapshot (7 days history)
4. Can restore to any weekly snapshot (4 weeks history)
5. Can restore to any monthly snapshot (1 year history)
