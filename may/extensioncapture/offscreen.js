async function dataUrlToImageBitmap(dataUrl) {
  const resp = await fetch(dataUrl);
  const blob = await resp.blob();
  return await createImageBitmap(blob);
}

const BUILD = "0.1.3";
console.log(`ExtensionCapture offscreen loaded (v${BUILD})`);

async function canvasToPngBlob(canvas) {
  // OffscreenCanvas uses convertToBlob (HTMLCanvasElement uses toBlob).
  if (typeof canvas?.convertToBlob === "function") {
    return await canvas.convertToBlob({ type: "image/png" });
  }
  if (typeof canvas?.toBlob === "function") {
    return await new Promise((resolve, reject) => {
      canvas.toBlob((b) => (b ? resolve(b) : reject(new Error("canvas.toBlob returned null"))), "image/png");
    });
  }
  throw new Error("Canvas cannot be converted to Blob.");
}

async function writeBlobToClipboardPng(blob) {
  const item = new ClipboardItem({ "image/png": blob });
  await navigator.clipboard.write([item]);
}

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
  (async () => {
    try {
      if (msg?.type === "OFFSCREEN_PING") {
        sendResponse({ ok: true, build: BUILD });
        return;
      }
      if (msg?.type !== "OFFSCREEN_STITCH_AND_COPY") return;
      const { widthPx, heightPx, slices } = msg.payload || {};
      if (!widthPx || !heightPx || !Array.isArray(slices) || slices.length === 0) {
        throw new Error("Invalid stitch payload.");
      }

      const canvas = new OffscreenCanvas(widthPx, heightPx);
      const ctx = canvas.getContext("2d");
      if (!ctx) throw new Error("Could not create 2D canvas context.");

      for (const slice of slices) {
        const bmp = await dataUrlToImageBitmap(slice.dataUrl);
        const y = slice.offsetYPx ?? 0;
        ctx.drawImage(bmp, 0, y);
        bmp.close?.();
      }

      const pngBlob = await canvasToPngBlob(canvas);
      await writeBlobToClipboardPng(pngBlob);

      sendResponse({ ok: true });
    } catch (e) {
      console.error("Offscreen stitch/copy failed:", e);
      sendResponse({ ok: false, error: e?.message ?? String(e) });
    }
  })();
  return true;
});

