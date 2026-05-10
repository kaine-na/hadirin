<script lang="ts">
	import { Camera, X, RefreshCw, CheckCircle, AlertCircle } from 'lucide-svelte';

	interface Props {
		onCapture: (file: File) => void;
		onCancel: () => void;
	}

	let { onCapture, onCancel }: Props = $props();

	let videoEl = $state<HTMLVideoElement | null>(null);
	let canvasEl = $state<HTMLCanvasElement | null>(null);
	let stream = $state<MediaStream | null>(null);
	let capturedBlob = $state<Blob | null>(null);
	let capturedURL = $state<string | null>(null);
	let error = $state<string | null>(null);
	let loading = $state(true);

	async function startCamera() {
		loading = true;
		error = null;
		try {
			stream = await navigator.mediaDevices.getUserMedia({
				video: { facingMode: 'user', width: { ideal: 640 }, height: { ideal: 480 } }
			});
			if (videoEl) {
				videoEl.srcObject = stream;
				await videoEl.play();
			}
		} catch (e) {
			error = 'Tidak dapat mengakses kamera. Pastikan izin kamera sudah diberikan.';
		} finally {
			loading = false;
		}
	}

	function stopCamera() {
		stream?.getTracks().forEach((t) => t.stop());
		stream = null;
	}

	function capture() {
		if (!videoEl || !canvasEl) return;

		canvasEl.width = videoEl.videoWidth;
		canvasEl.height = videoEl.videoHeight;
		const ctx = canvasEl.getContext('2d');
		if (!ctx) return;

		// Mirror flip untuk selfie yang natural
		ctx.translate(canvasEl.width, 0);
		ctx.scale(-1, 1);
		ctx.drawImage(videoEl, 0, 0);

		canvasEl.toBlob(
			(blob) => {
				if (!blob) return;
				capturedBlob = blob;
				capturedURL = URL.createObjectURL(blob);
				stopCamera();
			},
			'image/jpeg',
			0.85
		);
	}

	function retake() {
		if (capturedURL) {
			URL.revokeObjectURL(capturedURL);
		}
		capturedBlob = null;
		capturedURL = null;
		startCamera();
	}

	function confirm() {
		if (!capturedBlob) return;
		const file = new File([capturedBlob], `selfie_${Date.now()}.jpg`, { type: 'image/jpeg' });
		onCapture(file);
		stopCamera();
	}

	function cancel() {
		stopCamera();
		if (capturedURL) URL.revokeObjectURL(capturedURL);
		onCancel();
	}

	$effect(() => {
		startCamera();
		return () => {
			stopCamera();
			if (capturedURL) URL.revokeObjectURL(capturedURL);
		};
	});
</script>

<!-- Overlay fullscreen di mobile, centered di desktop -->
<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4"
	role="dialog"
	aria-modal="true"
	aria-label="Ambil foto selfie"
>
	<div class="relative w-full max-w-md overflow-hidden rounded-2xl bg-slate-900 shadow-2xl">
		<!-- Header -->
		<div class="flex items-center justify-between px-4 py-3">
			<div class="flex items-center gap-2 text-white">
				<Camera size={18} strokeWidth={2} />
				<span class="text-sm font-semibold">Foto Selfie</span>
			</div>
			<button
				type="button"
				onclick={cancel}
				class="rounded-full p-1.5 text-slate-400 transition-colors hover:bg-slate-700 hover:text-white"
				aria-label="Tutup kamera"
			>
				<X size={18} strokeWidth={2} />
			</button>
		</div>

		<!-- Kamera / Preview -->
		<div class="relative aspect-[4/3] w-full bg-black">
			{#if error}
				<div class="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
					<AlertCircle size={40} strokeWidth={1.5} class="text-red-400" />
					<p class="text-sm text-slate-300">{error}</p>
				</div>
			{:else if loading}
				<div class="flex h-full items-center justify-center">
					<RefreshCw size={32} strokeWidth={1.5} class="animate-spin text-slate-400" />
				</div>
			{:else if capturedURL}
				<!-- Preview foto yang sudah diambil -->
				<img
					src={capturedURL}
					alt="Foto selfie"
					class="h-full w-full object-cover"
					style="transform: scaleX(-1);"
				/>
				<div class="absolute inset-0 flex items-end justify-center pb-4">
					<div class="flex items-center gap-2 rounded-full bg-emerald-600/90 px-3 py-1.5 text-xs font-medium text-white">
						<CheckCircle size={14} strokeWidth={2} />
						Foto siap digunakan
					</div>
				</div>
			{:else}
				<!-- Live camera feed -->
				<!-- svelte-ignore a11y_media_has_caption -->
				<video
					bind:this={videoEl}
					class="h-full w-full object-cover"
					style="transform: scaleX(-1);"
					playsinline
					muted
				></video>
				<!-- Panduan wajah -->
				<div class="pointer-events-none absolute inset-0 flex items-center justify-center">
					<div class="h-48 w-40 rounded-full border-2 border-white/40 border-dashed"></div>
				</div>
			{/if}
		</div>

		<!-- Canvas tersembunyi untuk capture -->
		<canvas bind:this={canvasEl} class="hidden"></canvas>

		<!-- Instruksi -->
		{#if !capturedURL && !error && !loading}
			<p class="px-4 py-2 text-center text-xs text-slate-400">
				Posisikan wajah Anda di dalam lingkaran, lalu tekan tombol kamera
			</p>
		{/if}

		<!-- Tombol aksi -->
		<div class="flex items-center justify-center gap-4 px-4 py-4">
			{#if capturedURL}
				<button
					type="button"
					onclick={retake}
					class="flex items-center gap-2 rounded-xl border border-slate-600 px-5 py-2.5 text-sm font-medium text-slate-300 transition-colors hover:bg-slate-700"
				>
					<RefreshCw size={16} strokeWidth={2} />
					Ulangi
				</button>
				<button
					type="button"
					onclick={confirm}
					class="flex items-center gap-2 rounded-xl bg-emerald-600 px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-emerald-700"
				>
					<CheckCircle size={16} strokeWidth={2} />
					Gunakan Foto
				</button>
			{:else if !error && !loading}
				<button
					type="button"
					onclick={capture}
					class="flex h-14 w-14 items-center justify-center rounded-full bg-white shadow-lg transition-transform active:scale-95 hover:bg-slate-100"
					aria-label="Ambil foto"
				>
					<Camera size={24} strokeWidth={2} class="text-slate-800" />
				</button>
			{/if}
		</div>
	</div>
</div>
