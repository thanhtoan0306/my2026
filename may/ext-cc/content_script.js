/**
 * YouTube CC style + auto-enable.
 *
 * Goal: show captions with black background + yellow text.
 * Note: YouTube is an SPA; we watch DOM + URL changes.
 */

const BUILD = "0.1.0";

const PERF = {
  rafSync: 0,
  lastCaptionText: "",
  lastPinyinOut: "",
  worker: /** @type {Worker|null} */ (null),
  workerReady: false,
};

function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

function ensurePinyinWorker() {
  if (PERF.workerReady) return;
  PERF.workerReady = true;
  try {
    const url = chrome?.runtime?.getURL?.("pinyin_worker.js");
    if (!url) throw new Error("chrome.runtime.getURL unavailable");
    const w = new Worker(url);
    w.onmessage = (ev) => {
      const msg = ev?.data;
      if (!msg || msg.type !== "PINYIN_RESULT") return;
      // Only apply if it matches the latest caption (avoid stale updates).
      if (msg.text !== PERF.lastCaptionText) return;
      if (typeof msg.out === "string" && msg.out) {
        PERF.lastPinyinOut = msg.out;
        setDupUi(msg.out, true);
      }
    };
    PERF.worker = w;
  } catch (e) {
    PERF.worker = null;
  }
}

function isChineseText(s) {
  // Detect Han characters (Chinese/Japanese Kanji). Good enough for "tiếng Trung" captions.
  return /[\p{Script=Han}]/u.test(s || "");
}

function injectStyleOnce() {
  if (document.getElementById("ext-cc-style")) return;
  const style = document.createElement("style");
  style.id = "ext-cc-style";
  style.textContent = `
/* Force caption text color */
.ytp-caption-segment,
.captions-text span,
.caption-window span {
  color: #ffeb3b !important;
}

/* Give each segment a readable "card" background */
.ytp-caption-segment {
  background: rgba(0, 0, 0, 0.86) !important;
  padding: 2px 6px !important;
  border-radius: 4px !important;
  box-decoration-break: clone !important;
  -webkit-box-decoration-break: clone !important;
  text-shadow: 0 0 2px rgba(0,0,0,0.9), 0 1px 2px rgba(0,0,0,0.9) !important;
  -webkit-text-stroke: 0.6px rgba(0,0,0,0.95) !important;
}

/* Reduce any white outline YouTube might apply */
.ytp-caption-window-container * {
  filter: none !important;
}

/* Duplicate CC overlay (pink) */
#ext-cc-dup-root {
  position: fixed;
  left: 50%;
  bottom: 18%;
  transform: translateX(-50%);
  z-index: 999999;
  max-width: min(1100px, 92vw);
  width: fit-content;
  pointer-events: none;
  user-select: none;
  display: none;
}
#ext-cc-dup-line {
  display: inline-block;
  padding: 6px 10px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.86);
  color: #ff4da6;
  font-family: Roboto, Arial, sans-serif;
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
  let root = document.getElementById("ext-cc-dup-root");
  if (root) return root;
  root = document.createElement("div");
  root.id = "ext-cc-dup-root";

  const line = document.createElement("div");
  line.id = "ext-cc-dup-line";
  root.appendChild(line);

  document.documentElement.appendChild(root);
  return root;
}

function setDupUi(text, visible) {
  const root = ensureDupUi();
  const line = root.querySelector("#ext-cc-dup-line");
  if (line) line.textContent = text || "";
  root.style.display = visible ? "block" : "none";
}

function isWatchLikePage() {
  return location.pathname === "/watch" || location.pathname.startsWith("/shorts/");
}

function findCaptionsButton() {
  // Standard player captions button
  return /** @type {HTMLElement|null} */ (document.querySelector(".ytp-subtitles-button"));
}

function captionsEnabled(btn) {
  const pressed = btn?.getAttribute("aria-pressed");
  if (pressed === "true") return true;
  if (pressed === "false") return false;
  // Fallback heuristic: YouTube sometimes uses classes
  return btn?.classList?.contains("ytp-button-active") || false;
}

async function tryEnableCaptionsOnce() {
  if (!isWatchLikePage()) return;

  // Wait a bit for player controls to show up on navigation.
  for (let i = 0; i < 30; i++) {
    const btn = findCaptionsButton();
    if (btn) {
      if (!captionsEnabled(btn)) {
        btn.click();
        // Sometimes first click is swallowed until controls are ready.
        await sleep(200);
      }
      return;
    }
    await sleep(200);
  }
}

function readCurrentCaptionText() {
  // Try multiple known containers.
  const container =
    document.querySelector(".ytp-caption-window-container") ||
    document.querySelector(".captions-text") ||
    document.querySelector(".caption-window");
  const t = container?.textContent || "";
  return t.replace(/\s+/g, " ").trim();
}

function syncDuplicateCcFromDomNow() {
  if (!isWatchLikePage()) {
    setDupUi("", false);
    return;
  }
  const text = readCurrentCaptionText();
  if (text === PERF.lastCaptionText) {
    // No change: avoid extra work and UI churn.
    return;
  }
  PERF.lastCaptionText = text;

  if (!text) {
    setDupUi("", false);
    return;
  }
  const show = isChineseText(text);
  if (!show) {
    setDupUi("", false);
    return;
  }

  // Show last result immediately (reduces perceived delay), then update async from worker.
  if (PERF.lastPinyinOut) setDupUi(PERF.lastPinyinOut, true);
  else setDupUi(text, true);

  ensurePinyinWorker();
  if (PERF.worker) {
    PERF.worker.postMessage({ type: "PINYIN", text });
    return;
  }

  // Fallback to sync conversion if Worker isn't available.
  let out = text;
  try {
    if (typeof pinyinPro !== "undefined" && typeof pinyinPro.pinyin === "function") {
      const arr = pinyinPro.pinyin(text, { type: "array" });
      out = Array.isArray(arr) && arr.length ? arr.join(" ") : pinyinPro.pinyin(text);
    }
  } catch {}
  PERF.lastPinyinOut = out || text;
  setDupUi(PERF.lastPinyinOut, true);
}

function scheduleSyncDuplicateCc() {
  if (PERF.rafSync) return;
  PERF.rafSync = requestAnimationFrame(() => {
    PERF.rafSync = 0;
    syncDuplicateCcFromDomNow();
  });
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

function watchForPlayerAndAutoEnable() {
  let lastAttemptAt = 0;
  const obs = new MutationObserver(() => {
    const now = Date.now();
    if (now - lastAttemptAt < 750) return;
    lastAttemptAt = now;
    void tryEnableCaptionsOnce();
    scheduleSyncDuplicateCc();
  });
  obs.observe(document.documentElement, { childList: true, subtree: true });
  return obs;
}

(async function main() {
  console.log(`[ext-cc] loaded v${BUILD}`);
  injectStyleOnce();
  ensureDupUi();
  ensurePinyinWorker();
  scheduleSyncDuplicateCc();

  void tryEnableCaptionsOnce();

  watchForPlayerAndAutoEnable();
  watchUrlChanges(() => {
    injectStyleOnce();
    ensureDupUi();
    void tryEnableCaptionsOnce();
    PERF.lastCaptionText = "";
    scheduleSyncDuplicateCc();
  });
})();

