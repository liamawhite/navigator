// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/**
 * CSS Custom Properties-based theme system for graph visualization
 * This provides more efficient theme detection and reduces layout thrashing
 */

interface ThemeColors {
    background: string;
    foreground: string;
    primary: string;
    secondary: string;
    muted: string;
    mutedForeground: string;
    border: string;
    nodeBackground: string;
    nodeStroke: string;
    textStroke: string;
}

/**
 * CSS custom property names used for theme colors
 */
const CSS_VARIABLES = {
    background: '--background',
    foreground: '--foreground',
    primary: '--primary',
    secondary: '--secondary',
    muted: '--muted',
    mutedForeground: '--muted-foreground',
    border: '--border',
    // Graph-specific variables
    nodeBackground: '--graph-node-background',
    nodeStroke: '--graph-node-stroke',
    textStroke: '--graph-text-stroke',
} as const;

/**
 * Cache for CSS custom property values to avoid repeated DOM queries
 */
class ThemeCache {
    private cache = new Map<string, string>();
    private lastUpdate = 0;
    private readonly cacheTimeout = 100; // 100ms cache timeout

    public get(property: string): string {
        const now = Date.now();
        const cacheKey = property;

        // Use cache if it's recent
        if (
            now - this.lastUpdate < this.cacheTimeout &&
            this.cache.has(cacheKey)
        ) {
            return this.cache.get(cacheKey)!;
        }

        // Update cache
        const value = this.getCSSProperty(property);
        this.cache.set(cacheKey, value);
        this.lastUpdate = now;

        return value;
    }

    private getCSSProperty(property: string): string {
        if (typeof window === 'undefined') {
            return this.getFallbackValue(property);
        }

        try {
            const value = getComputedStyle(document.documentElement)
                .getPropertyValue(property)
                .trim();

            // Return fallback if CSS variable is not defined
            return value || this.getFallbackValue(property);
        } catch (error) {
            console.warn(`Failed to read CSS property ${property}:`, error);
            return this.getFallbackValue(property);
        }
    }

    private getFallbackValue(property: string): string {
        // Fallback values for SSR or when CSS variables are not available
        const fallbacks: Record<string, string> = {
            '--background': '#ffffff',
            '--foreground': '#0f0f0f',
            '--primary': '#3b82f6',
            '--secondary': '#f3f4f6',
            '--muted': '#f3f4f6',
            '--muted-foreground': '#6b7280',
            '--border': '#e5e7eb',
            '--graph-node-background': '#e5e7eb',
            '--graph-node-stroke': '#9ca3af',
            '--graph-text-stroke': '#ffffff',
        };

        return fallbacks[property] || '#000000';
    }

    public invalidate(): void {
        this.cache.clear();
        this.lastUpdate = 0;
    }
}

// Singleton theme cache
const themeCache = new ThemeCache();

/**
 * Get theme colors using CSS custom properties
 * This is more efficient than checking classList and avoids layout thrashing
 */
export function getThemeColorsFromCSS(): ThemeColors {
    return {
        background: themeCache.get(CSS_VARIABLES.background),
        foreground: themeCache.get(CSS_VARIABLES.foreground),
        primary: themeCache.get(CSS_VARIABLES.primary),
        secondary: themeCache.get(CSS_VARIABLES.secondary),
        muted: themeCache.get(CSS_VARIABLES.muted),
        mutedForeground: themeCache.get(CSS_VARIABLES.mutedForeground),
        border: themeCache.get(CSS_VARIABLES.border),
        nodeBackground: themeCache.get(CSS_VARIABLES.nodeBackground),
        nodeStroke: themeCache.get(CSS_VARIABLES.nodeStroke),
        textStroke: themeCache.get(CSS_VARIABLES.textStroke),
    };
}

/**
 * Convert HSL color string to hex (for D3 compatibility)
 */
function hslToHex(hsl: string): string {
    // Parse HSL string like "220 14.3% 95.9%"
    const match = hsl.match(/(\d+\.?\d*)\s+(\d+\.?\d*)%\s+(\d+\.?\d*)%/);
    if (!match) return hsl; // Return original if not HSL format

    const h = parseFloat(match[1]) / 360;
    const s = parseFloat(match[2]) / 100;
    const l = parseFloat(match[3]) / 100;

    const hue2rgb = (p: number, q: number, t: number) => {
        if (t < 0) t += 1;
        if (t > 1) t -= 1;
        if (t < 1 / 6) return p + (q - p) * 6 * t;
        if (t < 1 / 2) return q;
        if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
        return p;
    };

    let r, g, b;
    if (s === 0) {
        r = g = b = l; // achromatic
    } else {
        const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
        const p = 2 * l - q;
        r = hue2rgb(p, q, h + 1 / 3);
        g = hue2rgb(p, q, h);
        b = hue2rgb(p, q, h - 1 / 3);
    }

    const toHex = (c: number) => {
        const hex = Math.round(c * 255).toString(16);
        return hex.length === 1 ? '0' + hex : hex;
    };

    return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
}

/**
 * Get a processed theme color that's compatible with D3
 */
export function getProcessedThemeColor(colorKey: keyof ThemeColors): string {
    const colors = getThemeColorsFromCSS();
    const color = colors[colorKey];

    // Convert HSL to hex if needed
    if (color.includes('%')) {
        return hslToHex(color);
    }

    return color;
}

/**
 * Check if current theme is dark based on CSS custom properties
 * This is more reliable than checking class names
 */
export function isDarkTheme(): boolean {
    const backgroundColor = getProcessedThemeColor('background');

    // Parse hex color and calculate brightness
    const hex = backgroundColor.replace('#', '');
    const r = parseInt(hex.substr(0, 2), 16);
    const g = parseInt(hex.substr(2, 2), 16);
    const b = parseInt(hex.substr(4, 2), 16);

    // Calculate brightness using standard formula
    const brightness = (r * 299 + g * 587 + b * 114) / 1000;

    return brightness < 128; // Dark theme if brightness is low
}

/**
 * Set up CSS custom properties for graph-specific theming
 * This allows the graph to react to theme changes automatically
 */
export function initializeGraphTheme(): void {
    if (typeof window === 'undefined') return;

    const style = document.createElement('style');
    style.textContent = `
        :root {
            --graph-node-background: hsl(var(--secondary));
            --graph-node-stroke: hsl(var(--border));
            --graph-text-stroke: hsl(var(--background));
        }
        
        .dark {
            --graph-node-background: hsl(var(--secondary));
            --graph-node-stroke: hsl(var(--muted-foreground));
            --graph-text-stroke: hsl(var(--background));
        }
    `;

    document.head.appendChild(style);
}

/**
 * Create a theme change observer that invalidates the cache
 */
export function createThemeObserver(
    callback: () => void
): MutationObserver | null {
    if (typeof window === 'undefined') return null;

    const observer = new MutationObserver((mutations) => {
        let themeChanged = false;

        mutations.forEach((mutation) => {
            if (
                mutation.type === 'attributes' &&
                mutation.attributeName === 'class'
            ) {
                themeChanged = true;
            }
        });

        if (themeChanged) {
            // Invalidate cache and notify callback
            themeCache.invalidate();
            callback();
        }
    });

    observer.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ['class'],
    });

    return observer;
}

/**
 * Force invalidation of theme cache (useful for testing or manual refresh)
 */
export function invalidateThemeCache(): void {
    themeCache.invalidate();
}
