/**
 * Netflix subtitles style + auto-enable (best effort).
 *
 * - Force subtitles: black background + yellow text (CSS)
 * - If subtitle contains Han chars, show a pink duplicate line converted to pinyin (tone marks)
 * - Netflix is SPA; observe DOM + URL changes
 */

const BUILD = "0.1.0";

const PERF = {
  rafSync: 0,
  lastSubtitleText: "",
  lastPinyinOut: "",
};

function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

function isChineseText(s) {
  return /[\p{Script=Han}]/u.test(s || "");
}

function injectStyleOnce() {
  if (document.getElementById("ext-cc-netflix-style")) return;
  const style = document.createElement("style");
  style.id = "ext-cc-netflix-style";
  style.textContent = `
/* Netflix subtitle text (selectors vary; target multiple) */
.player-timedtext,
.player-timedtext-text-container,
.player-timedtext-text-container span,
.player-timedtext-text-container * {
  color: #ffeb3b !important;
}

/* Add readable background per line */
.player-timedtext-text-container span,
.player-timedtext-text-container {
  background: rgba(0, 0, 0, 0.86) !important;
  padding: 2px 6px !important;
  border-radius: 4px !important;
  box-decoration-break: clone !important;
  -webkit-box-decoration-break: clone !important;
  text-shadow: 0 0 2px rgba(0,0,0,0.9), 0 1px 2px rgba(0,0,0,0.9) !important;
  -webkit-text-stroke: 0.6px rgba(0,0,0,0.95) !important;
}

/* Duplicate overlay (pink pinyin) */
#ext-cc-nf-dup-root {
  position: fixed;
  left: 50%;
  bottom: 18%;
  transform: translateX(-50%);
  z-index: 2147483647;
  max-width: min(1100px, 92vw);
  width: fit-content;
  pointer-events: none;
  user-select: none;
  display: none;
}
#ext-cc-nf-dup-line {
  display: inline-block;
  padding: 6px 10px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.86);
  color: #ff4da6;
  font-family: Netflix Sans, Roboto, Arial, sans-serif;
  font-size: 30px;
  font-weight: 600;
  line-height: 1.25;
  text-align: center;
  white-space: pre-wrap;
  text-shadow: 0 0 2px rgba(0,0,0,0.9), 0 1px 2px rgba(0,0,0,0.9);
  -webkit-text-stroke: 0.6px rgba(0,0,0,0.95);
  box-decoration-break: clone;
  -webkit-box-decoration-break: clone;
}
`;
  document.documentElement.appendChild(style);
}

function ensureDupUi() {
  let root = document.getElementById("ext-cc-nf-dup-root");
  if (root) return root;
  root = document.createElement("div");
  root.id = "ext-cc-nf-dup-root";
  const line = document.createElement("div");
  line.id = "ext-cc-nf-dup-line";
  root.appendChild(line);
  document.documentElement.appendChild(root);
  return root;
}

function setDupUi(text, visible) {
  const root = ensureDupUi();
  const line = root.querySelector("#ext-cc-nf-dup-line");
  if (line) line.textContent = text || "";
  root.style.display = visible ? "block" : "none";
}

function isPlayerPage() {
  // Netflix URLs vary; rely on DOM presence too.
  return location.pathname.startsWith("/watch") || !!document.querySelector("video");
}

function readCurrentSubtitleText() {
  // Common Netflix timedtext container.
  const el =
    document.querySelector(".player-timedtext-text-container") ||
    document.querySelector(".player-timedtext") ||
    document.querySelector('[data-uia="player-subtitle"]');
  const t = el?.textContent || "";
  return t.replace(/\s+/g, " ").trim();
}

