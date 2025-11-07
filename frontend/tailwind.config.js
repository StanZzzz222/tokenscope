const { heroui } = require("@heroui/theme");
const lineclamp = require('@tailwindcss/line-clamp'),

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
    "./node_modules/@heroui/theme/dist/components/*"
  ],
  darkMode: "class",
  plugins: [heroui(), lineclamp()],
};