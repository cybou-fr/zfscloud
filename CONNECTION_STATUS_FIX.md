# Connection Status Fix

## Issue
GUI was showing "Connected to server" status even when the server was not running.

## Root Cause
In `client/gui/app.go`, the `updateStats()` function was checking if the client object existed rather than checking if there was an active connection:

```go
// BUGGY CODE
a.stats.Connected = (a.client != nil)  // Always true once client is created
```

## Solution

### 1. Added `IsConnected()` method to Client
**File:** `client/client.go`

```go
// IsConnected returns true if the client has an active connection to the server
func (c *Client) IsConnected() bool {
	return c.conn != nil
}
```

This method properly checks if the underlying gRPC connection (`conn`) exists, which is only non-nil when actually connected to the server.

### 2. Updated GUI stats update
**File:** `client/gui/app.go`

```go
// FIXED CODE
a.stats.Connected = a.client.IsConnected()  // Properly checks connection state
```

## Behavior

### Before Fix
- ❌ Shows "Connected" immediately after launching GUI
- ❌ Remains "Connected" even when server is stopped
- ❌ No way to distinguish between client object existing vs actual connection

### After Fix
- ✅ Shows "Disconnected" when server is not running
- ✅ Shows "Connected" only when actually connected to server
- ✅ Accurately reflects connection state via `conn != nil` check

## Testing Steps

1. **Launch GUI without server:**
   ```powershell
   .\bin\zfscloud-gui.exe
   ```
   - Expected: Status shows "Disconnected from server" (red indicator)

2. **Start server:**
   ```powershell
   .\bin\zfscloud-server.exe
   ```

3. **Connect from GUI:**
   - Click "Connect" or configure server settings
   - Expected: Status changes to "Connected to server" (green indicator)

4. **Stop server:**
   - Stop the server process
   - Expected: Next heartbeat will fail, status shows "Disconnected"

## Files Modified
- `client/client.go` - Added `IsConnected()` method
- `client/gui/app.go` - Updated `updateStats()` to use `IsConnected()`

## Build Status
✅ Successfully compiled: `bin/zfscloud-gui.exe`

## Impact
- **User Experience:** Connection status now accurately reflects server availability
- **Code Quality:** Proper encapsulation of connection state checking
- **Maintainability:** Centralized connection check logic in Client
