<script lang="ts">
  interface Props {
    icon?: any;
    title: string;
    description?: string;
    ctaLabel?: string;
    ctaHref?: string;
    onCta?: () => void;
  }

  let { icon: Icon, title, description, ctaLabel, ctaHref, onCta }: Props = $props();

  function handleCta() {
    if (onCta) onCta();
  }
</script>

<div class="flex flex-col items-center justify-center py-16 text-center px-4">
  {#if Icon}
    <div class="mb-4">
      <Icon class="w-16 h-16 text-slate-300" />
    </div>
  {/if}

  <h3 class="text-lg font-semibold text-slate-600">{title}</h3>

  {#if description}
    <p class="text-sm text-slate-400 mt-2 max-w-sm">{description}</p>
  {/if}

  {#if ctaLabel}
    <div class="mt-6">
      {#if ctaHref}
        <a
          href={ctaHref}
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-600 text-white text-sm font-medium hover:bg-blue-700 transition-colors duration-150"
        >
          {ctaLabel}
        </a>
      {:else}
        <button
          type="button"
          onclick={handleCta}
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-600 text-white text-sm font-medium hover:bg-blue-700 transition-colors duration-150"
        >
          {ctaLabel}
        </button>
      {/if}
    </div>
  {/if}
</div>
