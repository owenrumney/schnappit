# Schnappit - Screenshot Tool Design Document

## Overview

Schnappit is a lightweight screenshot capture and annotation tool for macOS, written in Go. While Swift is the typical choice for macOS apps, this project prioritizes maintainability and developer familiarity over native conventions.

## Feasibility Assessment

### Can this be done in Go? Yes, with caveats.

| Capability | Library | Status | Notes |
|------------|---------|--------|-------|
| Global Hotkeys | `golang.design/x/hotkey` | Mature | Cross-platform, works with Fyne |
| Screenshot Capture | `kbinani/screenshot` | Working | Uses CGO on macOS |
| Native macOS APIs | DarwinKit | Partial | 33 frameworks, no ScreenCaptureKit |
| Cross-platform UI | Fyne | Mature | Canvas primitives for annotation |

### Key Findings

1. **DarwinKit** (formerly MacDriver) provides bindings for AppKit, CoreGraphics, and 31 other Apple frameworks. However:
   - No ScreenCaptureKit bindings (the modern screenshot API)
   - Maintainers don't recommend it for "large/complex programs"
   - Still requires Xcode for framework headers (CGO)

2. **Screenshot APIs**: Apple deprecated `CGWindowListCreateImage` in macOS 15 in favor of ScreenCaptureKit. The `kbinani/screenshot` library still works but uses the deprecated API.

3. **Global Hotkeys**: `golang.design/x/hotkey` works cross-platform. On macOS, requires either:
   - Explicit mainthread handling via `mainthread.Init()`
   - A GUI framework that handles this (Fyne, Gio, Ebiten)

4. **Cross-platform Bonus**: Using Fyne for UI makes Linux/Windows support achievable with the same codebase.

---

## Recommended Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Schnappit                            │
├─────────────────────────────────────────────────────────────┤
│  UI Layer (Fyne)                                            │
│  ├── Main Window (hidden/system tray)                       │
│  ├── Region Selector Overlay                                │
│  └── Annotation Editor                                      │
├─────────────────────────────────────────────────────────────┤
│  Core Services                                              │
│  ├── Hotkey Manager (golang.design/x/hotkey)                │
│  ├── Screenshot Capture (kbinani/screenshot)                │
│  ├── Clipboard Manager (golang.design/x/clipboard)          │
│  └── File Storage                                           │
├─────────────────────────────────────────────────────────────┤
│  Platform Layer                                             │
│  ├── macOS: CGO + CoreGraphics (via screenshot lib)         │
│  ├── Linux: X11/Wayland                                     │
│  └── Windows: GDI                                           │
└─────────────────────────────────────────────────────────────┘
```

### Why Fyne over DarwinKit?

| Aspect | DarwinKit | Fyne |
|--------|-----------|------|
| Cross-platform | macOS only | macOS, Linux, Windows, mobile |
| Complexity tolerance | "Not recommended for complex apps" | Production-ready |
| Canvas/Drawing | Via CoreGraphics (manual) | Built-in canvas primitives |
| Hotkey integration | Manual | Works with golang.design/x/hotkey |
| Learning curve | Must understand Objective-C APIs | Go-native API |

**Recommendation**: Use **Fyne** for the UI layer, with potential DarwinKit usage for macOS-specific features (system tray, native dialogs) if Fyne's implementations prove insufficient.

---

## MVP Feature Breakdown

### 1. Global Hotkey Trigger (Cmd+Shift+X)

**Library**: `golang.design/x/hotkey`

```go
// Example registration
hk := hotkey.New([]hotkey.Modifier{hotkey.ModCmd, hotkey.ModShift}, hotkey.KeyX)
hk.Register()

go func() {
    for range hk.Keydown() {
        // Trigger screenshot flow
    }
}()
```

**Requirements**:
- macOS: Accessibility permission required (System Settings > Privacy & Security > Accessibility)
- Fyne handles mainthread requirements automatically

### 2. Region Selection

**Approach**: Full-screen transparent overlay with click-drag selection

```
┌────────────────────────────────────────┐
│ ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │
│ ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │
│ ░░░░░░┌─────────────────┐░░░░░░░░░░░░░ │
│ ░░░░░░│  Selected Area  │░░░░░░░░░░░░░ │
│ ░░░░░░│   (Clear)       │░░░░░░░░░░░░░ │
│ ░░░░░░└─────────────────┘░░░░░░░░░░░░░ │
│ ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │
└────────────────────────────────────────┘
  ░ = Semi-transparent overlay (dims screen)
