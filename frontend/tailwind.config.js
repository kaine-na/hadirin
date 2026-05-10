import forms from '@tailwindcss/forms';

/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			colors: {
				primary: {
					50: '#f0f4f8',
					100: '#d9e2ec',
					200: '#bcccdc',
					300: '#9fb3c8',
					400: '#617d9b',
					500: '#3a5e82',
					600: '#2a4a6b',
					700: '#1e3a5f',
					800: '#16293f',
					900: '#0f1c2d',
					DEFAULT: '#1e3a5f'
				},
				accent: {
					50: '#eff6ff',
					100: '#dbeafe',
					200: '#bfdbfe',
					300: '#93c5fd',
					400: '#60a5fa',
					500: '#3b82f6',
					600: '#2563eb',
					700: '#1d4ed8',
					DEFAULT: '#3b82f6'
				},
				success: {
					50: '#ecfdf5',
					100: '#d1fae5',
					500: '#10b981',
					600: '#059669',
					700: '#047857',
					DEFAULT: '#10b981'
				},
				warning: {
					50: '#fffbeb',
					100: '#fef3c7',
					500: '#f59e0b',
					600: '#d97706',
					700: '#b45309',
					DEFAULT: '#f59e0b'
				},
				danger: {
					50: '#fef2f2',
					100: '#fee2e2',
					500: '#ef4444',
					600: '#dc2626',
					700: '#b91c1c',
					DEFAULT: '#ef4444'
				},
				surface: '#ffffff',
				canvas: '#f8fafc'
			},
			fontFamily: {
				sans: ['Inter', 'system-ui', 'sans-serif']
			},
			boxShadow: {
				subtle: '0 1px 2px 0 rgb(15 23 42 / 0.04), 0 1px 3px 0 rgb(15 23 42 / 0.06)',
				raise: '0 4px 10px -2px rgb(15 23 42 / 0.08), 0 2px 4px -2px rgb(15 23 42 / 0.05)',
				lift: '0 12px 24px -8px rgb(15 23 42 / 0.12), 0 4px 8px -4px rgb(15 23 42 / 0.06)'
			},
			keyframes: {
				'fade-in': {
					'0%': { opacity: '0', transform: 'translateY(4px)' },
					'100%': { opacity: '1', transform: 'translateY(0)' }
				},
				'modal-in': {
					'0%': { opacity: '0', transform: 'scale(0.96)' },
					'100%': { opacity: '1', transform: 'scale(1)' }
				},
				'slide-in': {
					'0%': { opacity: '0', transform: 'translateX(12px)' },
					'100%': { opacity: '1', transform: 'translateX(0)' }
				},
				'pulse-ring': {
					'0%': { transform: 'scale(1)', opacity: '0.6' },
					'100%': { transform: 'scale(1.6)', opacity: '0' }
				},
				shimmer: {
					'0%': { backgroundPosition: '-400px 0' },
					'100%': { backgroundPosition: '400px 0' }
				},
				'spin-slow': {
					to: { transform: 'rotate(360deg)' }
				}
			},
			animation: {
				'fade-in': 'fade-in 260ms ease-out both',
				'modal-in': 'modal-in 180ms ease-out both',
				'slide-in': 'slide-in 220ms ease-out both',
				'pulse-ring': 'pulse-ring 1.8s cubic-bezier(0.4, 0, 0.6, 1) infinite',
				shimmer: 'shimmer 1.4s linear infinite',
				'spin-slow': 'spin-slow 3s linear infinite'
			},
			transitionDuration: {
				150: '150ms'
			}
		}
	},
	plugins: [forms]
};
