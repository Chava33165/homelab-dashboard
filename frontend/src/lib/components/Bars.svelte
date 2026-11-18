<script lang="ts">
  // Columnas (apiladas si hay varias series) hechas con HTML: nitidas a cualquier ancho y
  // cada columna es un boton -- tocarla muestra sus valores en la linea de lectura de
  // arriba. Por defecto se lee la ultima (el mes/dia actual) o la marcada en `selected`.
  export type BarSeries = { name: string; color: string };
  export type Bar = { label: string; detail?: string; values: number[] };

  let {
    bars,
    series,
    format,
    selected = bars.length - 1,
    height = 120,
    tickEvery = 1,
    label,
    onpick,
  }: {
    bars: Bar[];
    series: BarSeries[];
    format: (n: number) => string;
    selected?: number;
    height?: number;
    tickEvery?: number;
    label: string;
    onpick?: (i: number) => void; // ademas de mostrarla, avisa que barra se toco (p. ej. para abrir ese dia)
  } = $props();

  // svelte-ignore state_referenced_locally -- la seleccion inicial es solo el punto de partida.
  let current = $state(selected);
  let max = $derived(Math.max(1e-9, ...bars.map((b) => b.values.reduce((a, v) => a + v, 0))));
  let active = $derived(bars[Math.min(current, bars.length - 1)]);
</script>

<div class="bars-wrap">
  <p class="readout" aria-live="polite">
    <span class="readout-label">{active?.detail ?? active?.label}</span>
    {#if active}
      {#each series as s, i (s.name)}
        <span class="readout-item">
          {#if series.length > 1}<i style:background={s.color}></i>{s.name}{/if}
          <b>{format(active.values[i] ?? 0)}</b>
        </span>
      {/each}
    {/if}
  </p>

  <div class="plot" style:height="{height}px" role="group" aria-label={label}>
    {#each bars as b, i (b.label + i)}
      {@const total = b.values.reduce((a, v) => a + v, 0)}
      <button
        class="col"
        class:on={i === current}
        onclick={() => {
          current = i;
          onpick?.(i);
        }}
        aria-pressed={i === current}
        aria-label="{b.detail ?? b.label}: {series.map((s, k) => `${s.name} ${format(b.values[k] ?? 0)}`).join(', ')}"
      >
        <span class="stack" style:height="{(total / max) * 100}%">
          {#each b.values as v, k (k)}
            {#if v > 0}
              <span class="seg" style:flex-grow={v} style:background={series[k].color}></span>
            {/if}
          {/each}
        </span>
      </button>
    {/each}
  </div>

  <div class="ticks" aria-hidden="true">
    {#each bars as b, i (b.label + i)}
      <span class:on={i === current}>{i % tickEvery === 0 || i === bars.length - 1 ? b.label : ""}</span>
    {/each}
  </div>

  {#if series.length > 1}
    <div class="legend">
      {#each series as s (s.name)}
        <span><i style:background={s.color}></i>{s.name}</span>
      {/each}
    </div>
  {/if}
</div>

<style>
  .readout {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 4px 14px;
    min-height: 22px;
    margin: 0 0 12px;
    font-size: 13px;
    color: var(--text-secondary);
    font-variant-numeric: tabular-nums;
  }

  .readout-label {
    font-weight: 700;
    color: var(--text);
  }

  .readout-item {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .readout-item b {
    color: var(--text);
  }

  i {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 2px;
  }

  /* Linea base de tinta, rejilla nula: los datos son lo unico que "grita". */
  .plot {
    display: flex;
    align-items: flex-end;
    gap: 2px;
    border-bottom: 1px solid var(--border-strong);
  }

  .col {
    flex: 1;
    min-width: 0;
    height: 100%;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding: 0;
    border: none;
    background: none;
    cursor: pointer;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
  }

  .col:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 2px;
    border-radius: 4px;
  }

  .stack {
    width: 100%;
    max-width: 24px;
    min-height: 0;
    display: flex;
    flex-direction: column-reverse;
    gap: 2px;
    border-radius: 4px 4px 0 0;
    overflow: hidden;
    opacity: 0.55;
    transform-origin: bottom;
    animation: grow var(--dur-slow) var(--ease-out) both;
    transition: opacity var(--dur-fast) ease;
  }

  .col.on .stack {
    opacity: 1;
  }

  @media (hover: hover) and (pointer: fine) {
    .col:hover .stack {
      opacity: 0.85;
    }
  }

  .seg {
    flex-basis: 0;
    min-height: 2px;
  }

  @keyframes grow {
    from {
      transform: scaleY(0);
    }
  }

  .ticks {
    display: flex;
    gap: 2px;
    margin-top: 6px;
  }

  .ticks span {
    flex: 1;
    min-width: 0;
    text-align: center;
    font-size: 10px;
    font-weight: 600;
    color: var(--text-tertiary);
    white-space: nowrap;
    overflow: visible;
  }

  .ticks span.on {
    color: var(--text);
  }

  .legend {
    display: flex;
    gap: 14px;
    margin-top: 10px;
    font-size: 12px;
    color: var(--text-secondary);
  }

  .legend span {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  @media (prefers-reduced-motion: reduce) {
    .stack {
      animation: none;
    }
  }
</style>
