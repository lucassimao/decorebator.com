import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // Brand Colors
        primary: {
          50: '#FFF5F2',
          100: '#FFE8E0',
          200: '#FFD0C1',
          300: '#FFB8A2',
          400: '#FFA083',
          500: '#FF6B35', // Main primary
          600: '#E65220',
          700: '#CC3D15',
          800: '#B32C0D',
          900: '#991F08',
        },
        secondary: {
          50: '#E6F3F9',
          100: '#CCE7F3',
          200: '#99CFE7',
          300: '#66B7DB',
          400: '#339FCF',
          500: '#004E89', // Main secondary
          600: '#003E6E',
          700: '#002F53',
          800: '#002038',
          900: '#00111D',
        },
        accent: {
          50: '#F5EDFC',
          100: '#EBDCF9',
          200: '#D7B9F3',
          300: '#C396ED',
          400: '#AF73E7',
          500: '#7B2CBF', // Main accent
          600: '#6323A0',
          700: '#4B1A81',
          800: '#331162',
          900: '#1B0843',
        },
        success: {
          50: '#E6FCF5',
          100: '#CDF9EB',
          200: '#9BF3D7',
          300: '#69EDC3',
          400: '#37E7AF',
          500: '#06D6A0', // Main success
          600: '#05AB80',
          700: '#048060',
          800: '#035540',
          900: '#022B20',
        },
        // Neutrals
        slate: {
          50: '#F8FAFC',
          100: '#F1F5F9',
          200: '#E2E8F0',
          300: '#CBD5E1',
          400: '#94A3B8',
          500: '#64748B',
          600: '#475569',
          700: '#334155',
          800: '#1E293B',
          900: '#0F172A',
        },
      },
      fontFamily: {
        sans: ['var(--font-geist-sans)', 'Inter', 'system-ui', 'sans-serif'],
        mono: ['var(--font-geist-mono)', 'JetBrains Mono', 'monospace'],
      },
      fontSize: {
        // Custom typography scale
        'display-1': ['4.5rem', { lineHeight: '1.1', fontWeight: '700' }], // 72px
        'display-2': ['3.75rem', { lineHeight: '1.2', fontWeight: '700' }], // 60px
        h1: ['3rem', { lineHeight: '1.2', fontWeight: '700' }], // 48px
        h2: ['2.25rem', { lineHeight: '1.3', fontWeight: '600' }], // 36px
        h3: ['1.5rem', { lineHeight: '1.4', fontWeight: '600' }], // 24px
        h4: ['1.25rem', { lineHeight: '1.5', fontWeight: '600' }], // 20px
        'body-lg': ['1.25rem', { lineHeight: '1.6', fontWeight: '400' }], // 20px
        body: ['1rem', { lineHeight: '1.75', fontWeight: '400' }], // 16px
        'body-sm': ['0.875rem', { lineHeight: '1.5', fontWeight: '400' }], // 14px
      },
      spacing: {
        // 8pt grid system
        '18': '4.5rem', // 72px
        '22': '5.5rem', // 88px
        '26': '6.5rem', // 104px
        '30': '7.5rem', // 120px
      },
      borderRadius: {
        '4xl': '2rem',
        '5xl': '2.5rem',
      },
      boxShadow: {
        soft: '0 2px 8px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06)',
        medium: '0 4px 16px rgba(0, 0, 0, 0.08), 0 2px 4px rgba(0, 0, 0, 0.06)',
        hard: '0 8px 32px rgba(0, 0, 0, 0.12), 0 4px 8px rgba(0, 0, 0, 0.08)',
        glow: '0 0 24px rgba(255, 107, 53, 0.3)',
        'glow-purple': '0 0 24px rgba(123, 44, 191, 0.3)',
      },
      animation: {
        'fade-in': 'fadeIn 0.6s ease-out',
        'slide-up': 'slideUp 0.6s ease-out',
        'slide-down': 'slideDown 0.6s ease-out',
        'scale-in': 'scaleIn 0.4s ease-out',
        'gradient-shift': 'gradientShift 8s ease infinite',
        float: 'float 6s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        gradientShift: {
          '0%, 100%': { backgroundPosition: '0% 50%' },
          '50%': { backgroundPosition: '100% 50%' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%': { transform: 'translateY(-20px)' },
        },
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-mesh':
          'radial-gradient(at 40% 20%, #7B2CBF 0px, transparent 50%), radial-gradient(at 80% 0%, #004E89 0px, transparent 50%), radial-gradient(at 0% 50%, #FF6B35 0px, transparent 50%)',
      },
    },
  },
  plugins: [],
}

export default config
