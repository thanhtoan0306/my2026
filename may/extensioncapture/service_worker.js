const BUILD = "0.1.3";
const OFFSCREEN_URL = chrome.runtime.getURL(`offscreen.html?v=${BUILD}`);

console.log(`ExtensionCapture service worker loaded (v${BUILD})`);

const BLOCKED_URL_PREFIXES = [
  "chrome://",
  "chrome-extension://",
  "edge://",
  "about:",
  "view-source:",
  "devtools://"
];

async function ensureOffscreen() {
  const has = await chrome.offscreen.hasDocument?.();
  if (has) {
    // Offscreen documents can survive reloads; verify it's the current build.
    try {
      const ping = await chrome.runtime.sendMessage({ type: "OFFSCREEN_PING" });
      if (ping?.ok && ping?.build === BUILD) return;
      await chrome.offscreen.closeDocument();
    } catch {
      try {
        await chrome.offscreen.closeDocument();
      } catch {
        // ignore
      }
    }
  }
  await chrome.offscreen.createDocument({
    url: OFFSCREEN_URL,
    reasons: ["DOM_SCRAPING"],
    justification: "Stitch captured viewports on a canvas and write image to clipboard."
  });
}

async function notify({ title, message, isError = false }) {
  try {
    await chrome.notifications.create({
      type: "basic",
      iconUrl: chrome.runtime.getURL(
        isError
          ? "data:image/svg+xml;charset=utf-8," +
              encodeURIComponent(
                `<svg xmlns="http://www.w3.org/2000/svg" width="128" height="128"><rect width="128" height="128" rx="24" fill="#B00020"/><path d="M40 40 L88 88 M88 40 L40 88" stroke="#fff" stroke-width="14" stroke-linecap="round"/></svg>`
              )
          : "data:image/svg+xml;charset=utf-8," +
              encodeURIComponent(
                `<svg xmlns="http://www.w3.org/2000/svg" width="128" height="128"><rect width="128" height="128" rx="24" fill="#0F7B0F"/><path d="M34 68 L54 88 L94 44" stroke="#fff" stroke-width="14" stroke-linecap="round" stroke-linejoin="round"/></svg>`
              )
      ),
      title,
      message
    });
  } catch (e) {
    // Notifications may be blocked; fall back to console.
    console.warn("Notification failed:", e);
  }
}

async function setActionError(message) {
  try {
    await chrome.action.setBadgeBackgroundColor({ color: "#B00020" });
    await chrome.action.setBadgeText({ text: "ERR" });
    await chrome.action.setTitle({ title: `ExtensionCapture: ${message}` });
  } catch {
    // ignore
  }
}

async function clearActionError(tabId) {
  try {
    const details = typeof tabId === "number" ? { tabId } : {};
    await chrome.action.setBadgeText({ ...details, text: "" });
    // If we don't set a title, Chrome falls back to manifest action.default_title.
    await chrome.action.setTitle({ ...details, title: "Capture full page to clipboard" });
  } catch {
    // ignore
  }
}

async function getActiveTab() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.id) throw new Error("No active tab found.");
  return tab;
}

function isBlockedUrl(url) {
  if (!url) return true;
  return BLOCKED_URL_PREFIXES.some((p) => url.startsWith(p));
}

chrome.action.onClicked.addListener(async () => {
  try {
    const tab = await getActiveTab();
    const url = tab.url || "";

    if (isBlockedUrl(url)) {
      const msg = "Cannot run on this page (try an http/https tab).";
      await setActionError(msg);
      // Don't throw: thrown errors show up in chrome://extensions "Errors".
      console.warn("ExtensionCapture blocked URL:", url);
      await notify({
        title: "ExtensionCapture",
        message: msg,
        isError: true
      });
      return;
    }

    await clearActionError(tab.id);

    // Inject (or re-inject) the content script to orchestrate scrolling.
    await chrome.scripting.executeScript({
      target: { tabId: tab.id },
      files: ["content_script.js"]
    });

    await ensureOffscreen();

    // Start capture flow.
    const result = await chrome.tabs.sendMessage(tab.id, { type: "EXTCAP_START" });
    if (result?.ok) {
      await notify({
        title: "Capture success",
        message: "Full-page screenshot copied to clipboard."
      });
    } else {
      const msg = result?.error || "Capture failed.";
      await setActionError(msg);
      await notify({ title: "Capture error", message: msg, isError: true });
    }
  } catch (e) {
    console.error("ExtensionCapture failed:", e);
    const msg = e?.message ?? String(e);
    await setActionError(msg);
    await notify({ title: "Capture error", message: msg, isError: true });
  }
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  (async () => {
    try {
      if (msg?.type === "EXTCAP_CAPTURE_VIEWPORT") {
        const windowId = sender.tab?.windowId;
        if (typeof windowId !== "number") throw new Error("Missing sender windowId.");
        const dataUrl = await chrome.tabs.captureVisibleTab(windowId, { format: "png" });
        sendResponse({ ok: true, dataUrl });
        return;
      }

      if (msg?.type === "EXTCAP_STITCH_AND_COPY") {
        await ensureOffscreen();
        const result = await chrome.runtime.sendMessage({
          type: "OFFSCREEN_STITCH_AND_COPY",
          payload: msg.payload
        });
        sendResponse(result);
        return;
      }

      sendResponse({ ok: false, error: "Unknown message type." });
    } catch (e) {
      sendResponse({ ok: false, error: e?.message ?? String(e) });
    }
  })();

  // Keep the message channel open for async responses.
  return true;
});

