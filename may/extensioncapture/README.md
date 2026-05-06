# ExtensionCapture

Google Chrome (Manifest V3) extension: click the toolbar icon to capture a **full-page** screenshot and copy it to the **clipboard**.

## Load it

1. Open `chrome://extensions`
2. Enable **Developer mode**
3. Click **Load unpacked**
4. Select this folder: `may/extensioncapture`

## Use it

- Click the extension’s toolbar icon.
- Wait a moment (it scrolls the page to capture).
- Paste into Slack/Docs/Images/etc.

## Notes / limitations

- Some pages may block capture (Chrome restriction pages, some privileged URLs).
- Very tall pages can be memory-heavy (stitching creates one big PNG).

