// Small helpers for a fast page: byte sizes, a bundle budget, and responsive images.

/** formatBytes: under 1000 → "512 B"; under 1 000 000 → "1.5 kB"; otherwise "2.3 MB" (one decimal, powers of 1000). */
export function formatBytes(n: number): string {
  if (n < 1000) return `${n} B`;
  if (n < 1_000_000) return `${(n / 1000).toFixed(1)} kB`;
  return `${(n / 1_000_000).toFixed(1)} MB`;
}

export interface Asset {
  name: string;
  bytes: number; // compressed (gzip) size, as Vite's build prints it
}

/**
 * checkBudget adds up the .js files and the .css files and returns one message per type that is
 * over budget, JS first: "JS is 182.4 kB, over the 170.0 kB budget". Other files are ignored.
 * Within budget, it returns [].
 */
export function checkBudget(assets: Asset[], budget: { js: number; css: number }): string[] {
  const total = (ext: string) => assets.filter((a) => a.name.endsWith(ext)).reduce((sum, a) => sum + a.bytes, 0);
  const messages: string[] = [];
  for (const [label, ext, limit] of [
    ["JS", ".js", budget.js],
    ["CSS", ".css", budget.css],
  ] as const) {
    const size = total(ext);
    if (size > limit) messages.push(`${label} is ${formatBytes(size)}, over the ${formatBytes(limit)} budget`);
  }
  return messages;
}

/**
 * pickWidth chooses which image width to download: the smallest available width that is at least
 * cssWidth × devicePixelRatio, or the largest one if none is big enough. widths may be in any order.
 */
export function pickWidth(cssWidth: number, devicePixelRatio: number, widths: number[]): number {
  const needed = cssWidth * devicePixelRatio;
  const sorted = [...widths].sort((a, b) => a - b);
  return sorted.find((w) => w >= needed) ?? sorted[sorted.length - 1];
}

/** srcSet builds "/img/NAME-320.webp 320w, /img/NAME-640.webp 640w" with the widths in ascending order. */
export function srcSet(name: string, widths: number[]): string {
  return [...widths]
    .sort((a, b) => a - b)
    .map((w) => `/img/${name}-${w}.webp ${w}w`)
    .join(", ");
}
