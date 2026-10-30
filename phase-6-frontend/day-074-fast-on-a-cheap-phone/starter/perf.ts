// Small helpers for a fast page: byte sizes, a bundle budget, and responsive images.

/** formatBytes: under 1000 → "512 B"; under 1 000 000 → "1.5 kB"; otherwise "2.3 MB" (one decimal, powers of 1000). */
export function formatBytes(n: number): string {
  // Your code here
  return "";
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
  // Your code here
  return [];
}

/**
 * pickWidth chooses which image width to download: the smallest available width that is at least
 * cssWidth × devicePixelRatio, or the largest one if none is big enough. widths may be in any order.
 */
export function pickWidth(cssWidth: number, devicePixelRatio: number, widths: number[]): number {
  // Your code here
  return 0;
}

/** srcSet builds "/img/NAME-320.webp 320w, /img/NAME-640.webp 640w" with the widths in ascending order. */
export function srcSet(name: string, widths: number[]): string {
  // Your code here
  return "";
}
