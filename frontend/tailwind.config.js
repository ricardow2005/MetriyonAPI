export default {
  content: ["./index.html", "./src/**/*.{ts,html}"],
  theme: {
    extend: {
      fontFamily: { sans: ["Inter", "Segoe UI", "sans-serif"], mono: ["JetBrains Mono", "Cascadia Code", "monospace"] },
      colors: { forge: { 400: "#ffb04a", 500: "#f59e32", 600: "#db7f1d" } },
      boxShadow: { glow: "0 0 0 1px rgba(245,158,50,.15), 0 12px 35px rgba(0,0,0,.25)" }
    }
  },
  plugins: []
};
