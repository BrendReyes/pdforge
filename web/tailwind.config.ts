import type { Config } from 'tailwindcss';

export default {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        soft: '0 18px 42px rgba(0, 0, 0, 0.06)',
        card: '0 12px 28px rgba(0, 0, 0, 0.10)',
        glow: '0 0 20px rgba(255, 255, 255, 0.05)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.06)',
      },
      borderRadius: {
        xl2: '1.5rem',
      },
      colors: {
        ink: '#18181b',
        muted: '#71717a',
        line: 'rgba(161, 161, 170, 0.24)',
        brand: '#3f3f46',
      },
      ringColor: {
        DEFAULT: '#a1a1aa', // zinc-400 — prevents Tailwind's blue default ring
      },
      animation: {
        'fade-in-up': 'fadeInUp 0.45s ease-out both',
        'fade-in': 'fadeIn 0.35s ease-out both',
        'slide-in-right': 'slideInRight 0.3s ease-out both',
        'slide-out-right': 'slideOutRight 0.3s ease-in both',
        'pulse-subtle': 'pulseSubtle 2s ease-in-out infinite',
      },
      keyframes: {
        fadeInUp: {
          '0%': { opacity: '0', transform: 'translateY(12px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(16px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' },
        },
        slideOutRight: {
          '0%': { opacity: '1', transform: 'translateX(0)' },
          '100%': { opacity: '0', transform: 'translateX(16px)' },
        },
        pulseSubtle: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.7' },
        },
      },
    },
  },
  plugins: [],
} satisfies Config;