<script lang="ts">
  import { api } from "../../api/client";
  import type { FinPeriodo } from "../../api/types";
  import { pesos, resumen } from "../../stores/finanzas";
  import Bars from "../Bars.svelte";

  // ¿Cuánto gasto? Día / Semana (jueves a miércoles) / Mes, con flechas para ir atrás.
  type Vista = "dia" | "semana" | "mes";
  let vista = $state<Vista>("dia");
  let fecha = $state<string | null>(null); // null = hoy
  let p = $state<FinPeriodo | null>(null);
  let cargando = $state(false);

  const MES = ["ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"];
  const MES_L = ["enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"];
  const DIA = ["dom", "lun", "mar", "mié", "jue", "vie", "sáb"];
  const d = (iso: string) => new Date(iso + "T12:00:00");

  $effect(() => {
    const q = new URLSearchParams({ vista, ...(fecha ? { fecha } : {}) });
    void $resumen; // se vuelve a pedir cuando cambian los datos (p. ej. un gasto nuevo)
    cargando = true;
    api<FinPeriodo>(`/finanzas/gastos/periodo?${q}`)
      .then((r) => (p = r))
      .catch(() => (p = null))
      .finally(() => (cargando = false));
  });

  function etiqueta(x: FinPeriodo): string {
    const i = d(x.inicio);
    const hoy = $resumen?.hoy;
    if (x.vista === "dia") return x.inicio === hoy ? "Hoy" : `${DIA[i.getDay()]} ${i.getDate()} ${MES[i.getMonth()]}`;
    if (x.vista === "semana") return x.siguiente === null ? "Esta semana" : `Semana del ${i.getDate()} ${MES[i.getMonth()]}`;
    return x.siguiente === null ? "Este mes" : `${MES_L[i.getMonth()]} ${i.getFullYear()}`;
  }

  let cambio = $derived(p && p.anterior_total > 0 ? (p.total / p.anterior_total - 1) * 100 : null);
  let nombres = $derived(Object.fromEntries(($resumen?.tarjetas ?? []).map((t) => [t.id, t.nombre])));
  let barras = $derived(
    (p?.barras ?? []).map((b) => {
      const x = d(b.fecha);
      return { label: vista === "semana" ? DIA[x.getDay()] : String(x.getDate()), detail: `${DIA[x.getDay()]} ${x.getDate()} ${MES[x.getMonth()]}`, values: [b.total] };
    })
  );
  // la lectura de las barras arranca en hoy (o en el ultimo dia del periodo si ya paso)
  let hoyIdx = $derived(Math.max(0, (p?.barras ?? []).filter((b) => b.fecha <= ($resumen?.hoy ?? "")).length - 1));
  let maxCat = $derived(Math.max(1, ...Object.values(p?.por_categoria ?? {})));

  function cambiarVista(v: Vista) {
    vista = v;
    fecha = null;
  }

  function hora(ts: number) {
    return new Date(ts * 1000).toLocaleTimeString("es-MX", { hour: "numeric", minute: "2-digit" });
  }
</script>

