<script lang="ts">
	import { onMount } from 'svelte';
	import { attendanceApi } from '$lib/api';
	import { fraudApi } from '$lib/api/fraud';
	import { toast } from '$lib/stores/toast.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Table from '$lib/components/Table.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import SkeletonLoader from '$lib/components/ui/SkeletonLoader.svelte';
	import SelfieCapture from '$lib/components/SelfieCapture.svelte';
	import GPSIndicator from '$lib/components/GPSIndicator.svelte';
	import type { Column } from '$lib/components/types';
	import {
		formatDate,
		formatTime,
		statusLabel,
		statusColor,
		startOfMonthISO,
		todayISO
	} from '$lib/utils/format';
	import type { Attendance } from '$lib/types';
	import { LogIn, LogOut, Clock, CheckCircle, AlertCircle, RefreshCw, Camera } from 'lucide-svelte';

	let today = $state<Attendance | null>(null);
	let loadingToday = $state(true);
	let clocking = $state(false);

	let records = $state<Attendance[]>([]);
	let loadingRecords = $state(false);
	let startDate = $state(startOfMonthISO());
	let endDate = $state(todayISO());

	// State untuk selfie dan GPS
	let showSelfieCapture = $state(false);
	let selfieFile = $state<File | null>(null);
	let gpsData = $state<{ latitude: number; longitude: number; accuracy: number } | null>(null);
	let gpsLoading = $state(false);
	let pendingClockIn = $state(false);

	// Real-time clock
	let currentTime = $state(new Date());
	let clockInterval: ReturnType<typeof setInterval>;

	const columns: Column<Attendance>[] = [
		{ key: 'date', label: 'Tanggal' },
		{ key: 'clock_in', label: 'Clock In' },
		{ key: 'clock_out', label: 'Clock Out' },
		{ key: 'status', label: 'Status' },
		{ key: 'notes', label: 'Keterangan' }
	];

	async function loadToday() {
		loadingToday = true;
		try {
			today = await attendanceApi.today();
		} catch {
			today = null;
		} finally {
			loadingToday = false;
		}
	}

	async function loadRecords() {
		loadingRecords = true;
		try {
			const res = await attendanceApi.me({
				start_date: startDate,
				end_date: endDate,
				page: 1,
				page_size: 100
			});
			records = res.items;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal memuat rekap');
		} finally {
			loadingRecords = false;
		}
	}

	// Ambil GPS location
	async function getGPS(): Promise<void> {
		if (!navigator.geolocation) {
			toast.error('GPS tidak didukung di perangkat ini');
			return;
		}

		gpsLoading = true;
		return new Promise((resolve) => {
			navigator.geolocation.getCurrentPosition(
				(pos) => {
					gpsData = {
						latitude: pos.coords.latitude,
						longitude: pos.coords.longitude,
						accuracy: pos.coords.accuracy
					};
					gpsLoading = false;
					resolve();
				},
				(err) => {
					gpsLoading = false;
					if (err.code === err.PERMISSION_DENIED) {
						toast.error('Izin GPS ditolak. Aktifkan lokasi untuk clock-in.');
					}
					resolve();
				},
				{ enableHighAccuracy: true, timeout: 10000, maximumAge: 0 }
			);
		});
	}

	// Mulai proses clock-in: ambil GPS dulu, lalu minta selfie
	async function startClockIn() {
		pendingClockIn = true;
		await getGPS();
		showSelfieCapture = true;
	}

	// Selfie sudah diambil, lanjut clock-in
	async function onSelfieCapture(file: File) {
		selfieFile = file;
		showSelfieCapture = false;
		await doClockIn();
	}

	// Batalkan selfie
	function onSelfieCancel() {
		showSelfieCapture = false;
		pendingClockIn = false;
		selfieFile = null;
	}

	// Eksekusi clock-in dengan fraud validation
	async function doClockIn() {
		clocking = true;
		try {
			// 1. Clock-in ke attendance API
			const att = await attendanceApi.clockIn();
			toast.success('Berhasil clock in! Selamat bekerja 💪');

			// 2. Jalankan fraud validation di background (non-blocking)
			if (att?.id && (gpsData || selfieFile)) {
				try {
					const fraudResult = await fraudApi.validateClockIn(
						att.id,
						gpsData ?? { latitude: 0, longitude: 0, accuracy: 0 },
						selfieFile ?? undefined
					);

					if (fraudResult.fraud_detected) {
						toast.error(
							`Peringatan: ${fraudResult.fraud_count} anomali terdeteksi. HR akan meninjau absensi Anda.`
						);
					}
				} catch {
					// Fraud validation gagal — tidak blokir clock-in
				}
			}

			await Promise.all([loadToday(), loadRecords()]);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal clock in');
		} finally {
			clocking = false;
			pendingClockIn = false;
			selfieFile = null;
		}
	}

	async function clockOut() {
		clocking = true;
		try {
			await attendanceApi.clockOut();
			toast.success('Berhasil clock out! Selamat beristirahat 🏠');
			await Promise.all([loadToday(), loadRecords()]);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal clock out');
		} finally {
			clocking = false;
		}
	}

	onMount(() => {
		loadToday();
		loadRecords();
		// Ambil GPS saat halaman dimuat
		getGPS();
		// Update jam setiap detik
		clockInterval = setInterval(() => {
			currentTime = new Date();
		}, 1000);
		return () => clearInterval(clockInterval);
	});

	const canClockIn = $derived(!today?.clock_in);
	const canClockOut = $derived(!!today?.clock_in && !today?.clock_out);
	const gpsBlocked = $derived(gpsData !== null && gpsData.accuracy > 100);

	const timeStr = $derived(
		currentTime.toLocaleTimeString('id-ID', {
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		})
	);

	const dateStr = $derived(
		currentTime.toLocaleDateString('id-ID', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		})
	);
