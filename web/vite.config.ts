import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 8890,
    proxy: {
      '/v1': {
        target: 'http://localhost:8891',
        changeOrigin: true,
      },
    },
  },
});
