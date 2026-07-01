import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        // shadcn aliases (reference CSS vars for compatibility)
        background: "#10131a",
        foreground: "#e1e2ec",
        card: "#1d2027",
        "card-foreground": "#e1e2ec",
        popover: "#272a31",
        "popover-foreground": "#e1e2ec",
        primary: "#adc6ff",
        "primary-foreground": "#002e6a",
        secondary: "#b9c8de",
        "secondary-foreground": "#233143",
        muted: "#191b23",
        "muted-foreground": "#c2c6d6",
        accent: "#4d8eff",
        "accent-foreground": "#00285d",
        destructive: "#ffb4ab",
        "destructive-foreground": "#690005",
        border: "#424754",
        input: "#424754",
        ring: "#adc6ff",

        // Material 3 surface hierarchy
        surface: {
          DEFAULT: "#10131a",
          dim: "#10131a",
          bright: "#363941",
          "container-lowest": "#0b0e15",
          "container-low": "#191b23",
          container: "#1d2027",
          "container-high": "#272a31",
          "container-highest": "#32353c",
          variant: "#32353c",
        },
        "on-surface": "#e1e2ec",
        "on-surface-variant": "#c2c6d6",
        "inverse-surface": "#e1e2ec",
        "inverse-on-surface": "#2e3038",
        outline: {
          DEFAULT: "#8c909f",
          variant: "#424754",
        },
        "surface-tint": "#adc6ff",

        // Primary
        "primary-container": "#4d8eff",
        "on-primary": "#002e6a",
        "on-primary-container": "#00285d",
        "inverse-primary": "#005ac2",
        "primary-fixed": "#d8e2ff",
        "primary-fixed-dim": "#adc6ff",
        "on-primary-fixed": "#001a42",
        "on-primary-fixed-variant": "#004395",

        // Secondary
        "secondary-container": "#39485a",
        "on-secondary": "#233143",
        "on-secondary-container": "#a7b6cc",
        "secondary-fixed": "#d4e4fa",
        "secondary-fixed-dim": "#b9c8de",
        "on-secondary-fixed": "#0d1c2d",
        "on-secondary-fixed-variant": "#39485a",

        // Tertiary
        tertiary: "#ffb786",
        "on-tertiary": "#502400",
        "tertiary-container": "#df7412",
        "on-tertiary-container": "#461f00",
        "tertiary-fixed": "#ffdcc6",
        "tertiary-fixed-dim": "#ffb786",
        "on-tertiary-fixed": "#311400",
        "on-tertiary-fixed-variant": "#723600",

        // Error
        error: "#ffb4ab",
        "on-error": "#690005",
        "error-container": "#93000a",
        "on-error-container": "#ffdad6",

        // Status
        success: "#4ade80",
        warning: "#facc15",
        info: "#adc6ff",
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
        mono: ["JetBrains Mono", "Fira Code", "monospace"],
        "headline-lg": ["Inter"],
        "headline-md": ["Inter"],
        "headline-sm": ["Inter"],
        "headline-lg-mobile": ["Inter"],
        "body-lg": ["Inter"],
        "body-md": ["Inter"],
        "label-md": ["Inter"],
        "mono-code": ["JetBrains Mono"],
      },
      fontSize: {
        "headline-lg": ["30px", { lineHeight: "38px", letterSpacing: "-0.02em", fontWeight: "700" }],
        "headline-md": ["24px", { lineHeight: "32px", letterSpacing: "-0.01em", fontWeight: "600" }],
        "headline-sm": ["20px", { lineHeight: "28px", fontWeight: "600" }],
        "headline-lg-mobile": ["24px", { lineHeight: "32px", fontWeight: "700" }],
        "body-lg": ["16px", { lineHeight: "24px", fontWeight: "400" }],
        "body-md": ["14px", { lineHeight: "20px", fontWeight: "400" }],
        "label-md": ["12px", { lineHeight: "16px", letterSpacing: "0.02em", fontWeight: "500" }],
        "mono-code": ["13px", { lineHeight: "20px", fontWeight: "400" }],
      },
      spacing: {
        base: "4px",
        xs: "4px",
        sm: "8px",
        md: "16px",
        lg: "24px",
        xl: "32px",
        gutter: "20px",
        margin: "24px",
      },
      borderRadius: {
        DEFAULT: "0.25rem",
        sm: "0.25rem",
        md: "0.75rem",
        lg: "0.5rem",
        xl: "0.75rem",
        "2xl": "1rem",
        "3xl": "1.5rem",
        full: "9999px",
      },
    },
  },
  plugins: [],
};

export default config;