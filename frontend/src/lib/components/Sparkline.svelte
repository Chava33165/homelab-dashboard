<script lang="ts">
  let {
    data,
    color = "var(--c-action)",
    height = 42,
    fill = true,
  }: { data: number[]; color?: string; height?: number; fill?: boolean } = $props();

  const width = 100;
  const id = `spark-${Math.random().toString(36).slice(2)}`;

  let points = $derived.by(() => {
    if (data.length < 2) return "";
    const max = Math.max(...data, 1);
    const min = Math.min(...data, 0);
    const range = max - min || 1;
    const step = width / (data.length - 1);
    return data.map((v, i) => `${(i * step).toFixed(2)},${(height - ((v - min) / range) * height).toFixed(2)}`).join(" ");
  });

  let fillPath = $derived(points ? `0,${height} ${points} ${width},${height}` : "");
</script>

<svg viewBox="0 0 {width} {height}" preserveAspectRatio="none" class="sparkline" style:height="{height}px">
  {#if fill && fillPath}
    <defs>
      <linearGradient {id} x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color={color} stop-opacity="0.22" />
        <stop offset="100%" stop-color={color} stop-opacity="0" />
      </linearGradient>
    </defs>
    <polygon points={fillPath} fill="url(#{id})" />
  {/if}
  {#if points}
    <polyline
      {points}
      fill="none"
      stroke={color}
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      vector-effect="non-scaling-stroke"
    />
  {/if}
</svg>

<style>
  .sparkline {
    display: block;
    width: 100%;
    overflow: visible;
    margin-top: 6px;
  }
</style>
