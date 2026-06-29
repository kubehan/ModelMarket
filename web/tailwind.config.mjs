/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,ts}'],
  theme: {
    screens: {
      'sm':  '640px',
      'md':  '768px',
      'lg':  '1024px',
      'xl':  '1280px',
      '2xl': '1536px',
      '3xl': '1920px',
      '4xl': '2560px',
    },
    extend: {
      colors: {
        bg: {
          base:   'var(--bg-base)',
          1:      'var(--bg-1)',
          2:      'var(--bg-2)',
          3:      'var(--bg-3)',
        },
        fg: {
          1:      'var(--fg-1)',
          2:      'var(--fg-2)',
          3:      'var(--fg-3)',
          4:      'var(--fg-4)',
        },
        line: {
          1:      'var(--line-1)',
          2:      'var(--line-2)',
          3:      'var(--line-3)',
        },
        violet: {
          400: 'var(--a-violet-soft)',
          500: 'var(--a-violet)',
          600: 'var(--a-violet)',
        },
        cyan: {
          400: 'var(--a-cyan-soft)',
          500: 'var(--a-cyan)',
          600: 'var(--a-cyan)',
        },
        emerald: {
          400: 'var(--a-emerald-soft)',
          500: 'var(--a-emerald)',
        },
        amber: {
          400: 'var(--a-amber-soft)',
          500: 'var(--a-amber)',
        },
        rose: {
          400: 'var(--a-rose-soft)',
          500: 'var(--a-rose)',
        }
      },
      fontFamily: {
        sans: [
          'Inter', 'Noto Sans SC',
          '-apple-system', 'BlinkMacSystemFont',
          'PingFang SC', 'Microsoft YaHei',
          'Segoe UI', 'Helvetica Neue', 'sans-serif'
        ],
        mono: ['JetBrains Mono', 'SF Mono', 'Menlo', 'monospace'],
        display: ['Inter', 'Noto Sans SC', 'sans-serif']
      },
      backgroundImage: {
        'grid-fade': "linear-gradient(to bottom, rgba(255,255,255,0.04), transparent 80%)",
        'gradient-text': 'var(--grad-text)',
        'gradient-brand': 'var(--grad-brand)',
        'gradient-radial': 'radial-gradient(circle at center, var(--tw-gradient-stops))',
      },
      boxShadow: {
        'glow-violet': '0 0 0 1px var(--a-violet), 0 8px 24px -8px var(--a-violet)',
        'glow-cyan':   '0 0 0 1px var(--a-cyan), 0 8px 24px -8px var(--a-cyan)',
        'card':        'var(--shadow-card)',
        'card-hover':  '0 16px 48px -16px rgba(var(--accent-a-rgb), 0.35), inset 0 1px 0 0 rgba(255,255,255,0.08)',
      },
      borderRadius: {
        'xl2': '1.25rem',
        'xl3': '1.75rem'
      },
      keyframes: {
        'pulse-soft': {
          '0%, 100%': { opacity: '0.4' },
          '50%':      { opacity: '1' }
        },
        'fade-up': {
          '0%':   { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        'gradient-x': {
          '0%, 100%': { backgroundPosition: '0% 50%' },
          '50%':       { backgroundPosition: '100% 50%' }
        },
        'aurora': {
          '0%, 100%': { transform: 'translate3d(0,0,0) rotate(0deg)' },
          '50%':       { transform: 'translate3d(30px,-20px,0) rotate(8deg)' }
        }
      },
      animation: {
        'pulse-soft': 'pulse-soft 3s ease-in-out infinite',
        'fade-up':    'fade-up 0.6s ease-out both',
        'gradient-x': 'gradient-x 8s ease infinite',
        'aurora':     'aurora 18s ease-in-out infinite'
      }
    }
  },
  plugins: []
}
