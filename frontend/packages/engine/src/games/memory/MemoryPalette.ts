import type { MemoryPair } from "@usbi/schema";

export interface MemoryPairWithColor extends MemoryPair {
  color: string;
}

export interface MemoryCardStyle {
  [key: string]: string | number | undefined;
}

export const DEFAULT_MEMORY_BACK_COLOR = '#18529d';

export function createMemoryPairs(count: number): MemoryPairWithColor[] {
  return Array.from({ length: Math.max(0, Math.trunc(count)) }, (_, index) => ({
    id: `pair-${index + 1}`,
    content1: '',
    content2: '',
    color: generateMemoryColor(index),
  }));
}

export function createMemoryColorOptions(count: number): string[] {
  return Array.from({ length: Math.max(1, Math.trunc(count)) }, (_, index) => generateMemoryColor(index));
}

export function normalizeMemoryPairs(pairs: unknown): MemoryPairWithColor[] {
  const rawPairs = readMemoryPairs(pairs);
  const usedColors = new Set<string>();

  return rawPairs
    .map((pair, index) => {
      const id = typeof pair.id === 'string' && pair.id.trim().length > 0 ? pair.id : `pair-${index + 1}`;
      const content1 = typeof pair.content1 === 'string' ? pair.content1 : '';
      const content2 = typeof pair.content2 === 'string' ? pair.content2 : '';
      const color = pickUniqueColor(isHexColor(pair.color) ? pair.color : generateMemoryColor(index), usedColors, index);
      usedColors.add(color);
      return { id, content1, content2, color };
    });
}

export function filterPlayableMemoryPairs(pairs: MemoryPairWithColor[]): MemoryPairWithColor[] {
  return pairs.filter((pair) => pair.content1.length > 0 && pair.content2.length > 0);
}

export function normalizeMemoryBackColor(color: unknown): string {
  return isHexColor(color) ? color.toLowerCase() : DEFAULT_MEMORY_BACK_COLOR;
}

export function getMemoryReadableTextColor(background: string): '#0f172a' | '#ffffff' {
  const rgb = hexToRgb(background);
  if (!rgb) return '#ffffff';

  const whiteContrast = contrastRatio(rgb, { r: 255, g: 255, b: 255 });
  const darkContrast = contrastRatio(rgb, { r: 15, g: 23, b: 42 });
  return whiteContrast >= darkContrast ? '#ffffff' : '#0f172a';
}

export function getMemoryBackCardStyle(color: string): MemoryCardStyle {
  const backColor = normalizeMemoryBackColor(color);

  return {
    backgroundColor: backColor,
    backgroundImage: [
      'radial-gradient(circle at 20% 22%, rgba(255,255,255,0.25) 0 7%, transparent 8%)',
      'radial-gradient(circle at 78% 28%, rgba(255,255,255,0.18) 0 5%, transparent 6%)',
      'radial-gradient(circle at 30% 78%, rgba(255,255,255,0.12) 0 4%, transparent 5%)',
      'radial-gradient(circle at 72% 76%, rgba(0,0,0,0.12) 0 4%, transparent 5%)',
      'repeating-linear-gradient(135deg, rgba(255,255,255,0.10) 0 12px, rgba(255,255,255,0.02) 12px 24px)',
    ].join(', '),
    backgroundBlendMode: 'soft-light',
    color: getMemoryReadableTextColor(backColor),
  };
}

export function getMemoryFrontCardStyle(color: string): MemoryCardStyle {
  return {
    backgroundColor: color,
    backgroundImage: 'linear-gradient(180deg, rgba(255,255,255,0.18) 0%, rgba(255,255,255,0.04) 40%, rgba(0,0,0,0.08) 100%)',
    color: getMemoryReadableTextColor(color),
  };
}

function readMemoryPairs(pairs: unknown): Array<Partial<MemoryPairWithColor>> {
  if (Array.isArray(pairs)) return pairs as Array<Partial<MemoryPairWithColor>>;
  if (typeof pairs === 'object' && pairs !== null && Array.isArray((pairs as { pairs?: unknown }).pairs)) {
    return (pairs as { pairs: Array<Partial<MemoryPairWithColor>> }).pairs;
  }
  return [];
}

function pickUniqueColor(preferred: string, usedColors: Set<string>, index: number): string {
  const normalized = preferred.toLowerCase();
  if (!usedColors.has(normalized)) {
    return normalized;
  }

  for (let offset = 1; offset <= 720; offset++) {
    const candidate = generateMemoryColor(index + offset);
    if (!usedColors.has(candidate)) {
      return candidate;
    }
  }

  return normalized;
}

function generateMemoryColor(index: number): string {
  const hue = (index * 137.508 + 18) % 360;
  const saturation = 68;
  const lightness = 44 + ((index % 3) * 6);
  return hslToHex(hue, saturation, lightness);
}

function isHexColor(value: unknown): value is string {
  return typeof value === 'string' && /^#[0-9a-fA-F]{6}$/.test(value);
}

function hslToHex(h: number, s: number, l: number): string {
  const saturation = s / 100;
  const lightness = l / 100;
  const chroma = (1 - Math.abs(2 * lightness - 1)) * saturation;
  const huePrime = (h % 360) / 60;
  const x = chroma * (1 - Math.abs((huePrime % 2) - 1));
  const segment = getRgbSegment(huePrime, chroma, x);
  const match = lightness - chroma / 2;

  const red = Math.round((segment.r + match) * 255);
  const green = Math.round((segment.g + match) * 255);
  const blue = Math.round((segment.b + match) * 255);

  return `#${[red, green, blue].map((value) => value.toString(16).padStart(2, '0')).join('')}`;
}

function getRgbSegment(huePrime: number, chroma: number, x: number): { r: number; g: number; b: number } {
  if (huePrime < 1) return { r: chroma, g: x, b: 0 };
  if (huePrime < 2) return { r: x, g: chroma, b: 0 };
  if (huePrime < 3) return { r: 0, g: chroma, b: x };
  if (huePrime < 4) return { r: 0, g: x, b: chroma };
  if (huePrime < 5) return { r: x, g: 0, b: chroma };
  return { r: chroma, g: 0, b: x };
}

function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const normalized = hex.trim();
  if (!isHexColor(normalized)) return null;

  const red = Number.parseInt(normalized.slice(1, 3), 16);
  const green = Number.parseInt(normalized.slice(3, 5), 16);
  const blue = Number.parseInt(normalized.slice(5, 7), 16);

  return { r: red, g: green, b: blue };
}

function contrastRatio(a: { r: number; g: number; b: number }, b: { r: number; g: number; b: number }): number {
  const luminanceA = relativeLuminance(a);
  const luminanceB = relativeLuminance(b);
  const lighter = Math.max(luminanceA, luminanceB);
  const darker = Math.min(luminanceA, luminanceB);
  return (lighter + 0.05) / (darker + 0.05);
}

function relativeLuminance(color: { r: number; g: number; b: number }): number {
  const channels = [color.r, color.g, color.b].map((value) => {
    const normalized = value / 255;
    return normalized <= 0.03928
      ? normalized / 12.92
      : ((normalized + 0.055) / 1.055) ** 2.4;
  });

  return (channels[0] * 0.2126) + (channels[1] * 0.7152) + (channels[2] * 0.0722);
}
