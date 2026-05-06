# Netflix Subtitles Black/Yellow + Pinyin

Chrome extension (MV3) for Netflix:

- Best-effort auto-enable subtitles.
- Forces subtitles to **black background + yellow text** via injected CSS.
- If current subtitle contains Chinese characters, shows a **pink duplicate line** converted to **pinyin with tone marks** (via `pinyin-pro`).

## Install (Developer mode)

1. Open `chrome://extensions`
2. Enable **Developer mode**
3. Click **Load unpacked**
4. Select this folder: `may/ext-cc-netflix`

## Notes

- Netflix DOM/selectors change often; auto-enable is best-effort.
- If you want the duplicate line closer/farther, adjust `bottom` in `#ext-cc-nf-dup-root`.

