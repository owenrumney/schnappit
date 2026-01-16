# Schnappit

A lightweight screenshot capture and annotation tool for macOS, written in Go.

## Features

- **Region Selection** - Click and drag to select any screen region
- **Annotation Tools** - Add arrows and rectangles to highlight areas
- **Quick Export** - Copy to clipboard or save to file
- **Global Hotkey** - Trigger capture from anywhere (default: `Cmd+Shift+X`)
- **Menu Bar App** - Runs quietly in your menu bar

## Installation

### Homebrew (Recommended)

```bash
brew tap owenrumney/schnappit
brew install --cask schnappit
```

### Download

Download the latest `.dmg` from the [Releases](https://github.com/owenrumney/schnappit/releases) page.

1. Open the `.dmg` file
2. Drag Schnappit to Applications
3. Launch from Applications

**Note:** On first launch, macOS may show a security warning because the app isn't notarized. To open it:
- Right-click (or Control-click) on Schnappit in Applications
- Select "Open" from the context menu
- Click "Open" in the dialog

Alternatively, run this in Terminal:
```bash
xattr -cr /Applications/Schnappit.app
```

### From Source

```bash
# Clone the repository
git clone https://github.com/owenrumney/schnappit.git
cd schnappit

# Build the application bundle
make bundle

# Copy to Applications
cp -r build/Schnappit.app /Applications/

# Or run directly
make run
```

### Requirements

- macOS 14.0 (Sonoma) or later
- Go 1.21 or later (for building from source)

## Usage

1. **Launch** - Start Schnappit from Applications or run `make run`
2. **Capture** - Press `Cmd+Shift+X` or click the menu bar icon and select "Capture Screenshot"
3. **Select Region** - Click and drag to select the area you want to capture
4. **Annotate** - Use the toolbar to add arrows or rectangles
5. **Export** - Click the copy icon to copy to clipboard, or save icon to save to file

### Keyboard Shortcuts

| Action | Shortcut |
|--------|----------|
| Capture Screenshot | `Cmd+Shift+X` (configurable) |
| Confirm Selection | `Enter` |
| Cancel Selection | `Escape` |

## Configuration

Schnappit stores its configuration at `~/.config/schnappit/config.json`.

```json
{
  "hotkey": "cmd+shift+x"
}
```

### Hotkey Format

Hotkeys are specified as modifier keys plus a key, separated by `+`:

**Modifiers:** `cmd`, `ctrl`, `shift`, `alt` (or `option`)

**Keys:** `a-z`, `0-9`, `f1-f12`, `space`, `enter`, `escape`, `tab`, `delete`

**Examples:**
- `cmd+shift+x` - Command + Shift + X
- `ctrl+alt+s` - Control + Option + S
- `cmd+shift+4` - Command + Shift + 4

You can also override the config file using the `SCHNAPPIT_HOTKEY` environment variable.

## Permissions

Schnappit requires the following macOS permissions:

1. **Screen Recording** - To capture screenshots
   - System Settings → Privacy & Security → Screen Recording → Enable Schnappit

2. **Accessibility** - For global hotkey support
   - System Settings → Privacy & Security → Accessibility → Enable Schnappit

You'll be prompted to grant these permissions on first use.

## Screenshots Saved To

Screenshots are saved to `~/Pictures/schnappit/` with timestamp-based filenames:
```
schnappit-2024-01-15-143052.png
```

## Development

```bash
# Run in development mode
make dev

# Run tests
make test

# Format code
make fmt

# Lint code
make lint
```

## License

MIT License - see LICENSE file for details.

## Credits

Built with:
- [Fyne](https://fyne.io/) - Cross-platform GUI toolkit
- [golang.design/x/hotkey](https://github.com/golang-design/hotkey) - Global hotkey support
- [golang.design/x/clipboard](https://github.com/golang-design/clipboard) - Clipboard access
