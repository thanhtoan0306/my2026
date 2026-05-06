function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

function clamp(n, min, max) {
  return Math.max(min, Math.min(max, n));
}

async function getPageMetrics() {
  const docEl = document.documentElement;
  const body = document.body;
  const scrollHeight = Math.max(
    docEl.scrollHeight,
    body?.scrollHeight ?? 0,
    docEl.offsetHeight,
    body?.offsetHeight ?? 0
  );

  const scrollWidth = Math.max(
    docEl.scrollWidth,
    body?.scrollWidth ?? 0,
    docEl.offsetWidth,
    body?.offsetWidth ?? 0
  );

  // viewport in CSS pixels
  const viewportW = window.innerWidth;
  const viewportH = window.innerHeight;
  const dpr = window.devicePixelRatio || 1;

  return { scrollHeight, scrollWidth, viewportW, viewportH, dpr };
}

async function captureFullPage() {
  const { scrollHeight, scrollWidth, viewportH, dpr } = await getPageMetrics();

  // Scroll in CSS px; captures come back in device px.
  const totalSteps = Math.ceil(scrollHeight / viewportH);
  const originalX = window.scrollX;
  const originalY = window.scrollY;

  const slices = [];

  for (let step = 0; step < totalSteps; step++) {
    const y = clamp(step * viewportH, 0, scrollHeight - viewportH);
    window.scrollTo(0, y);
    // Allow layout/paint to settle; pages with lazy-load need a moment.
    await sleep(150);

    const resp = await chrome.runtime.sendMessage({ type: "EXTCAP_CAPTURE_VIEWPORT" });
    if (!resp?.ok) throw new Error(resp?.error || "Viewport capture failed.");

    slices.push({
      dataUrl: resp.dataUrl,
      offsetYCss: y
    });
  }

  window.scrollTo(originalX, originalY);

  const payload = {
    // Stitch size in device pixels.
    widthPx: Math.round(scrollWidth * dpr),
    heightPx: Math.round(scrollHeight * dpr),
    dpr,
    slices: slices.map((s) => ({
      dataUrl: s.dataUrl,
      offsetYPx: Math.round(s.offsetYCss * dpr)
    }))
  };

  const stitched = await chrome.runtime.sendMessage({
    type: "EXTCAP_STITCH_AND_COPY",
    payload
  });

  if (!stitched?.ok) throw new Error(stitched?.error || "Stitch/copy failed.");
  return stitched;
}

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
  (async () => {
    try {
      if (msg?.type !== "EXTCAP_START") return;
      const result = await captureFullPage();
      sendResponse(result);
    } catch (e) {
      sendResponse({ ok: false, error: e?.message ?? String(e) });
    }
  })();
  return true;
});

