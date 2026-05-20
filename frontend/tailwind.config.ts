import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        ink: "#172033",
        muted: "#667085",
        surface: "#f6f8fb",
        line: "#d8dee8",
        brand: "#2563eb",
      },
    },
  },
  plugins: [],
};

export default config;

