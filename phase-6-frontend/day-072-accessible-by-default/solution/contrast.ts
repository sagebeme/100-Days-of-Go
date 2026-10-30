// WCAG 2 colour contrast. Formulas: https://www.w3.org/TR/WCAG21/#dfn-relative-luminance

/**
 * relativeLuminance takes "#rrggbb" and returns a number from 0 (black) to 1 (white).
 * For each channel: c = value / 255; linear = c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4.
 * Then L = 0.2126 * R + 0.7152 * G + 0.0722 * B.
 */
export function relativeLuminance(hex: string): number {
  const m = /^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i.exec(hex);
  if (!m) throw new Error(`not a #rrggbb colour: ${hex}`);
  const [r, g, b] = m.slice(1).map((h) => {
    const c = parseInt(h, 16) / 255;
    return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** contrastRatio is (lighter + 0.05) / (darker + 0.05), from 1 (same colour) to 21 (black on white). */
export function contrastRatio(a: string, b: string): number {
  const la = relativeLuminance(a);
  const lb = relativeLuminance(b);
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
}

/** meetsAA: normal text needs a ratio of at least 4.5, large text (24px, or 18.66px bold) at least 3. */
export function meetsAA(ratio: number, largeText: boolean): boolean {
  return ratio >= (largeText ? 3 : 4.5);
}
