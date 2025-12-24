# Quick Start: Live Updates Testing

## 30-Second Setup

```bash
# Build binaries
go build -o bin/paso-daemon ./cmd/daemon
go build -o bin/paso .

# Install daemon (one-time setup)
sudo cp bin/paso-daemon /usr/local/bin/
./scripts/install_systemd.sh

# Verify daemon running
systemctl --user status paso
```

## 5-Minute Test

**Terminal 1:**
```bash
./bin/paso
# Creates project, adds a few tasks
```

**Terminal 2:**
```bash
./bin/paso
# Observe real-time sync from Terminal 1
```

**In Terminal 1:**
- Press 'n' to create new task
- Type "Test task"
- Press Enter, then Tab to confirm

**Expected in Terminal 2:**
- ✅ Task appears within ~100ms
- ✅ Shows "Synced with other instances" notification

## Full Test Suite

```bash
# See TESTING_GUIDE.md for:
# - Event batching verification
# - Project scoping tests
# - Reconnection scenarios
# - Concurrent operations
# - Error handling
```

## Verify Implementation

```bash
# All tests pass with race detector
go test -race ./...

# Check compilation
go build -v ./...

# View implementation summary
cat IMPLEMENTATION_SUMMARY.md
```

## Troubleshooting

**Daemon won't start:**
```bash
journalctl --user -u paso -n 20
systemctl --user restart paso
```

**No sync happening:**
```bash
# Check socket exists
ls -la ~/.paso/paso.sock

# Restart daemon
systemctl --user restart paso
```

**Want to reset:**
```bash
# Stop daemon
systemctl --user stop paso

# Remove socket
rm ~/.paso/paso.sock

# Start daemon
systemctl --user start paso
```

## Key Files to Review

1. **IMPLEMENTATION_SUMMARY.md** - Complete overview of changes
2. **TESTING_GUIDE.md** - Detailed test procedures
3. **internal/events/client.go** - Event client with batching & reconnection
4. **internal/daemon/server.go** - Pub-sub daemon with filtering
5. **main.go** - Application integration point

## Performance Checklist

After testing, verify:
- [ ] Latency: 50-150ms end-to-end
- [ ] Batching: 10 rapid ops → 1-2 refreshes
- [ ] Memory: Stable ~50MB per instance
- [ ] CPU: 0% idle, <5% during operations
- [ ] Reconnection: Works within 5s after daemon restart

## Next Steps

1. Complete manual testing (TESTING_GUIDE.md)
2. Review implementation (IMPLEMENTATION_SUMMARY.md)
3. Commit when satisfied: `git add -A && git commit -m "feat: implement live updates system"`
4. Create PR to main branch
5. Merge and deploy

---

**Questions or issues?**
- Check TESTING_GUIDE.md for detailed test procedures
- Review IMPLEMENTATION_SUMMARY.md for architecture details
- Use `journalctl --user -u paso -f` to watch daemon logs
- Check `~/.paso/` for socket file and database
