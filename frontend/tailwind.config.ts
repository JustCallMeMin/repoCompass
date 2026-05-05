import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}", "./lib/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#17201d",
        moss: "#415f4d",
        field: "#f4f1e8",
        paper: "#fffdf7",
        rust: "#b25b35",
        gold: "#d59b3b",
      },
      fontFamily: {
        display: ["var(--font-display)", "Georgia", "serif"],
        body: ["var(--font-body)", "Verdana", "sans-serif"],
      },
    },
  },
  plugins: [],
};

export default config;
