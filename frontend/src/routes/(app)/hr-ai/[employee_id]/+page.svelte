<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { aiApi, employeesApi } from '$lib/api';
	import { toast } from '$lib/stores/toast.svelte';
	import { isHRorManager } from '$lib/stores/auth.svelte';
	import Button from '$lib/components/Button.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import { formatDate, formatDateTime, roleLabel, startOfMonthISO, todayISO } from '$lib/utils/format';
	import type { AIReport, Employee } from '$lib/types';

	const employeeId = $derived($page.params.employee_id);

	let employee = $state<Employee | null>(null);
	let reports = $state<AIReport[]>([]);
	let loading = $state(true);
	let generating = $state(false);

	let periodStart = $state(startOfMonthISO());
	let periodEnd = $state(todayISO());
	let customPrompt = $state('');

	let selectedReport = $state<AIReport | null>(null);

	async function loadAll() {
		loading = true;
		try {
			const [emp, reps] = await Promise.all([
				employeesApi.get(employeeId),
				aiApi.listReports(employeeId)
			]);
			employee = emp;
			reports = reps ?? [];
			selectedReport = reports[0] ?? null;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal memuat data');
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		if (!isHRorManager()) {
			goto('/dashboard');
			return;
		}
		loadAll();
	});

	async function generate() {
		generating = true;
		try {
			const report = await aiApi.analyze(employeeId, {
				period_start: periodStart,
				period_end: periodEnd,
				custom_prompt: customPrompt || undefined
			});
			toast.success('Laporan AI berhasil dibuat');
			reports = [report, ...reports];
			selectedReport = report;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal generate laporan');
		} finally {
			generating = false;
		}
	}
</script>

<svelte:head>
	<title>Laporan AI — {employee?.name ?? 'Karyawan'}</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<a href="/hr-ai" class="text-sm text-primary-700 hover:underline">&larr; Kembali ke daftar</a>
	</div>

	{#if loading}
		<div class="card p-10 text-center text-sm text-slate-500">Memuat...</div>
	{:else if !employee}
		<div class="card p-10 text-center text-sm text-slate-500">Karyawan tidak ditemukan.</div>
	{:else}
		<!-- Employee header -->
		<div class="card p-6">
			<div class="flex items-center gap-4">
				{#if employee.photo_url}
					<img
						src={employee.photo_url}
						alt={employee.name}
						class="h-16 w-16 rounded-full object-cover"
					/>
				{:else}
					<div
						class="flex h-16 w-16 items-center justify-center rounded-full bg-primary-100 text-xl font-bold text-primary-700"
					>
						{employee.name.charAt(0).toUpperCase()}
					</div>
				{/if}
				<div>
					<h1 class="text-2xl font-bold text-slate-900">{employee.name}</h1>
					<p class="text-sm text-slate-500">
						{employee.position || '-'}{employee.department ? ` · ${employee.department}` : ''}
					</p>
					<div class="mt-2 flex flex-wrap items-center gap-2">
						<Badge color="primary">{roleLabel(employee.role)}</Badge>
						{#if employee.joined_at}
							<Badge color="slate">Bergabung {formatDate(employee.joined_at)}</Badge>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<!-- Generate form -->
		<div class="card p-6">
			<h2 class="text-lg font-semibold text-slate-900">Generate Laporan Kinerja AI</h2>
			<p class="mt-1 text-sm text-slate-500">
				Sistem akan menganalisis absensi dan dokumen karyawan pada periode yang dipilih.
			</p>

			<div class="mt-5 grid gap-4 md:grid-cols-2">
				<div>
					<label for="ps" class="label">Periode Mulai</label>
					<input id="ps" type="date" class="input" bind:value={periodStart} />
				</div>
				<div>
					<label for="pe" class="label">Periode Selesai</label>
					<input id="pe" type="date" class="input" bind:value={periodEnd} />
				</div>
			</div>

			<div class="mt-4">
				<label for="cp" class="label">
					Prompt Tambahan <span class="text-xs font-normal text-slate-400">(opsional)</span>
				</label>
				<textarea
					id="cp"
					rows="3"
					class="input"
					placeholder="Contoh: fokus pada konsistensi kehadiran di bulan ini"
					bind:value={customPrompt}
				></textarea>
			</div>

			<div class="mt-5 flex justify-end">
				<Button onclick={generate} loading={generating}>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
					</svg>
					Generate Laporan
				</Button>
			</div>
		</div>

		<!-- Reports -->
		{#if reports.length === 0}
			<div class="card p-10 text-center text-sm text-slate-500">
				Belum ada laporan AI untuk karyawan ini.
			</div>
		{:else}
			<div class="grid gap-4 lg:grid-cols-3">
				<!-- History -->
				<div class="card p-4 lg:col-span-1">
					<h3 class="mb-3 text-sm font-semibold text-slate-900">Riwayat Laporan</h3>
					<div class="space-y-2">
						{#each reports as r}
							<button
								type="button"
								class="w-full rounded-lg border p-3 text-left transition-colors {selectedReport?.id ===
								r.id
									? 'border-primary-500 bg-primary-50'
									: 'border-slate-200 hover:bg-slate-50'}"
								onclick={() => (selectedReport = r)}
							>
								<p class="text-sm font-medium text-slate-900">
									{formatDate(r.period_start)} — {formatDate(r.period_end)}
								</p>
								<p class="mt-1 text-xs text-slate-500">
									Dibuat {formatDateTime(r.created_at)}
								</p>
								{#if r.model_used}
									<p class="mt-1 text-xs text-slate-400">Model: {r.model_used}</p>
								{/if}
							</button>
						{/each}
					</div>
				</div>

				<!-- Detail -->
				<div class="card p-6 lg:col-span-2">
					{#if selectedReport}
						<div class="mb-4 flex items-center justify-between border-b border-slate-200 pb-4">
							<div>
								<h3 class="text-lg font-semibold text-slate-900">Laporan Kinerja</h3>
								<p class="text-sm text-slate-500">
									Periode: {formatDate(selectedReport.period_start)} — {formatDate(
										selectedReport.period_end
									)}
								</p>
							</div>
							<Badge color="primary">AI Report</Badge>
						</div>
						<article class="prose prose-sm max-w-none whitespace-pre-wrap text-sm leading-relaxed text-slate-800">
							{selectedReport.response}
						</article>
					{:else}
						<p class="text-sm text-slate-500">Pilih laporan dari riwayat di sebelah kiri.</p>
					{/if}
				</div>
			</div>
		{/if}
	{/if}
</div>