```

**Implementation**:
- Create borderless, fullscreen, transparent Fyne window
- Track mouse events for drag selection
- Capture selected region using `screenshot.CaptureRect(bounds)`

### 3. Annotation Editor

**Library**: Fyne canvas primitives + custom widgets

**Tools Required**:
| Tool | Fyne Primitive | Notes |
|------|----------------|-------|
| Arrow | `canvas.Line` + triangle | Custom widget needed |
| Rectangle | `canvas.Rectangle` | StrokeColor, StrokeWidth |
| Text | `canvas.Text` | With editable input overlay |
| Freehand | Series of `canvas.Line` | Capture mouse path |

**Editor Layout**:
```
┌──────────────────────────────────────────────────┐
│ [Arrow] [Rect] [Text] [Color] │ [Copy] [Save] [X]│
├──────────────────────────────────────────────────┤
│                                                  │
│                                                  │
│            Screenshot + Annotations              │
│                                                  │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Annotation Data Model**:
```go
type Annotation interface {
    Draw(canvas fyne.Canvas)
    Contains(pos fyne.Position) bool
    Move(delta fyne.Position)
}

type ArrowAnnotation struct {
    Start, End fyne.Position
    Color      color.Color
    Width      float32
}

type RectAnnotation struct {
    Bounds     fyne.Position
    Size       fyne.Size
    Color      color.Color
    Width      float32
    Filled     bool
}
```

### 4. Output Handling

**Default**: Copy to clipboard
**Library**: `golang.design/x/clipboard`

```go
import "golang.design/x/clipboard"

func copyToClipboard(img image.Image) error {
    var buf bytes.Buffer
    png.Encode(&buf, img)
    return clipboard.Write(clipboard.FmtImage, buf.Bytes())
}
```

**Save Location**: `~/Pictures/schnappit/`
- Filename format: `schnappit-2024-01-15-143052.png`
- Create directory if not exists

---

## Project Structure

```
schnappit/
├── cmd/
│   └── schnappit/
│       └── main.go           # Entry point
├── internal/
│   ├── app/
│   │   └── app.go            # Application lifecycle
│   ├── capture/
│   │   ├── capture.go        # Screenshot capture logic
│   │   └── region.go         # Region selection overlay
│   ├── editor/
│   │   ├── editor.go         # Annotation editor window
│   │   ├── canvas.go         # Custom canvas with annotations
│   │   └── tools/
│   │       ├── arrow.go
│   │       ├── rectangle.go
│   │       └── text.go
│   ├── hotkey/
│   │   └── hotkey.go         # Global hotkey management
│   └── output/
│       ├── clipboard.go      # Clipboard operations
│       └── file.go           # File save operations
├── go.mod
├── go.sum
├── Makefile
└── DESIGN.md
```

---

## Dependencies

```go
require (
    fyne.io/fyne/v2           // UI framework
    golang.design/x/hotkey    // Global hotkeys
    golang.design/x/clipboard // Clipboard access
)
// Screen capture: vendor ro31337/screenshot_macos CGO code into internal/capture/
```

### Screenshot Capture Options

| Library | API | macOS Version | Notes |
|---------|-----|---------------|-------|
| `kbinani/screenshot` | CGWindowListCreateImage | Any | Deprecated in macOS 15, but still works |
| `ro31337/screenshot_macos` | ScreenCaptureKit | 12.3+ | **Working CGO example with modern APIs** |
| `tfsoares/screencapturekit-go` | ScreenCaptureKit | 12.3+ | Video/streaming focused, no screenshot API |

**Recommendation**: Use `ro31337/screenshot_macos` as the basis for capture. It provides:

```go
// Already implemented with ScreenCaptureKit + CGO
Capture(x, y, width, height int) *image.RGBA
NumActiveDisplays() int
GetDisplayBounds(displayIndex int) image.Rectangle
```

This is a single-file CGO implementation that:
- Uses `SCShareableContent` (modern API, not deprecated)
- Supports multi-monitor setups
- Returns standard Go `image.RGBA`
- Can be vendored directly into `internal/capture/`

