# YT CC Black/Yellow

Chrome extension (MV3) for YouTube:

- Auto-enables captions (CC) when you open a video.
- Forces captions to **black background + yellow text** via injected CSS.
- If current caption contains Chinese characters, shows a **pink duplicate line** converted to **pinyin**.

## Install (Developer mode)

1. Open `chrome://extensions`
2. Enable **Developer mode**
3. Click **Load unpacked**
4. Select this folder: `may/ext-cc`

## Notes

- Works on `https://www.youtube.com/*`
- YouTube UI changes can break selectors; the extension retries enabling CC on SPA navigations.
- Pinyin conversion uses embedded `tiny-pinyin` browser build (global `Pinyin`).

