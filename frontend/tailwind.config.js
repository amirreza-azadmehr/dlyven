/** @type {import('tailwindcss').Config} */
const token = (name) => `rgb(var(--${name}) / <alpha-value>)`

export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: token('bg'),
        surface: token('surface'),
        elevated: token('elevated'),
        line: token('line'),
        fg: token('fg'),
        muted: token('muted'),
        subtle: token('subtle'),
        accent: {
          DEFAULT: token('accent'),
          2: token('accent-2'),
          fg: token('accent-fg'),
          ink: token('accent-ink'),
        },
        success: token('success'),
        warning: token('warning'),
        danger: token('danger'),
      },
      borderColor: {
        DEFAULT: 'rgb(var(--line) / 0.09)',
      },
      fontFamily: {
        sans: ['var(--font-sans)'],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Consolas', 'monospace'],
      },
      fontSize: {
        '2xs': ['0.6875rem', { lineHeight: '1rem' }],
      },
      borderRadius: {
        lg: '10px',
        xl: '12px',
        '2xl': '16px',
        '3xl': '20px',
      },
      boxShadow: {
        soft: '0 1px 0 0 rgb(255 255 255 / 0.03) inset, 0 1px 2px rgb(0 0 0 / 0.20), 0 8px 24px -12px rgb(0 0 0 / 0.45)',
        lift: '0 1px 0 0 rgb(255 255 255 / 0.04) inset, 0 12px 40px -12px rgb(0 0 0 / 0.55)',
        glow: '0 0 0 1px rgb(var(--accent) / 0.35), 0 10px 30px -10px rgb(var(--accent) / 0.55)',
        focus: '0 0 0 2px rgb(var(--bg)), 0 0 0 4px rgb(var(--accent-ink) / 0.9)',
      },
      keyframes: {
        shimmer: {
          '0%': { transform: 'translateX(calc(-100% * var(--dir)))' },
          '100%': { transform: 'translateX(calc(200% * var(--dir)))' },
        },
        'pulse-soft': {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.55' },
        },
      },
      animation: {
        shimmer: 'shimmer 1.6s ease-in-out infinite',
        'pulse-soft': 'pulse-soft 1.8s ease-in-out infinite',
      },
    },
  },
  plugins: [],
}
