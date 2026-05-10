<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { AttendanceSummary } from '$lib/api/analytics';

	interface Props {
		data: AttendanceSummary | null;
	}

	let { data }: Props = $props();

	let canvas: HTMLCanvasElement = $state() as HTMLCanvasElement;
	let chart: import('chart.js').Chart | null = null;

	async function initChart() {
		if (!data) return;

		const { Chart, registerables } = await import('chart.js');
		Chart.register(...registerables);

		if (chart) {
			chart.destroy();
		}

		const total = data.total_present + data.total_late + data.total_absent + data.total_leave + data.total_sick;

		chart = new Chart(canvas, {
			type: 'doughnut',
			data: {
				labels: ['Hadir', 'Terlambat', 'Alpha', 'Izin', 'Sakit'],
				datasets: [
					{
						data: [
							data.total_present,
							data.total_late,
							data.total_absent,
							data.total_leave,
							data.total_sick
						],
						backgroundColor: [
							'rgba(37, 99, 235, 0.85)',
							'rgba(245, 158, 11, 0.85)',
							'rgba(239, 68, 68, 0.85)',
							'rgba(99, 102, 241, 0.85)',
							'rgba(20, 184, 166, 0.85)'
						],
						borderWidth: 2,
						borderColor: '#ffffff',
						hoverOffset: 4
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				cutout: '65%',
				plugins: {
					legend: {
						position: 'right',
						labels: {
							font: { size: 12 },
							usePointStyle: true,
							padding: 12,
							generateLabels: (chart) => {
								const ds = chart.data.datasets[0];
								return (chart.data.labels as string[]).map((label, i) => ({
									text: `${label} (${ds.data[i]})`,
									fillStyle: (ds.backgroundColor as string[])[i],
									strokeStyle: '#fff',
									lineWidth: 2,
									hidden: false,
									index: i
								}));
							}
						}
					},
					tooltip: {
						callbacks: {
							label: (ctx) => {
								const pct = total > 0 ? ((ctx.parsed / total) * 100).toFixed(1) : '0';
								return ` ${ctx.label}: ${ctx.parsed} hari (${pct}%)`;
							}
						}
					}
				}
			}
		});
	}

	onMount(() => {
		if (data) {
			initChart();
		}
	});

	onDestroy(() => {
		chart?.destroy();
	});

	$effect(() => {
		if (data && canvas) {
			initChart();
		}
	});
</script>

<div class="rounded-xl bg-white p-6 shadow-md">
	<h3 class="mb-1 text-sm font-semibold text-slate-900">Distribusi Status Kehadiran</h3>
	<p class="mb-4 text-xs text-slate-500">Proporsi status kehadiran seluruh karyawan</p>
	{#if !data || data.total_working_days === 0}
		<div class="flex h-48 items-center justify-center text-sm text-slate-400">
			Belum ada data untuk periode ini
		</div>
	{:else}
		<div class="h-56">
			<canvas bind:this={canvas}></canvas>
		</div>
	{/if}
</div>
