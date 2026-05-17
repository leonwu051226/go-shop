/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        void: '#0b0c10',
        abyss: '#0d0e15',
        neon: {
          cyan: '#00f3ff',
          pink: '#ff003c',
          yellow: '#fcee0a',
        },
      },
      boxShadow: {
        glass: '0 4px 30px rgba(0, 0, 0, 0.5)',
        cyan: '0 0 18px rgba(0, 243, 255, 0.55)',
        pink: '0 0 18px rgba(255, 0, 60, 0.55)',
      },
      fontFamily: {
        display: ['Rajdhani', 'Inter', 'system-ui', 'sans-serif'],
      },
      backgroundImage: {
        'cyber-grid':
          'linear-gradient(rgba(0, 243, 255, 0.08) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 0, 60, 0.07) 1px, transparent 1px)',
      },
    },
  },
  plugins: [],
}
