import { browser } from '$app/environment';
import type { User } from '$lib/types';

const TOKEN_KEY = 'sk_token';
const USER_KEY = 'sk_user';

function readStorage<T>(key: string): T | null {
	if (!browser) return null;
	const raw = localStorage.getItem(key);
	if (!raw) return null;
	try {
		return JSON.parse(raw) as T;
	} catch {
		return null;
	}
}

function writeStorage(key: string, value: unknown) {
	if (!browser) return;
	if (value === null || value === undefined) {
		localStorage.removeItem(key);
	} else {
		localStorage.setItem(key, JSON.stringify(value));
	}
}

// State menggunakan Svelte 5 runes
let _user = $state<User | null>(readStorage<User>(USER_KEY));
let _token = $state<string | null>(browser ? localStorage.getItem(TOKEN_KEY) : null);

export const auth = {
	get user() {
		return _user;
	},
	get token() {
		return _token;
	},
	get isLoggedIn() {
		return _token !== null && _user !== null;
	}
};

export function setAuth(user: User, token: string) {
	_user = user;
	_token = token;
	if (browser) {
		localStorage.setItem(TOKEN_KEY, token);
		writeStorage(USER_KEY, user);
	}
}

export function clearAuth() {
	_user = null;
	_token = null;
	if (browser) {
		localStorage.removeItem(TOKEN_KEY);
		localStorage.removeItem(USER_KEY);
	}
}

export function hasRole(...roles: string[]) {
	if (!_user) return false;
	return roles.includes(_user.role);
}

export function isHR() {
	return hasRole('super_admin', 'hr_admin');
}

export function isHRorManager() {
	return hasRole('super_admin', 'hr_admin', 'manager');
}
