<script lang="ts">
  import { TrendingUp, TrendingDown } from 'lucide-svelte';

  type Color = 'blue' | 'green' | 'yellow' | 'purple' | 'red';

  interface Props {
    title: string;
    value: string | number;
    icon: any;
    color: Color;
    trend?: string;
    trendUp?: boolean;
  }

  let { title, value, icon: Icon, color, trend, trendUp }: Props = $props();

  const colorMap: Record<Color, { bg: string; text: string; iconBg: string }> = {
    blue:   { bg: 'bg-blue-50',   text: 'text-blue-600',   iconBg: 'bg-blue-100' },
    green:  { bg: 'bg-green-50',  text: 'text-green-600',  iconBg: 'bg-green-100' },
    yellow: { bg: 'bg-yellow-50', text: 'text-yellow-600', iconBg: 'bg-yellow-100' },
    purple: { bg: 'bg-purple-50', text: 'text-purple-600', iconBg: 'bg-purple-100' },
    red:    { bg: 'bg-red-50',    text: 'text-red-600',    iconBg: 'bg-red-100' },
  };

  const colors = $derived(colorMap[color] ?? colorMap.blue);
</script>

<div class="rounded-2xl shadow-md border border-slate-100 bg-white p-6 hover:shadow-xl hover:-translate-y-1 transition-all duration-200 cursor-default">
  <div class="flex items-start justify-between">
    <div class="flex-1 min-w-0">
      <p class="text-sm font-medium text-slate-500 truncate">{title}</p>
      <p class="mt-1 text-3xl font-bold text-slate-800 leading-tight">{value}</p>
      {#if trend}
        <div class="mt-2 flex items-center gap-1">
          {#if trendUp}
            <TrendingUp class="w-4 h-4 text-green-500" />
            <span class="text-xs font-medium text-green-600">{trend}</span>
          {:else}
            <TrendingDown class="w-4 h-4 text-red-500" />
            <span class="text-xs font-medium text-red-600">{trend}</span>
          {/if}
        </div>
      {/if}
    </div>
    <div class="ml-4 flex-shrink-0">
      <div class="w-12 h-12 rounded-xl {colors.iconBg} flex items-center justify-center">
        <Icon class="w-6 h-6 {colors.text}" />
      </div>
    </div>
  </div>
</div>
