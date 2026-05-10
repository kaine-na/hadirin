<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { DepartmentStat } from '$lib/api/analytics';

	interface Props {
		data: DepartmentStat[];
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

		const labels = data.map((d) => d.department);

		chart = new Chart(canvas, {
			type: 'bar',
			data: {
				labels,
				datasets: [
					{
						label: 'Hadir',
						data: data.map((d) => d.total_present),
						backgroundColor: 'rgba(37, 99, 235, 0.85)',
						borderRadius: 4,
						borderSkipped: false
					},
					{
						label: 'Terlambat',
						data: data.map((d) => d.total_late),
						backgroundColor: 'rgba(245, 158, 11, 0.85)',
						borderRadius: 4,
						borderSkipped: false
					},
					{
						label: 'Alpha',
						data: data.map((d) => d.total_absent),
						backgroundColor: 'rgba(239, 68, 68, 0.85)',
						borderRadius: 4,
						borderSkipped: false
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
						callbacks: {
							afterBody: (items) => {
								const idx = items[0]?.dataIndex;
								if (idx !== undefined && data[idx]) {
									return [`Tingkat kehadiran: ${data[idx].attendance_rate.toFixed(1)}%`];
								}
								return [];
							}
						}
					}
				},
				scales: {
					x: {
						grid: { display: false },
						ticks: { font: { size: 11 } }
					},
					y: {
						beginAtZero: true,
						grid: { color: 'rgba(226, 232, 240, 0.8)' },
						ticks: { font: { size: 11 }, stepSize: 1 }
					}
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
	<h3 class="mb-1 text-sm font-semibold text-slate-900">Kehadiran per Departemen</h3>
	<p class="mb-4 text-xs text-slate-500">Perbandingan kehadiran antar departemen</p>
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
