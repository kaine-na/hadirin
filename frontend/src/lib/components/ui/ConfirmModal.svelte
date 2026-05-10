<script lang="ts">
  import { AlertTriangle } from 'lucide-svelte';

  interface Props {
    open: boolean;
    title: string;
    message: string;
    confirmLabel?: string;
    cancelLabel?: string;
    onConfirm: () => void;
    onCancel: () => void;
    danger?: boolean;
  }

  let {
    open,
    title,
    message,
    confirmLabel = 'Hapus',
    cancelLabel = 'Batal',
    onConfirm,
    onCancel,
    danger = true,
  }: Props = $props();

  function handleOverlayClick(e: MouseEvent) {
    if (e.target === e.currentTarget) onCancel();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onCancel();
  }
</script>

{#if open}
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    role="dialog"
    aria-modal="true"
    aria-labelledby="confirm-modal-title"
    class="fixed inset-0 z-50 flex items-center justify-center p-4"
    onclick={handleOverlayClick}
    onkeydown={handleKeydown}
  >
    <!-- Overlay -->
    <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" aria-hidden="true"></div>

    <!-- Modal -->
    <div class="relative z-10 w-full max-w-md rounded-2xl shadow-2xl bg-white p-6 animate-in fade-in zoom-in-95 duration-150">
      <div class="flex items-start gap-4">
        {#if danger}
          <div class="flex-shrink-0 w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
            <AlertTriangle class="w-5 h-5 text-red-600" />
          </div>
        {/if}
        <div class="flex-1 min-w-0">
          <h2 id="confirm-modal-title" class="text-base font-semibold text-slate-800">
            {title}
          </h2>
          <p class="mt-1 text-sm text-slate-500 leading-relaxed">{message}</p>
        </div>
      </div>

      <div class="mt-6 flex justify-end gap-3">
        <button
          type="button"
          onclick={onCancel}
          class="px-4 py-2 rounded-lg text-sm font-medium text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors duration-150"
        >
          {cancelLabel}
        </button>
        <button
          type="button"
          onclick={onConfirm}
          class="px-4 py-2 rounded-lg text-sm font-medium transition-colors duration-150
            {danger
              ? 'bg-red-600 text-white hover:bg-red-700'
              : 'bg-blue-600 text-white hover:bg-blue-700'}"
        >
          {confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}
