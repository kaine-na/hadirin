<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { employeesApi } from '$lib/api';
	import { toast } from '$lib/stores/toast.svelte';
	import { isHRorManager } from '$lib/stores/auth.svelte';
	import { roleLabel } from '$lib/utils/format';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import SkeletonLoader from '$lib/components/ui/SkeletonLoader.svelte';
	import type { Employee } from '$lib/types';
	import { Brain, Sparkles, Search, ChevronRight, TrendingUp, Users } from 'lucide-svelte';

	let employees = $state<Employee[]>([]);
	let loading = $state(true);
	let search = $state('');

	onMount(async () => {
		if (!isHRorManager()) {
			goto('/dashboard');
			return;
		}
		try {
			const res = await employeesApi.list({ page: 1, page_size: 500 });
			employees = res.items;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal memuat karyawan');
		} finally {
			loading = false;
		}
	});

	const filtered = $derived(
		employees.filter((e) => {
			if (!search) return true;
			const q = search.toLowerCase();
			return (
				e.name.toLowerCase().includes(q) ||
				e.email.toLowerCase().includes(q) ||
				(e.department ?? '').toLowerCase().includes(q)
			);
		})
	);
</script>

<svelte:head>
	<title>HR AI — SaaS Karyawan</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<PageHeader
			title="HR AI Dashboard"
			description="Analisis kinerja karyawan menggunakan kecerdasan buatan. Pilih karyawan untuk memulai."
		/>
		<div
			class="flex items-center gap-2 rounded-xl border border-violet-100 bg-gradient-to-r from-violet-50 to-purple-50 px-4 py-2"
		>
			<Sparkles size={16} strokeWidth={1.75} class="text-violet-600" />
			<span class="text-sm font-medium text-violet-700">Powered by AI</span>
		</div>
	</div>

	<!-- Info Banner -->
	<div
		class="rounded-xl border border-blue-100 bg-gradient-to-r from-blue-50 to-indigo-50 p-5 shadow-sm"
	>
		<div class="flex items-start gap-4">
			<div
				class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-blue-600 text-white shadow-sm"
			>
				<Brain size={20} strokeWidth={1.75} />
			</div>
			<div>
				<p class="font-semibold text-slate-900">Cara Kerja HR AI</p>
				<p class="mt-1 text-sm text-slate-600">
					Pilih karyawan di bawah, lalu AI akan menganalisis data absensi dan dokumen kerja mereka
					untuk menghasilkan laporan kinerja komprehensif dengan rekomendasi actionable.
				</p>
				<div class="mt-3 flex flex-wrap gap-4">
					<div class="flex items-center gap-1.5 text-xs text-slate-500">
						<TrendingUp size={12} strokeWidth={2} class="text-blue-500" />
						Analisis tren kehadiran
					</div>
					<div class="flex items-center gap-1.5 text-xs text-slate-500">
						<Sparkles size={12} strokeWidth={2} class="text-violet-500" />
						Skor kinerja otomatis
					</div>
					<div class="flex items-center gap-1.5 text-xs text-slate-500">
						<Brain size={12} strokeWidth={2} class="text-indigo-500" />
						Rekomendasi pengembangan
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Search -->
	<div class="relative">
		<Search
			size={16}
			strokeWidth={1.75}
			class="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400"
		/>
		<input
			type="text"
			placeholder="Cari karyawan berdasarkan nama, email, atau departemen..."
			class="w-full rounded-xl border border-slate-200 bg-white py-3 pl-10 pr-4 text-sm text-slate-900 placeholder-slate-400 shadow-sm transition-all focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
			bind:value={search}
		/>
	</div>

	<!-- Employee Grid -->
	{#if loading}
		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array(6) as _}
				<SkeletonLoader rows={3} />
			{/each}
		</div>
	{:else if filtered.length === 0}
		<EmptyState
			icon={Users}
			title="Tidak ada karyawan"
			description={search ? 'Tidak ada karyawan yang cocok dengan pencarian.' : 'Belum ada karyawan terdaftar.'}
			ctaLabel={search ? undefined : 'Tambah Karyawan'}
			ctaHref={search ? undefined : '/employees'}
		/>
	{:else}
		<div>
			<p class="mb-3 text-xs font-medium uppercase tracking-wider text-slate-400">
				{filtered.length} karyawan · Klik untuk analisis
			</p>
			<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
				{#each filtered as emp}
					<a
						href={`/hr-ai/${emp.id}`}
						class="group flex items-center gap-3 rounded-xl border border-slate-100 bg-white p-4 shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-violet-200 hover:shadow-md"
					>
						{#if emp.photo_url}
							<img
								src={emp.photo_url}
								alt={emp.name}
								class="h-12 w-12 shrink-0 rounded-full object-cover ring-2 ring-white"
							/>
						{:else}
							<div
								class="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-violet-100 to-purple-100 text-base font-semibold text-violet-700 ring-2 ring-white"
							>
								{emp.name.charAt(0).toUpperCase()}
							</div>
						{/if}
						<div class="min-w-0 flex-1">
							<p class="truncate text-sm font-semibold text-slate-900">{emp.name}</p>
							<p class="truncate text-xs text-slate-500">{emp.position || roleLabel(emp.role)}</p>
							{#if emp.department}
								<p class="truncate text-xs text-slate-400">{emp.department}</p>
							{/if}
						</div>
						<div
							class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-slate-50 text-slate-400 transition-all group-hover:bg-violet-100 group-hover:text-violet-600"
						>
							<ChevronRight size={16} strokeWidth={2} />
						</div>
					</a>
				{/each}
			</div>
		</div>
	{/if}
</div>
