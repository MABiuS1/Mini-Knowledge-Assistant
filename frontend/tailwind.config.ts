import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        ink: "#f5f7ff",
        muted: "#a8b0d8",
        surface: "#050713",
        line: "rgba(190, 199, 255, 0.18)",
        brand: "#7c5cff",
      },
    },
  },
  plugins: [],
};

export default config;