The code bridges Go ↔ Objective-C using dispatch semaphores for async callback handling.

---

## Platform Considerations

### macOS Permissions Required

1. **Accessibility** - For global hotkeys
   - System Settings > Privacy & Security > Accessibility

2. **Screen Recording** - For screenshot capture
   - System Settings > Privacy & Security > Screen Recording

### Application Bundle

For distribution, the Go binary needs to be wrapped in a `.app` bundle:

```
Schnappit.app/
├── Contents/
│   ├── Info.plist
│   ├── MacOS/
│   │   └── schnappit        # Go binary
│   └── Resources/
│       └── icon.icns
```

**Info.plist** must include:
- `NSScreenCaptureUsageDescription` - Why screen recording is needed
- `NSAppleEventsUsageDescription` - For accessibility

---

## Known Limitations & Risks

### 1. Screenshot API Choice (RESOLVED)
- **Previous concern**: `kbinani/screenshot` uses deprecated `CGWindowListCreateImage`
- **Solution**: `ro31337/screenshot_macos` provides working ScreenCaptureKit implementation
- **Action**: Vendor the CGO code from `ro31337/screenshot_macos` into `internal/capture/`

### 2. macOS Version Requirement
- **Impact**: ScreenCaptureKit requires macOS 12.3+
- **Mitigation**: Acceptable - macOS 12 (Monterey) released Oct 2021, most users are on 13+

### 3. Fyne Overlay Window Behavior
- **Risk**: Full-screen transparent overlay may not work perfectly on all macOS versions
- **Mitigation**: Test extensively; may need DarwinKit for NSWindow configuration

### 4. Multi-monitor Support
- **Risk**: Region selection across monitors is complex
- **Mitigation**: MVP targets single monitor; `kbinani/screenshot` supports multi-monitor for future

---

## Development Phases

### Phase 1: Core Capture
- [ ] Project setup with Fyne
- [ ] Global hotkey registration
- [ ] Basic fullscreen capture
- [ ] Copy to clipboard

### Phase 2: Region Selection
- [ ] Transparent overlay window
- [ ] Click-drag region selection
- [ ] Capture selected region only

### Phase 3: Annotation Editor
- [ ] Editor window with screenshot display
- [ ] Rectangle annotation tool
- [ ] Arrow annotation tool
- [ ] Color picker
- [ ] Save/Copy buttons

### Phase 4: Polish
- [ ] System tray integration
- [ ] Settings/preferences
- [ ] Application bundle for distribution
- [ ] Auto-save to ~/Pictures/schnappit

---

## Alternative Approaches Considered

### 1. Pure DarwinKit
- **Pros**: Native macOS look, smaller binary
- **Cons**: macOS only, "not recommended for complex apps", steeper learning curve

### 2. Wails (Go + Web UI)
- **Pros**: Modern web-based UI, good ecosystem
- **Cons**: Heavier runtime, less native feel

### 3. Gio
- **Pros**: Immediate mode GUI, performant
- **Cons**: Lower-level than Fyne, more work for standard UI elements

### 4. Shell out to screencapture
- **Pros**: Uses Apple's tool directly
- **Cons**: Less control, can't do region selection UI, not cross-platform

**Decision**: Fyne provides the best balance of capability, maintainability, and cross-platform potential.

---

## References

- [DarwinKit (GitHub)](https://github.com/progrium/darwinkit)
- [Fyne Toolkit](https://fyne.io/)
- [golang.design/x/hotkey](https://pkg.go.dev/golang.design/x/hotkey)
- [kbinani/screenshot](https://github.com/kbinani/screenshot) - Legacy CGWindowListCreateImage approach
- [ro31337/screenshot_macos](https://github.com/ro31337/screenshot_macos) - **ScreenCaptureKit CGO implementation (recommended)**
- [screencapturekit-go](https://github.com/tfsoares/screencapturekit-go) - ScreenCaptureKit bindings (video/streaming focused)
- [ScreenCaptureKit (Apple Docs)](https://developer.apple.com/documentation/screencapturekit/)
- [SCShareableContent (Apple Docs)](https://developer.apple.com/documentation/screencapturekit/scshareablecontent)