<div class="gp">
  <div class="card-head">
    <h2>¿Cuánto gasto?</h2>
    <div class="seg mini" role="radiogroup" aria-label="Periodo">
      {#each [["dia", "Día"], ["semana", "Semana"], ["mes", "Mes"]] as [k, l] (k)}
        <button role="radio" aria-checked={vista === k} onclick={() => cambiarVista(k as Vista)}>{l}</button>
      {/each}
    </div>
  </div>

  {#if p}
    <div class="nav">
      <button class="arrow" aria-label="Periodo anterior" onclick={() => (fecha = p!.anterior)}>‹</button>
      <span class="et">{etiqueta(p)}</span>
      <button class="arrow" aria-label="Periodo siguiente" disabled={!p.siguiente} onclick={() => (fecha = p!.siguiente)}>›</button>
    </div>

    <div class="tot" class:dim={cargando}>
      <b>{pesos(p.total)}</b>
      <span>
        {#if cambio !== null}<em class:up={cambio > 0}>{cambio > 0 ? "▲" : "▼"} {Math.abs(cambio).toFixed(0)}%</em> vs anterior{:else if p.anterior_total === 0 && p.total > 0}antes: $0{/if}
        {#if vista !== "dia"} · {pesos(p.promedio_diario, 0)} al día{/if}
      </span>
    </div>

    {#if vista !== "dia" && barras.length}
      {#key p.inicio + vista}
        <Bars label="Gasto por día" bars={barras} series={[{ name: "Gastado", color: "var(--c-action)" }]} format={(n) => pesos(n)} height={90} tickEvery={vista === "mes" ? 5 : 1}
          selected={hoyIdx} onpick={(i) => { vista = "dia"; fecha = p!.barras[i].fecha; }} />
      {/key}
      <p class="note">Toca una barra para ver ese día.</p>
    {/if}

    {#if Object.keys(p.por_categoria).length}
      <ul class="cats">
        {#each Object.entries(p.por_categoria) as [c, v] (c)}
          <li><span>{c}</span><i style:width="{(v / maxCat) * 100}%"></i><b>{pesos(v)}</b></li>
        {/each}
      </ul>
      <div class="chips met">
        {#each Object.entries(p.por_metodo) as [m, v] (m)}<span>{m === "efectivo" ? "Efectivo" : nombres[m] ?? m} · {pesos(v)}</span>{/each}
      </div>
    {/if}

    {#if p.lista.length}
      <ul class="lista">
        {#each p.lista.slice(0, vista === "dia" ? 50 : 12) as g (g.id)}
          <li>
            <span><b>{g.nota || g.categoria}</b><small>{g.categoria} · {g.metodo === "efectivo" ? "Efectivo" : nombres[g.metodo] ?? g.metodo}{g.tipo === "compra_meses" ? " a meses" : ""} · {vista === "dia" ? hora(g.ts) : new Date(g.ts * 1000).toLocaleDateString("es-MX", { day: "numeric", month: "short" })}</small></span>
            <b>{pesos(g.monto)}</b>
          </li>
        {/each}
      </ul>
      {#if vista !== "dia" && p.lista.length > 12}<p class="note">y {p.lista.length - 12} gastos más · toca un día para verlos todos</p>{/if}
    {:else}
      <p class="note">{vista === "dia" ? "Sin gastos este día 🎉" : "Sin gastos en este periodo."}</p>
    {/if}
  {:else}
    <p class="note">Cargando…</p>
  {/if}
</div>

<style>
  .card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 10px;
  }

  h2 {
    margin: 0;
    font-size: 17px;
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .seg.mini {
    display: flex;
    padding: 3px;
    border-radius: 99em;
    border: 1.5px solid var(--border-strong);
  }

  .seg.mini button {
    min-height: 30px;
    padding: 0 11px;
    border: none;
    border-radius: 99em;
    background: none;
    font: inherit;
    font-size: 12px;
    font-weight: 700;
    color: var(--text-secondary);
    cursor: pointer;
  }

  .seg.mini button[aria-checked="true"] {
    background: var(--text);
    color: var(--bg);
  }

  .nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 6px;
  }

  .et {
    font-weight: 800;
    font-size: 15px;
  }

  .arrow {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 1.5px solid var(--border-strong);
    background: none;
    color: var(--text);
    font-size: 22px;
    line-height: 1;
    cursor: pointer;
  }

  .arrow:disabled {
    opacity: 0.3;
    cursor: default;
  }

  .tot {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin: 6px 0 14px;
    transition: opacity var(--dur-fast) ease;
  }

  .tot.dim {
    opacity: 0.55;
  }

  .tot b {
    font-family: var(--font-display);
    font-size: 40px;
    font-weight: 800;
    letter-spacing: -0.03em;
    font-variant-numeric: tabular-nums;
  }

  .tot span {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .tot em {
    font-style: normal;
    font-weight: 800;
    color: var(--text);
  }

  .tot em.up {
    color: var(--c-action);
  }

  .cats {
    list-style: none;
    margin: 14px 0 10px;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .cats li {
    display: grid;
    grid-template-columns: 6.5em 1fr auto;
    align-items: center;
    gap: 10px;
    font-size: 13px;
  }

  .cats span {
    color: var(--text-secondary);
  }

  .cats i {
    display: block;
    height: 10px;
    min-width: 4px;
    border-radius: 0 4px 4px 0;
    background: var(--c-action);
  }

  .met {
    gap: 6px;
    margin-bottom: 8px;
  }

  .met span {
    padding: 5px 10px;
    border-radius: 99em;
    border: 1px solid var(--border-strong);
    font-size: 12px;
    font-weight: 700;
  }

  .lista {
    list-style: none;
    margin: 6px 0 0;
    padding: 0;
  }

  .lista li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 10px;
    padding: 9px 0;
    border-top: 1px solid var(--border);
  }

  .lista li > span {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .lista li b {
    font-size: 14px;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .lista small {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .note {
    margin: 8px 0 0;
    font-size: 12px;
    color: var(--text-tertiary);
  }
</style>
