# Sharing SSRDesktop with a Friend

## Quick steps (you)

```bash
cd may/ssrdesktop
./build.sh    # if you have not built recently
./share.sh
```

Send your friend:

**`build/dist/SSRDesktop-macOS.zip`**

They only need the zip — not the Go source repo.

## What your friend does

1. Unzip the file.
2. Move `SSRDesktop.app` to **Applications**.
3. **Right-click → Open** the first time (unsigned app).
4. Click **Open** in the dialog.

See `build/dist/INSTALL.txt` (copied into the share flow) for full troubleshooting.

## macOS security (important)

This app is **not** signed with an Apple Developer certificate and **not** notarized. That is normal for hobby projects.

| Symptom | Fix |
|---------|-----|
| “cannot be opened because developer cannot be verified” | Right-click → **Open** |
| “App is damaged and can’t be opened” | Quarantine from download: `xattr -cr /Applications/SSRDesktop.app` |
| Gatekeeper after cloud download | Same as above, or System Settings → Privacy & Security → **Open Anyway** |

**Notarization** (Apple’s scan + stapling) costs a paid Developer account (~$99/year) and extra CI steps. Friends can still run the app with right-click Open.

## Which Macs work?

After `./build.sh`, the binary is **universal** (arm64 + Intel):

- Apple Silicon (M1/M2/M3/M4) ✓  
- Intel Mac ✓  
- macOS **11.0+** required (see `Info.plist`)

Ask your friend: **Apple menu → About This Mac** → macOS version.

## Best ways to send the file

| Method | Notes |
|--------|--------|
| **AirDrop** | Best Mac-to-Mac; preserves .app well |
| **iCloud / Drive / Dropbox** | Share a link; friend downloads and unzips |
| **USB** | Copy `SSRDesktop-macOS.zip` |
| **Email** | Often blocked or size-limited (~25 MB); zip is ~8–10 MB — may work |
| **WeChat / Telegram** | Sometimes re-wraps files; friend must **fully unzip** before running |

Always send the **`.zip`**, not a raw `.app` inside a chat app (chats can break bundles).

## What NOT to send

- The whole `my2026` git repo (unnecessary)
- Only `main.go` (friend cannot run without building)
- `build/SSRDesktop` bare binary without `.app` wrapper

## If you change the app later

1. Bump version in `build.sh` (`CFBundleShortVersionString`) if you want.
2. `./build.sh && ./share.sh`
3. Send the new zip; friend replaces old `.app` in Applications.

## Optional: proper distribution later

For public release without scary dialogs:

1. Enroll in [Apple Developer Program](https://developer.apple.com/programs/)
2. Sign: `codesign --options runtime --sign "Developer ID Application: …" SSRDesktop.app`
3. Notarize: `xcrun notarytool submit …` then `xcrun stapler staple SSRDesktop.app`
4. Ship stapled `.app` or `.dmg`

Out of scope for a hello-world shared with one friend.
