<script lang="ts">
  type SkeletonType = 'table' | 'card' | 'text';

  interface Props {
    rows?: number;
    type?: SkeletonType;
  }

  let { rows = 3, type = 'table' }: Props = $props();

  // Varying widths for a natural look
  const tableWidths = ['w-full', 'w-5/6', 'w-4/6', 'w-3/4', 'w-full', 'w-2/3'];
  const textWidths  = ['w-full', 'w-11/12', 'w-4/5', 'w-3/4', 'w-5/6', 'w-2/3'];

  const rowArray = $derived(Array.from({ length: rows }, (_, i) => i));
</script>

{#if type === 'table'}
  <div class="animate-pulse space-y-3">
    <!-- Header row -->
    <div class="flex gap-4 pb-2 border-b border-slate-100">
      <div class="h-4 bg-slate-200 rounded w-1/4"></div>
      <div class="h-4 bg-slate-200 rounded w-1/4"></div>
      <div class="h-4 bg-slate-200 rounded w-1/6"></div>
      <div class="h-4 bg-slate-200 rounded w-1/6"></div>
    </div>
    {#each rowArray as i}
      <div class="flex gap-4 items-center py-1">
        <div class="h-4 bg-slate-100 rounded {tableWidths[i % tableWidths.length]}"></div>
        <div class="h-4 bg-slate-100 rounded w-1/5"></div>
        <div class="h-4 bg-slate-100 rounded w-1/6"></div>
        <div class="h-6 bg-slate-100 rounded-full w-16"></div>
      </div>
    {/each}
  </div>

{:else if type === 'card'}
  <div class="animate-pulse grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
    {#each rowArray as _}
      <div class="rounded-2xl border border-slate-100 bg-white p-5 space-y-3">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-slate-200 rounded-xl flex-shrink-0"></div>
          <div class="flex-1 space-y-2">
            <div class="h-3 bg-slate-200 rounded w-3/4"></div>
            <div class="h-3 bg-slate-100 rounded w-1/2"></div>
          </div>
        </div>
        <div class="h-8 bg-slate-100 rounded w-1/3"></div>
        <div class="h-3 bg-slate-100 rounded w-2/3"></div>
      </div>
    {/each}
  </div>

{:else}
  <!-- text type -->
  <div class="animate-pulse space-y-2">
    {#each rowArray as i}
      <div class="h-4 bg-slate-200 rounded {textWidths[i % textWidths.length]}"></div>
    {/each}
  </div>
{/if}
