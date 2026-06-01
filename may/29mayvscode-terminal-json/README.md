# Terminal JSON Beautify (VS Code extension)

Beautify minified or messy JSON from the integrated terminal, clipboard, or editor. Object keys are sorted by default (same idea as `may/29mayjsoncurl-go`).

## Install locally

```bash
cd may/29mayvscode-terminal-json
npm install
npm run compile
```

Then in VS Code / Cursor:

1. **F5** — open `may/29mayvscode-terminal-json` and press F5 (uses `.vscode/launch.json`) to run an Extension Development Host, or  
2. **CLI** — from this folder: `npx @vscode/vsce package` then install the `.vsix` via **Extensions: Install from VSIX**.

## Commands

| Command | When to use |
|--------|-------------|
| **Beautify JSON from Terminal Selection** | Select text in terminal → right-click or `Cmd+Shift+J` |
| **Beautify JSON from Selection** | Editor selection (falls back to clipboard) |
| **Beautify JSON in Editor** | Replace selection or whole file in place (`Cmd+Shift+J` in editor) |
| **Beautify JSON from Clipboard** | Paste buffer only |

Terminal flow copies your selection, finds embedded JSON (e.g. log prefix + `{...}`), prettifies it, opens a new JSON tab, and copies result to the clipboard.

## Settings

- `terminalJsonBeautify.sortKeys` — sort object keys (default: `true`)
- `terminalJsonBeautify.indent` — spaces per level (default: `2`)
- `terminalJsonBeautify.openInEditor` — open result in new tab (default: `true`)

## Dev

```bash
npm run watch   # recompile on save while debugging with F5
```