function syncDuplicateFromDomNow() {
  if (!isPlayerPage()) {
    setDupUi("", false);
    return;
  }
  const text = readCurrentSubtitleText();
  if (text === PERF.lastSubtitleText) return;
  PERF.lastSubtitleText = text;

  if (!text) {
    setDupUi("", false);
    return;
  }
  if (!isChineseText(text)) {
    setDupUi("", false);
    return;
  }

  let out = PERF.lastPinyinOut || text;
  try {
    if (typeof pinyinPro !== "undefined" && typeof pinyinPro.pinyin === "function") {
      const arr = pinyinPro.pinyin(text, { type: "array" });
      if (Array.isArray(arr) && arr.length) out = arr.join(" ");
      else out = pinyinPro.pinyin(text);
    }
  } catch {
    // ignore, fallback to original text
  }
  PERF.lastPinyinOut = out;
  setDupUi(out, true);
}

function scheduleSyncDuplicate() {
  if (PERF.rafSync) return;
  PERF.rafSync = requestAnimationFrame(() => {
    PERF.rafSync = 0;
    syncDuplicateFromDomNow();
  });
}

function findSubtitlesButton() {
  // Netflix player "Audio & Subtitles" button selectors can change.
  return (
    document.querySelector('[data-uia="audio-subtitle-button"]') ||
    document.querySelector('[data-uia="player-audio-subtitle-button"]') ||
    document.querySelector('[aria-label*="Subtitles"]') ||
    document.querySelector('button[title*="Subtitles"]')
  );
}

function subtitlesMenuOpen() {
  return (
    !!document.querySelector('[data-uia="audio-subtitle-panel"]') ||
    !!document.querySelector('[data-uia="player-audio-subtitle-panel"]') ||
    !!document.querySelector('[role="dialog"][aria-label*="Audio"]')
  );
}

function findFirstSubtitleOption() {
  // Look for the subtitles list; choose first non-off option.
  const candidates = Array.from(
    document.querySelectorAll(
      '[data-uia^="subtitle-item-"], [data-uia*="subtitle-track-"], button[role="menuitemradio"]'
    )
  );
  // Prefer items that are not "Off"
  const good = candidates.find((el) => !/off/i.test(el.textContent || ""));
  return /** @type {HTMLElement|null} */ (good || candidates[0] || null);
}

async function tryEnableSubtitlesOnce() {
  if (!isPlayerPage()) return;

  // Try to open menu and click a subtitle option if none selected.
  for (let i = 0; i < 20; i++) {
    const btn = findSubtitlesButton();
    if (btn) {
      // If menu isn't open, open it.
      if (!subtitlesMenuOpen()) {
        btn.click();
        await sleep(250);
      }
      const opt = findFirstSubtitleOption();
      if (opt) {
        const pressed = opt.getAttribute("aria-checked") === "true" || opt.getAttribute("aria-selected") === "true";
        if (!pressed) opt.click();
        return;
      }
    }
    await sleep(250);
  }
}

function watchUrlChanges(onChange) {
  let lastHref = location.href;
  const obs = new MutationObserver(() => {
    if (location.href !== lastHref) {
      lastHref = location.href;
      onChange();
    }
  });
  obs.observe(document.documentElement, { childList: true, subtree: true });
  return obs;
}

function watchForPlayer() {
  let lastAttemptAt = 0;
  const obs = new MutationObserver(() => {
    const now = Date.now();
    if (now - lastAttemptAt < 750) return;
    lastAttemptAt = now;
    injectStyleOnce();
    ensureDupUi();
    void tryEnableSubtitlesOnce();
    scheduleSyncDuplicate();
  });
  obs.observe(document.documentElement, { childList: true, subtree: true });
  return obs;
}

(async function main() {
  console.log(`[ext-cc-netflix] loaded v${BUILD}`);
  injectStyleOnce();
  ensureDupUi();
  scheduleSyncDuplicate();
  void tryEnableSubtitlesOnce();

  watchForPlayer();
  watchUrlChanges(() => {
    injectStyleOnce();
    ensureDupUi();
    PERF.lastSubtitleText = "";
    scheduleSyncDuplicate();
    void tryEnableSubtitlesOnce();
  });
})();