</script>

<svelte:head>
	<title>Absensi — Hadir</title>
</svelte:head>

<!-- Selfie Capture Modal -->
{#if showSelfieCapture}
	<SelfieCapture onCapture={onSelfieCapture} onCancel={onSelfieCancel} />
{/if}

<div class="space-y-6">
	<PageHeader
		title="Absensi"
		description="Catat kehadiran harian Anda dan lihat rekap absensi pribadi."
	/>

	<!-- Clock Panel -->
	<div
		class="overflow-hidden rounded-2xl border border-slate-100 bg-white shadow-md transition-all duration-200 hover:shadow-lg"
	>
		<!-- Jam Real-time -->
		<div class="bg-gradient-to-r from-[#1e3a5f] to-[#2d5a8e] px-6 py-5 text-white">
			<div class="flex flex-wrap items-center justify-between gap-4">
				<div>
					<p class="text-4xl font-bold tabular-nums tracking-tight">{timeStr}</p>
					<p class="mt-1 text-sm text-blue-200">{dateStr}</p>
				</div>
				{#if loadingToday}
					<div class="text-blue-200 text-sm">Memuat status...</div>
				{:else if today}
					<div class="text-right">
						<Badge color={statusColor(today.status)}>{statusLabel(today.status)}</Badge>
						<div class="mt-2 flex gap-4 text-sm text-blue-100">
							<div>
								<p class="text-xs text-blue-300">Masuk</p>
								<p class="font-semibold">{formatTime(today.clock_in)}</p>
							</div>
							{#if today.clock_out}
								<div>
									<p class="text-xs text-blue-300">Pulang</p>
									<p class="font-semibold">{formatTime(today.clock_out)}</p>
								</div>
							{/if}
						</div>
					</div>
				{:else}
					<div class="flex items-center gap-2 text-amber-300">
						<AlertCircle size={16} strokeWidth={2} />
						<span class="text-sm">Belum absen hari ini</span>
					</div>
				{/if}
			</div>
		</div>

		<!-- GPS Indicator + Tombol Clock In/Out -->
		<div class="p-6 space-y-4">
			<!-- GPS Status -->
			<GPSIndicator
				accuracy={gpsData?.accuracy ?? null}
				latitude={gpsData?.latitude}
				longitude={gpsData?.longitude}
				loading={gpsLoading}
			/>

			<!-- Tombol Clock In / Out -->
			<div class="flex flex-wrap items-center gap-4">
				<button
					type="button"
					onclick={startClockIn}
					disabled={!canClockIn || clocking || pendingClockIn || gpsBlocked}
					class="inline-flex items-center gap-2 rounded-xl px-6 py-3 text-sm font-semibold shadow-sm transition-all duration-150 active:scale-95
						{canClockIn && !clocking && !pendingClockIn && !gpsBlocked
						? 'animate-pulse-subtle bg-emerald-600 text-white hover:bg-emerald-700 hover:shadow-md'
						: 'cursor-not-allowed bg-slate-100 text-slate-400'}"
				>
					{#if (clocking || pendingClockIn) && canClockIn}
						<RefreshCw size={16} strokeWidth={2} class="animate-spin" />
						Memproses...
					{:else}
						<Camera size={16} strokeWidth={2} />
						Clock In + Selfie
					{/if}
				</button>

				<button
					type="button"
					onclick={clockOut}
					disabled={!canClockOut || clocking}
					class="inline-flex items-center gap-2 rounded-xl px-6 py-3 text-sm font-semibold shadow-sm transition-all duration-150 active:scale-95
						{canClockOut && !clocking
						? 'bg-red-600 text-white hover:bg-red-700 hover:shadow-md'
						: 'cursor-not-allowed bg-slate-100 text-slate-400'}"
				>
					{#if clocking && canClockOut}
						<RefreshCw size={16} strokeWidth={2} class="animate-spin" />
						Memproses...
					{:else}
						<LogOut size={16} strokeWidth={2} />
						Clock Out
					{/if}
				</button>

				{#if today?.clock_out}
					<div class="flex items-center gap-2 text-sm text-emerald-600">
						<CheckCircle size={16} strokeWidth={2} />
						<span>Absensi hari ini sudah lengkap</span>
					</div>
				{/if}
			</div>

			<!-- Info selfie -->
			{#if canClockIn && !today?.clock_in}
				<p class="text-xs text-slate-400 flex items-center gap-1.5">
					<Camera size={12} strokeWidth={2} />
					Clock-in memerlukan foto selfie untuk verifikasi kehadiran
				</p>
			{/if}
		</div>
	</div>

	<!-- Rekap Absensi -->
	<div class="rounded-xl border border-slate-100 bg-white p-6 shadow-sm">
		<div class="mb-4 flex flex-wrap items-end justify-between gap-4">
			<div>
				<h2 class="text-lg font-semibold text-slate-900">Rekap Absensi</h2>
				<p class="text-sm text-slate-500">Catatan kehadiranmu dalam rentang tanggal tertentu.</p>
			</div>
			<div class="flex flex-wrap items-end gap-3">
				<div>
					<label for="start" class="label">Dari</label>
					<input id="start" type="date" class="input" bind:value={startDate} />
				</div>
				<div>
					<label for="end" class="label">Sampai</label>
					<input id="end" type="date" class="input" bind:value={endDate} />
				</div>
				<button
					type="button"
					onclick={loadRecords}
					disabled={loadingRecords}
					class="inline-flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm transition-all duration-150 hover:bg-slate-50 hover:shadow-md active:scale-95 disabled:opacity-50"
				>
					{#if loadingRecords}
						<RefreshCw size={14} strokeWidth={2} class="animate-spin" />
					{:else}
						<RefreshCw size={14} strokeWidth={2} />
					{/if}
					Terapkan
				</button>
			</div>
		</div>

		<Table {columns} rows={records} loading={loadingRecords} emptyMessage="Belum ada catatan absensi">
			{#snippet row(r: Attendance)}
				<tr class="transition-colors hover:bg-slate-50">
					<td class="px-4 py-3 text-sm text-slate-700">{formatDate(r.date)}</td>
					<td class="px-4 py-3 text-sm font-medium text-slate-900">{formatTime(r.clock_in)}</td>
					<td class="px-4 py-3 text-sm font-medium text-slate-900">{formatTime(r.clock_out)}</td>
					<td class="px-4 py-3 text-sm">
						<Badge color={statusColor(r.status)}>{statusLabel(r.status)}</Badge>
					</td>
					<td class="px-4 py-3 text-sm text-slate-500">{r.notes || '-'}</td>
				</tr>
			{/snippet}
		</Table>
	</div>
</div>
