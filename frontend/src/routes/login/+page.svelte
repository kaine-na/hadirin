<script lang="ts">
	import { goto } from '$app/navigation';
	import { authApi } from '$lib/api';
	import { setAuth, auth } from '$lib/stores/auth.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { onMount } from 'svelte';
	import { Mail, Lock, AlertCircle, Eye, EyeOff, Building2, Users, BarChart3, Shield } from 'lucide-svelte';

	let email = $state('');
	let password = $state('');
	let showPassword = $state(false);
	let loading = $state(false);
	let error = $state('');

	onMount(() => {
		if (auth.isLoggedIn) {
			goto('/dashboard');
		}
	});

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		if (!email || !password) {
			error = 'Email dan password wajib diisi';
			return;
		}
		loading = true;
		try {
			const res = await authApi.login({ email, password });
			setAuth(res.user, res.token);
			toast.success(`Selamat datang, ${res.user.name}`);
			await goto('/dashboard');
		} catch (err) {
			const msg = err instanceof Error ? err.message : 'Gagal login. Periksa email dan password Anda.';
			error = msg;
			toast.error(msg);
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Login — SaaS Karyawan</title>
</svelte:head>

<div class="min-h-screen flex">
	<!-- Left: Branding panel -->
	<div class="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-[#1e3a5f] to-[#2d5a8e] flex-col items-center justify-center p-12 text-white relative overflow-hidden">
		<!-- Decorative circles -->
		<div class="pointer-events-none absolute -right-16 -top-16 h-72 w-72 rounded-full bg-white/5"></div>
		<div class="pointer-events-none absolute -bottom-20 -left-8 h-56 w-56 rounded-full bg-white/5"></div>
		<div class="pointer-events-none absolute top-1/2 right-8 h-32 w-32 rounded-full bg-white/5"></div>

		<div class="relative z-10 max-w-md w-full">
			<!-- Building2 icon besar -->
			<div class="flex justify-center mb-8">
				<div class="flex h-20 w-20 items-center justify-center rounded-2xl bg-white/15 backdrop-blur-sm">
					<Building2 class="h-10 w-10 text-white" />
				</div>
			</div>

			<!-- Judul -->
			<h1 class="text-4xl font-bold text-center mb-3">SaaS Karyawan</h1>

			<!-- Tagline -->
			<p class="text-lg text-blue-200 text-center mb-12">
				Kelola karyawan lebih cerdas dengan AI
			</p>

			<!-- Feature bullets -->
			<div class="space-y-5">
				<div class="flex items-center gap-4">
					<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white/15">
						<Users class="h-5 w-5 text-white" />
					</div>
					<div>
						<p class="font-semibold text-white">Manajemen Karyawan Terpusat</p>
						<p class="text-sm text-blue-200">Data karyawan, absensi, dan dokumen dalam satu platform</p>
					</div>
				</div>

				<div class="flex items-center gap-4">
					<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white/15">
						<BarChart3 class="h-5 w-5 text-white" />
					</div>
					<div>
						<p class="font-semibold text-white">Analitik & Laporan Otomatis</p>
						<p class="text-sm text-blue-200">Insight kinerja tim dengan dashboard real-time</p>
					</div>
				</div>

				<div class="flex items-center gap-4">
					<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white/15">
						<Shield class="h-5 w-5 text-white" />
					</div>
					<div>
						<p class="font-semibold text-white">Keamanan Data Terjamin</p>
						<p class="text-sm text-blue-200">Enkripsi end-to-end dan kontrol akses berbasis peran</p>
					</div>
				</div>
			</div>

			<p class="mt-16 text-center text-xs text-blue-300">© 2026 Mirava Office. All rights reserved.</p>
		</div>
	</div>

	<!-- Right: Form panel -->
	<div class="flex-1 flex items-center justify-center p-8 bg-slate-50">
		<div class="w-full max-w-md page-enter">
			<!-- Logo mobile (hanya tampil di mobile) -->
			<div class="flex items-center justify-center gap-3 mb-8 lg:hidden">
				<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-[#1e3a5f] to-[#2d5a8e]">
					<Building2 class="h-5 w-5 text-white" />
				</div>
				<span class="text-lg font-bold text-slate-900">SaaS Karyawan</span>
			</div>

			<!-- Card form -->
			<div class="bg-white rounded-2xl shadow-lg p-8">
				<!-- Header -->
				<div class="mb-8">
					<h2 class="text-2xl font-bold text-slate-900">Selamat Datang</h2>
					<p class="mt-2 text-sm text-slate-500">
						Masuk ke dashboard manajemen karyawan Anda
					</p>
				</div>

				<!-- Form -->
				<form onsubmit={handleSubmit} class="space-y-5">
					<!-- Email input -->
					<div>
						<label for="email" class="label">Email</label>
						<div class="relative">
							<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
								<Mail class="h-4 w-4 text-slate-400" />
							</div>
							<input
								id="email"
								name="email"
								type="email"
								autocomplete="email"
								required
								placeholder="nama@perusahaan.com"
								value={email}
								oninput={(e) => (email = (e.target as HTMLInputElement).value)}
								class="block w-full rounded-lg border border-slate-200 bg-white pl-10 pr-3 py-2.5 text-sm text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-slate-50 disabled:text-slate-500 transition-colors"
							/>
						</div>
					</div>

					<!-- Password input -->
					<div>
						<label for="password" class="label">Password</label>
						<div class="relative">
							<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
								<Lock class="h-4 w-4 text-slate-400" />
							</div>
							<input
								id="password"
								name="password"
								type={showPassword ? 'text' : 'password'}
								autocomplete="current-password"
								required
								placeholder="••••••••"
								value={password}
								oninput={(e) => (password = (e.target as HTMLInputElement).value)}
								class="block w-full rounded-lg border border-slate-200 bg-white pl-10 pr-10 py-2.5 text-sm text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-slate-50 disabled:text-slate-500 transition-colors"
							/>
							<button
								type="button"
								class="absolute inset-y-0 right-0 flex items-center pr-3 text-slate-400 hover:text-slate-600 transition-colors"
								aria-label={showPassword ? 'Sembunyikan password' : 'Tampilkan password'}
								onclick={() => (showPassword = !showPassword)}
							>
								{#if showPassword}
									<EyeOff class="h-4 w-4" />
								{:else}
									<Eye class="h-4 w-4" />
								{/if}
							</button>
						</div>
					</div>

					<!-- Error alert -->
					{#if error}
						<div class="bg-red-50 border border-red-200 rounded-lg p-3 flex items-center gap-2 text-red-700">
							<AlertCircle class="h-4 w-4 shrink-0" />
							<span class="text-sm">{error}</span>
						</div>
					{/if}

					<!-- Submit button -->
					<button
						type="submit"
						disabled={loading}
						class="btn-primary w-full py-2.5 text-sm font-semibold mt-2"
					>
						{#if loading}
							<svg
								class="h-4 w-4 animate-spin"
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
								aria-hidden="true"
							>
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
							</svg>
							Memproses...
						{:else}
							Masuk
						{/if}
					</button>
				</form>

				<p class="mt-6 text-center text-xs text-slate-500">
					Lupa password? Hubungi admin HR perusahaan Anda.
				</p>
			</div>
		</div>
	</div>
</div>
