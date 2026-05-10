<script lang="ts">
  import { Search, X } from 'lucide-svelte';

  interface Props {
    value: string;
    placeholder?: string;
    onSearch?: (v: string) => void;
  }

  let {
    value = $bindable(),
    placeholder = 'Cari...',
    onSearch,
  }: Props = $props();

  function handleInput(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    value = input.value;
    onSearch?.(value);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      onSearch?.(value);
    }
  }

  function clearSearch() {
    value = '';
    onSearch?.('');
  }
</script>

<div class="relative w-full">
  <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
    <Search class="w-4 h-4 text-slate-400" />
  </div>

  <input
    type="search"
    {placeholder}
    {value}
    oninput={handleInput}
    onkeydown={handleKeydown}
    class="w-full rounded-lg border border-slate-200 bg-white py-2 pl-10 pr-9 text-sm text-slate-800
           placeholder:text-slate-400
           focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
           transition-shadow duration-150"
  />

  {#if value}
    <button
      type="button"
      onclick={clearSearch}
      aria-label="Hapus pencarian"
      class="absolute inset-y-0 right-0 flex items-center pr-3 text-slate-400 hover:text-slate-600 transition-colors"
    >
      <X class="w-4 h-4" />
    </button>
  {/if}
</div>
