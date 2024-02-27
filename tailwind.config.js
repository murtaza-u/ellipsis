/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./view/**/*.templ'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Roboto', 'sans-serif'],
        mono: ['Ubuntu Mono', 'monospace'],
      },
    },
  },
  plugins: [require('@tailwindcss/forms'), require('daisyui')],
  daisyui: {
    themes: ['emerald', 'forest'],
  },
}
