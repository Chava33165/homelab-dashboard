<script lang="ts">
  let {
    percent,
    label = "",
    alertAt = 85,
  }: { percent: number; label?: string; alertAt?: number } = $props();

  let clamped = $derived(Math.max(0, Math.min(100, percent)));
  let alert = $derived(clamped >= alertAt);
</script>

<div class="bar-wrap">
  {#if label}
    <div class="row">
      <span class="label">{label}</span>
      <span class="pct" class:alert>{clamped.toFixed(0)}%</span>
    </div>
  {/if}
  <div class="track">
    <!-- scaleX en vez de width: el cambio cada 2s se compone en GPU sin relayout. -->
    <div class="fill" class:alert style:transform="scaleX({clamped / 100})"></div>
  </div>
</div>

<style>
  .bar-wrap {
    width: 100%;
  }
  .row {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    margin-bottom: 7px;
  }
  .label {
    color: var(--text-secondary);
    font-weight: 500;
  }
  .pct {
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--text);
  }
  .pct.alert {
    color: var(--danger);
  }
  .track {
    height: 8px;
    border-radius: 99em;
    background: color-mix(in srgb, var(--c-glass) 10%, transparent);
    overflow: hidden;
  }
  .fill {
    height: 100%;
    width: 100%;
    border-radius: 99em;
    background: var(--text);
    transform-origin: left center;
    transition: transform var(--dur-slow) var(--ease-out);
  }
  .fill.alert {
    background: var(--danger);
  }
</style>
