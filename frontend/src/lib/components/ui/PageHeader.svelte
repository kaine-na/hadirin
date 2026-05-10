<script lang="ts">
  interface Props {
    title: string;
    description?: string;
    icon?: any;
  }

  let { title, description, icon: Icon }: Props = $props();
</script>

<div class="mb-6">
  <div class="flex items-center gap-3">
    {#if Icon}
      <div class="w-9 h-9 rounded-xl bg-blue-100 flex items-center justify-center flex-shrink-0">
        <Icon class="w-5 h-5 text-blue-600" />
      </div>
    {/if}
    <div>
      <h1 class="text-2xl font-bold text-slate-800 leading-tight">{title}</h1>
      {#if description}
        <p class="text-sm text-slate-500 mt-1">{description}</p>
      {/if}
    </div>
  </div>
</div>
