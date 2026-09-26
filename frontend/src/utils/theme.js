/**
 * Windows 11 Fluent 2 Dynamic Accent & Contrast System
 * Automatically optimizes luminance, contrast, and color ramps
 * to match Windows SystemUI readability standards in Dark and Light modes.
 */

export function hexToRgb(hex) {
  if (!hex) return { r: 0, g: 120, b: 212 };
  let clean = hex.replace('#', '').trim();
  if (clean.length === 3) {
    clean = clean.split('').map(c => c + c).join('');
  }
  const num = parseInt(clean.substring(0, 6), 16);
  if (isNaN(num)) return { r: 0, g: 120, b: 212 };
  return {
    r: (num >> 16) & 255,
    g: (num >> 8) & 255,
    b: num & 255
  };
}

export function rgbToHsl(r, g, b) {
  r /= 255;
  g /= 255;
  b /= 255;
  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  let h = 0;
  let s = 0;
  let l = (max + min) / 2;

  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    switch (max) {
      case r:
        h = (g - b) / d + (g < b ? 6 : 0);
        break;
      case g:
        h = (b - r) / d + 2;
        break;
      case b:
        h = (r - g) / d + 4;
        break;
    }
    h /= 6;
  }

  return {
    h: Math.round(h * 360),
    s: Math.round(s * 100),
    l: Math.round(l * 100)
  };
}

export function hslToHex(h, s, l) {
  s /= 100;
  l /= 100;
  const k = n => (n + h / 30) % 12;
  const a = s * Math.min(l, 1 - l);
  const f = n =>
    l - a * Math.max(-1, Math.min(k(n) - 3, Math.min(9 - k(n), 1)));
  const toHex = x => {
    const val = Math.round(x * 255);
    return val.toString(16).padStart(2, '0');
  };
  return `#${toHex(f(0))}${toHex(f(8))}${toHex(f(4))}`;
}

/**
 * Computes optimized Fluent 2 CSS color variables for any accent hex
 */
export function computeAccentPalette(hexColor, isDark = true) {
  const rgb = hexToRgb(hexColor);
  const { h, s, l } = rgbToHsl(rgb.r, rgb.g, rgb.b);

  // Directly preserve the user's chosen accent color for primary buttons
  const primaryHex = hexColor;

  if (isDark) {
    // Windows Dark Mode:
    // Text, icons, and borders must be bright and luminous (~75-78% lightness)
    const textLightness = 76;
    const textSaturation = Math.min(Math.max(s, 65), 90);
    const readableTextHex = hslToHex(h, textSaturation, textLightness);

    const hoverHex = hslToHex(h, s, Math.min(l + 7, 85));
    const activeHex = hslToHex(h, s, Math.max(l - 6, 20));

    return {
      '--accent-primary': primaryHex,
      '--accent-hover': hoverHex,
      '--accent-active': activeHex,
      '--accent-light': readableTextHex,
      '--accent-text': l > 68 ? '#000000' : '#ffffff',
      '--accent-contrast-text': '#ffffff',
      '--accent-badge-bg': `hsla(${h}, ${textSaturation}%, ${textLightness}%, 0.14)`,
      '--accent-badge-border': `hsla(${h}, ${textSaturation}%, ${textLightness}%, 0.30)`,
      '--accent-glow': `hsla(${h}, ${s}%, ${l}%, 0.32)`,
      '--border-focus': readableTextHex
    };
  } else {
    // Windows Light Mode:
    // Text, icons, and borders must be deep and crisp (~34-38% lightness)
    const textLightness = 36;
    const textSaturation = Math.min(Math.max(s, 70), 95);
    const readableTextHex = hslToHex(h, textSaturation, textLightness);

    const hoverHex = hslToHex(h, s, Math.max(l - 6, 18));
    const activeHex = hslToHex(h, s, Math.max(l - 10, 14));

    return {
      '--accent-primary': primaryHex,
      '--accent-hover': hoverHex,
      '--accent-active': activeHex,
      '--accent-light': readableTextHex,
      '--accent-text': l > 68 ? '#000000' : '#ffffff',
      '--accent-contrast-text': '#ffffff',
      '--accent-badge-bg': `hsla(${h}, ${textSaturation}%, ${textLightness}%, 0.09)`,
      '--accent-badge-border': `hsla(${h}, ${textSaturation}%, ${textLightness}%, 0.24)`,
      '--accent-glow': `hsla(${h}, ${s}%, ${l}%, 0.22)`,
      '--border-focus': readableTextHex
    };
  }
}

/**
 * Applies computed variables to the root document
 */
export function applyAccentToDocument(hexColor, isDark = true) {
  if (typeof document === 'undefined') return;
  const palette = computeAccentPalette(hexColor, isDark);
  const root = document.documentElement;
  for (const [key, value] of Object.entries(palette)) {
    root.style.setProperty(key, value);
  }
}
