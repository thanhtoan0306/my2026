export interface BeautifyOptions {
  sortKeys: boolean;
  indent: number;
}

function sortValue(value: unknown): unknown {
  if (value === null || typeof value !== "object") {
    return value;
  }
  if (Array.isArray(value)) {
    return value.map(sortValue);
  }
  const obj = value as Record<string, unknown>;
  const keys = Object.keys(obj).sort();
  const sorted: Record<string, unknown> = {};
  for (const k of keys) {
    sorted[k] = sortValue(obj[k]);
  }
  return sorted;
}

export function beautifyJson(
  raw: string,
  options: BeautifyOptions
): { ok: true; text: string } | { ok: false; error: string } {
  const trimmed = raw.trim();
  if (!trimmed) {
    return { ok: false, error: "No text to beautify." };
  }

  const candidate = extractJsonCandidate(trimmed);
  if (!candidate) {
    return { ok: false, error: "No valid JSON found in selection." };
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(candidate);
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e);
    return { ok: false, error: `Invalid JSON: ${msg}` };
  }

  const value = options.sortKeys ? sortValue(parsed) : parsed;
  const spaces = Math.max(0, options.indent);
  const text =
    spaces === 0
      ? JSON.stringify(value)
      : JSON.stringify(value, null, spaces);

  return { ok: true, text: text + "\n" };
}

/** Try full string, then scan for embedded `{...}` or `[...]` blocks. */
export function extractJsonCandidate(text: string): string | null {
  if (isValidJson(text)) {
    return text;
  }

  const blocks = findJsonBlocks(text);
  let best: string | null = null;
  for (const block of blocks) {
    if (!isValidJson(block)) {
      continue;
    }
    if (!best || block.length > best.length) {
      best = block;
    }
  }
  return best;
}

function isValidJson(text: string): boolean {
  try {
    JSON.parse(text);
    return true;
  } catch {
    return false;
  }
}

function findJsonBlocks(text: string): string[] {
  const results: string[] = [];
  for (let i = 0; i < text.length; i++) {
    const ch = text[i];
    if (ch !== "{" && ch !== "[") {
      continue;
    }
    const close = ch === "{" ? "}" : "]";
    const slice = extractBalanced(text, i, ch, close);
    if (slice) {
      results.push(slice);
    }
  }
  return results;
}

function extractBalanced(
  text: string,
  start: number,
  open: string,
  close: string
): string | null {
  let depth = 0;
  let inString = false;
  let escape = false;

  for (let i = start; i < text.length; i++) {
    const c = text[i];

    if (inString) {
      if (escape) {
        escape = false;
        continue;
      }
      if (c === "\\") {
        escape = true;
        continue;
      }
      if (c === '"') {
        inString = false;
      }
      continue;
    }

    if (c === '"') {
      inString = true;
      continue;
    }

    if (c === open) {
      depth++;
    } else if (c === close) {
      depth--;
      if (depth === 0) {
        return text.slice(start, i + 1);
      }
    }
  }

  return null;
}
