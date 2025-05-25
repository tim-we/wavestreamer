import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

const now = new Date();

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [preact()],
	define: {
		__BUILD_DATE__: JSON.stringify(now.toISOString())
	},
});
