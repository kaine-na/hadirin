<script lang="ts">
	interface Props {
		open: boolean;
		title?: string;
		size?: 'sm' | 'md' | 'lg' | 'xl';
		onclose?: () => void;
		children?: import('svelte').Snippet;
		footer?: import('svelte').Snippet;
	}

	let { open = $bindable(false), title, size = 'md', onclose, children, footer }: Props = $props();

	const sizeClass: Record<NonNullable<Props['size']>, string> = {
		sm: 'max-w-md',
		md: 'max-w-lg',
		lg: 'max-w-2xl',
		xl: 'max-w-4xl'
	};

	function close() {
		open = false;
		onclose?.();
	}

	function onKeyDown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open) close();
	}
</script>

<svelte:window onkeydown={onKeyDown} />

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4">
		<!-- Backdrop -->
		<button
			type="button"
			class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm"
			aria-label="Tutup modal"
			onclick={close}
		></button>

		<!-- Dialog -->
		<div
			class="relative z-10 w-full {sizeClass[size]} rounded-xl bg-white shadow-xl"
			role="dialog"
			aria-modal="true"
			aria-labelledby={title ? 'modal-title' : undefined}
		>
			{#if title}
				<div class="flex items-center justify-between border-b border-slate-200 px-5 py-4">
					<h2 id="modal-title" class="text-lg font-semibold text-slate-900">{title}</h2>
					<button
						type="button"
						class="rounded-lg p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
						aria-label="Tutup"
						onclick={close}
					>
						<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</button>
				</div>
			{/if}

			<div class="px-5 py-4">
				{@render children?.()}
			</div>

			{#if footer}
				<div class="flex items-center justify-end gap-2 border-t border-slate-200 px-5 py-3">
					{@render footer()}
				</div>
			{/if}
		</div>
	</div>
{/if}
