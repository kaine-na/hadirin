<script lang="ts">
	type Color = 'green' | 'yellow' | 'blue' | 'red' | 'slate' | 'primary';

	interface Props {
		color?: Color;
		children?: import('svelte').Snippet;
	}

	let { color = 'slate', children }: Props = $props();

	const colorClass: Record<Color, string> = {
		green: 'bg-green-100 text-green-800',
		yellow: 'bg-yellow-100 text-yellow-800',
		blue: 'bg-blue-100 text-blue-800',
		red: 'bg-red-100 text-red-800',
		slate: 'bg-slate-100 text-slate-700',
		primary: 'bg-primary-100 text-primary-800'
	};
</script>

<span
	class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {colorClass[color]}"
>
	{@render children?.()}
</span>
