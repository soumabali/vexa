/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  distDir: 'dist',
  images: {
    unoptimized: true,
  },
  trailingSlash: true,
  assetPrefix: undefined,
  
  // Disable server-side features for static export
  experimental: {
    // Disable features that don't work with static export
    optimizeCss: false,
  },

  // Webpack config for Tauri compatibility
  webpack: (config, { isServer }) => {
    // Don't bundle Tauri APIs on server side
    if (isServer) {
      config.externals.push('@tauri-apps/api');
      config.externals.push('@tauri-apps/plugin-clipboard-manager');
      config.externals.push('@tauri-apps/plugin-global-shortcut');
      config.externals.push('@tauri-apps/plugin-notification');
      config.externals.push('@tauri-apps/plugin-shell');
      config.externals.push('@tauri-apps/plugin-window-state');
    }

    // Handle WebAssembly for xterm.js
    config.experiments = {
      ...config.experiments,
      asyncWebAssembly: true,
      syncWebAssembly: true,
    };

    return config;
  },

  // Headers for CSP (will be overridden by Tauri)
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'Content-Security-Policy',
            value: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self' ws://localhost:* wss://localhost:* ws://127.0.0.1:* wss://127.0.0.1:*; img-src 'self' data:; font-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
        ],
      },
    ];
  },
};

export default nextConfig;
