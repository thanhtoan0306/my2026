/* eslint-disable no-restricted-globals */
/* global importScripts, pinyinPro */

try {
  importScripts("pinyin-pro.browser.js");
} catch (e) {}

/** @type {Map<string,string>} */
const CACHE = new Map();

function lruSet(key, val, max = 500) {
  if (CACHE.has(key)) CACHE.delete(key);
  CACHE.set(key, val);
  if (CACHE.size > max) {
    const firstKey = CACHE.keys().next().value;
    if (firstKey) CACHE.delete(firstKey);
  }
}

self.onmessage = (ev) => {
  const msg = ev?.data;
  if (!msg || msg.type !== "PINYIN") return;
  const text = String(msg.text || "");
  if (!text) {
    self.postMessage({ type: "PINYIN_RESULT", text, out: "" });
    return;
  }

  const cached = CACHE.get(text);
  if (cached != null) {
    self.postMessage({ type: "PINYIN_RESULT", text, out: cached });
    return;
  }

  let out = "";
  try {
    if (typeof pinyinPro !== "undefined" && typeof pinyinPro.pinyin === "function") {
      const arr = pinyinPro.pinyin(text, { type: "array" });
      out = Array.isArray(arr) && arr.length ? arr.join(" ") : String(pinyinPro.pinyin(text) || "");
    }
  } catch {
    out = "";
  }

  lruSet(text, out);
  self.postMessage({ type: "PINYIN_RESULT", text, out });
};

