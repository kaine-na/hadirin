<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { TrendPoint } from '$lib/api/analytics';

	interface Props {
		data: TrendPoint[];
	}

	let { data }: Props = $props();

	let canvas: HTMLCanvasElement = $state() as HTMLCanvasElement;
	let chart: import('chart.js').Chart | null = null;

	async function initChart() {
		const { Chart, registerables } = await import('chart.js');
		Chart.register(...registerables);

		if (chart) {
			chart.destroy();
		}

		const labels = data.map((d) => {
			const date = new Date(d.date);
			return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short' });
		});

		chart = new Chart(canvas, {
			type: 'line',
			data: {
				labels,
				datasets: [
					{
						label: 'Hadir',
						data: data.map((d) => d.present),
						borderColor: '#2563eb',
						backgroundColor: 'rgba(37, 99, 235, 0.08)',
						borderWidth: 2,
						pointRadius: 3,
						pointHoverRadius: 5,
						fill: true,
						tension: 0.3
					},
					{
						label: 'Terlambat',
						data: data.map((d) => d.late),
						borderColor: '#f59e0b',
						backgroundColor: 'rgba(245, 158, 11, 0.08)',
						borderWidth: 2,
						pointRadius: 3,
						pointHoverRadius: 5,
						fill: true,
						tension: 0.3
					},
					{
						label: 'Alpha',
						data: data.map((d) => d.absent),
						borderColor: '#ef4444',
						backgroundColor: 'rgba(239, 68, 68, 0.08)',
						borderWidth: 2,
						pointRadius: 3,
						pointHoverRadius: 5,
						fill: true,
						tension: 0.3
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: {
						position: 'top',
						labels: {
							font: { size: 12 },
							usePointStyle: true,
							padding: 16
						}
					},
					tooltip: {
						mode: 'index',
						intersect: false
					}
				},
				scales: {
					x: {
						grid: { display: false },
						ticks: { font: { size: 11 }, maxTicksLimit: 10 }
					},
					y: {
						beginAtZero: true,
						grid: { color: 'rgba(226, 232, 240, 0.8)' },
						ticks: { font: { size: 11 }, stepSize: 1 }
					}
				},
				interaction: {
					mode: 'nearest',
					axis: 'x',
					intersect: false
				}
			}
		});
	}

	onMount(() => {
		if (data.length > 0) {
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
	<h3 class="mb-1 text-sm font-semibold text-slate-900">Tren Kehadiran 30 Hari</h3>
	<p class="mb-4 text-xs text-slate-500">Grafik kehadiran harian karyawan</p>
	{#if data.length === 0}
		<div class="flex h-48 items-center justify-center text-sm text-slate-400">
			Belum ada data untuk periode ini
		</div>
	{:else}
		<div class="h-64">
			<canvas bind:this={canvas}></canvas>
		</div>
	{/if}
</div>
