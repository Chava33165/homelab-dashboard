<script lang="ts">
  import { onMount } from "svelte";

  // Mapa de actividad tipo GitHub: 53 semanas x 7 dias (lunes arriba), 5 niveles de un
  // solo tono (rosa). Los niveles salen de los cuartiles de los dias con actividad, asi
  // un dia "normal" no se ve vacio aunque haya un dia record enorme. Cada celda es un
  // boton (foco ambulante: solo la seleccionada entra al Tab; flechas para moverse).
  let {
    values,
    today,
    format,
    describe,
  }: {
    values: Record<string, number>;
    today: string; // AAAA-MM-DD en hora local de KORRA
    format: (n: number) => string;
    describe?: (date: string) => string;
  } = $props();

  const MONTHS = ["ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"];
  const WEEKDAYS = ["lun", "mar", "mié", "jue", "vie", "sáb", "dom"];

  function iso(d: Date): string {
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
  }

  // Fechas como dias de calendario (mediodia local) para no tropezar con cambios de horario.
  let days = $derived.by(() => {
    const [y, m, d] = today.split("-").map(Number);
    const end = new Date(y, m - 1, d, 12);
    const start = new Date(end);
    start.setDate(end.getDate() - ((end.getDay() + 6) % 7) - 52 * 7);
    const out: { date: string; week: number; dow: number; day: Date }[] = [];
    for (let i = 0, cur = new Date(start); cur <= end; i++, cur.setDate(cur.getDate() + 1)) {
      out.push({ date: iso(cur), week: Math.floor(i / 7), dow: i % 7, day: new Date(cur) });
    }
    return out;
  });

  let thresholds = $derived.by(() => {
    const nz = days.map((d) => values[d.date] ?? 0).filter((v) => v > 0).sort((a, b) => a - b);
    if (!nz.length) return [Infinity, Infinity, Infinity];
    const q = (p: number) => nz[Math.min(nz.length - 1, Math.floor(p * nz.length))];
    return [q(0.25), q(0.5), q(0.75)];
  });

  function level(v: number): number {
    if (v <= 0) return 0;
    return 1 + thresholds.filter((t) => v > t).length;
  }

  let monthLabels = $derived(
    days
      .filter((d) => d.day.getDate() <= 7 && d.dow === 0)
      .map((d) => ({ week: d.week, label: MONTHS[d.day.getMonth()] }))
  );

  // svelte-ignore state_referenced_locally -- arranca en hoy; luego lo mueve el usuario.
  let selected = $state(today);
  let sel = $derived(days.find((d) => d.date === selected) ?? days[days.length - 1]);
  let scroller: HTMLDivElement;
  let grid: HTMLDivElement;

  function label(d: { date: string; day: Date }): string {
    return `${WEEKDAYS[(d.day.getDay() + 6) % 7]} ${d.day.getDate()} ${MONTHS[d.day.getMonth()]}`;
  }

  function onKey(e: KeyboardEvent) {
    const step = { ArrowLeft: -7, ArrowRight: 7, ArrowUp: -1, ArrowDown: 1 }[e.key];
    if (step === undefined) return;
    e.preventDefault();
    const i = days.findIndex((d) => d.date === selected);
    const next = days[Math.max(0, Math.min(days.length - 1, i + step))];
    selected = next.date;
    requestAnimationFrame(() => grid.querySelector<HTMLButtonElement>(`[data-date="${next.date}"]`)?.focus());
  }

  // En el cel no caben 53 semanas: el mapa se desplaza solo y arranca mostrando hoy.
  onMount(() => {
    scroller.scrollLeft = scroller.scrollWidth;
  });
</script>

<p class="readout" aria-live="polite">
  <b>{label(sel)}</b>
  <span>{describe ? describe(sel.date) : format(values[sel.date] ?? 0)}</span>
</p>

<div class="scroller" bind:this={scroller}>
  <div class="months" style:grid-template-columns="repeat({days[days.length - 1].week + 1}, var(--cell))" aria-hidden="true">
    {#each monthLabels as m (m.week)}
      <span style:grid-column="{m.week + 1} / span 3">{m.label}</span>
    {/each}
  </div>
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -- el grid delega las flechas a sus botones -->
  <div class="grid" role="grid" aria-label="Actividad por día, último año" tabindex="-1" bind:this={grid} onkeydown={onKey}>
    {#each days as d (d.date)}
      <button
        class="cell l{level(values[d.date] ?? 0)}"
        class:sel={d.date === selected}
        style:grid-column={d.week + 1}
        style:grid-row={d.dow + 1}
        data-date={d.date}
        tabindex={d.date === selected ? 0 : -1}
        aria-label="{label(d)}: {format(values[d.date] ?? 0)}"
        onclick={() => (selected = d.date)}
      ></button>
    {/each}
  </div>
</div>

<div class="scale" aria-hidden="true">
  <span>Menos</span>
  {#each [0, 1, 2, 3, 4] as l (l)}<i class="cell l{l}"></i>{/each}
  <span>Más</span>
</div>

<style>
  .readout {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 10px;
    min-height: 22px;
    margin: 0 0 12px;
    font-size: 13px;
    color: var(--text-secondary);
    font-variant-numeric: tabular-nums;
  }

  .readout b {
    color: var(--text);
  }

  .scroller {
    --cell: 13px;
    --gap: 3px;
    overflow-x: auto;
    overscroll-behavior-x: contain;
    scrollbar-width: none;
    padding-bottom: 2px;
  }

  .scroller::-webkit-scrollbar {
    display: none;
  }

  .months,
  .grid {
    display: grid;
    gap: var(--gap);
    width: max-content;
  }

  .months {
    height: 16px;
    font-size: 10px;
    font-weight: 600;
    color: var(--text-tertiary);
  }

  .grid {
    grid-template-rows: repeat(7, var(--cell));
    grid-auto-columns: var(--cell);
    outline: none;
  }

  .cell {
    width: var(--cell);
    height: var(--cell);
    padding: 0;
    border: none;
    border-radius: 3px;
    cursor: pointer;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
  }

  /* Un solo tono: de "casi superficie" a rosa pleno. */
  .l0 {
    background: color-mix(in srgb, var(--text) 9%, transparent);
  }
  .l1 {
    background: color-mix(in srgb, var(--c-action) 30%, var(--bg-elevated));
  }
  .l2 {
    background: color-mix(in srgb, var(--c-action) 55%, var(--bg-elevated));
  }
  .l3 {
    background: color-mix(in srgb, var(--c-action) 78%, var(--bg-elevated));
  }
  .l4 {
    background: var(--c-action);
  }

  .cell.sel {
    outline: 2px solid var(--text);
    outline-offset: 1px;
  }

  .cell:focus-visible {
    outline: 2px solid var(--text);
    outline-offset: 1px;
  }

  .scale {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 3px;
    margin-top: 10px;
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .scale span {
    margin: 0 4px;
  }

  .scale .cell {
    --cell: 11px;
    display: inline-block;
    cursor: default;
  }
</style>
