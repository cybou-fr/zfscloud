# Changelog

## [Unreleased] - 2025-11-03

### Fixed
- **GUI Launch Issue** - Client was defaulting to interactive CLI mode instead of GUI
  - Changed default mode from `--interactive` to `--gui`
  - Added explicit `--gui` and `--interactive` flags
  - GUI now launches by default when running `zfscloud-gui.exe`
  - CLI mode requires explicit `--interactive` flag

### Added
- **GUI Mode Selection** - Added `runGUIMode()` function to properly initialize GUI
- **Usage Documentation** - Created `CLIENT_USAGE.md` with comprehensive usage guide
  - GUI mode instructions
  - CLI mode instructions
  - Configuration examples
  - Troubleshooting section
  - Common tasks guide

### Technical Details

**Before:**
```powershell
./bin/zfscloud-gui.exe
# Output: Interactive CLI prompt (❌ wrong)
```

**After:**
```powershell
./bin/zfscloud-gui.exe
# Output: GUI window launches (✅ correct)

./bin/zfscloud-gui.exe --interactive
# Output: Interactive CLI prompt (✅ correct for explicit CLI)
```

### Files Modified
- `client/cmd/main.go`
  - Changed `interactive` flag default from `true` to `false`
  - Added `gui` flag with default `true`
  - Added `runGUIMode()` function
  - Added import alias for GUI package (`clientgui`)
  - Updated mode selection logic in `main()`

### Files Added
- `CLIENT_USAGE.md` - Complete usage guide
- `CHANGELOG.md` - This file

### Build
- ✅ Compilation successful
- ✅ Binary size: 17.3 MB (no change)
- ✅ GUI launches correctly by default
- ✅ CLI accessible with `--interactive` flag

---

## Previous Releases

See individual phase documentation:
- [PHASE4_COMPLETE.md](./PHASE4_COMPLETE.md) - Confirmation dialogs & progress bars
- [GUI_FIXES_COMPLETE.md](./GUI_FIXES_COMPLETE.md) - Critical bug fixes
- [GUI_DEEP_ANALYSIS.md](./GUI_DEEP_ANALYSIS.md) - Comprehensive analysis
- [PHASE3_COMPLETE.md](./PHASE3_COMPLETE.md) - Tooltip system
- [PHASE2_COMPLETE.md](./PHASE2_COMPLETE.md) - Theme & widgets
- [IMPLEMENTATION_COMPLETE.md](./IMPLEMENTATION_COMPLETE.md) - Core features
