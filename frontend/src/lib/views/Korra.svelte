<script lang="ts">
  import { onMount } from "svelte";
  import { api, ApiError } from "../api/client";
  import type { KorraStats } from "../api/types";
  import GlassCard from "../components/GlassCard.svelte";
  import CountingNumber from "../components/CountingNumber.svelte";
  import Heatmap from "../components/Heatmap.svelte";
  import Bars from "../components/Bars.svelte";

  // El resumen lo genera KORRA cada 15 min: se pide una vez al entrar, sin polling.
  let stats = $state<KorraStats | null>(null);
  let error = $state<string | null>(null);

  onMount(async () => {
    try {
      stats = await api<KorraStats>("/korra/stats");
    } catch (e) {
      error =
        e instanceof ApiError && e.status === 404
          ? "KORRA todavía no ha generado estadísticas. Se crean cada 15 minutos mientras el asistente está encendido."
          : e instanceof Error
            ? e.message
            : "No se pudieron cargar las estadísticas.";
    }
  });

  const MONTHS = ["ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"];
  const WEEKDAYS = ["L", "M", "M", "J", "V", "S", "D"];
  const WEEKDAYS_LONG = ["lunes", "martes", "miércoles", "jueves", "viernes", "sábado", "domingo"];
  const PINK = "var(--c-action)";
  const INK = "var(--text)";

  // Nombres de herramientas del bot (bot.py) como los diria Chava.
  const TOOLS: Record<string, string> = {
    delegar_a_claude: "Mandar a Claude",
    seguimiento: "Ver trabajos",
    listar_archivos: "Buscar archivos",
    mandar_archivo: "Mandar archivos",
    archivos_entrada: "Archivos recibidos",
    contestar_con_voz: "Contestar con voz",
    ofrecer_opciones: "Botones",
    imprimir: "Imprimir",
    impresora_estado: "Estado impresora",
    recordatorio_crear: "Recordatorios",
    habilidad: "Abrir manual",
  };
  const toolName = (n: string) => TOOLS[n] ?? (n.charAt(0).toUpperCase() + n.slice(1)).replaceAll("_", " ");

  function fmtTokens(n: number): string {
    if (n >= 1e9) return `${(n / 1e9).toFixed(1)} B`;
    if (n >= 1e6) return `${(n / 1e6).toFixed(1)} M`;
    if (n >= 1e3) return `${(n / 1e3).toFixed(n >= 1e4 ? 0 : 1)} k`;
    return String(Math.round(n));
  }

  // DeepSeek cuesta fracciones de centavo: con 2 decimales todo saldria "$0.00".
  function fmtUSD(n: number): string {
    if (n === 0) return "$0";
    if (n < 0.01) return `$${n.toFixed(4)}`;
    if (n < 1) return `$${n.toFixed(3)}`;
    return `$${n.toFixed(2)}`;
  }

  function monthLabel(mes: string, long = false): string {
    const [y, m] = mes.split("-").map(Number);
    return long ? `${MONTHS[m - 1]} ${y}` : MONTHS[m - 1];
  }

  function since(ts: number): string {
    const min = Math.round((Date.now() / 1000 - ts) / 60);
    if (min < 1) return "ahora";
    if (min < 60) return `hace ${min} min`;
    const h = Math.round(min / 60);
    return h < 48 ? `hace ${h} h` : `hace ${Math.round(h / 24)} días`;
  }

  function niceDate(iso: string | null): string {
    if (!iso) return "—";
    const [y, m, d] = iso.split("-").map(Number);
    return `${d} ${MONTHS[m - 1]}${y !== new Date().getFullYear() ? ` ${y}` : ""}`;
  }

  type Metric = "mensajes" | "tokens" | "dinero";
  let metric = $state<Metric>("mensajes");

  let heatValues = $derived.by(() => {
    const out: Record<string, number> = {};
    for (const [day, d] of Object.entries(stats?.dias ?? {})) {
      out[day] =
        metric === "mensajes" ? d.mensajes
        : metric === "tokens" ? d.tokens_deepseek + d.tokens_claude
        : d.costo_deepseek + d.costo_claude;
    }
    return out;
  });
  let heatFormat = $derived(
    metric === "mensajes" ? (n: number) => `${n} ${n === 1 ? "mensaje" : "mensajes"}`
    : metric === "tokens" ? (n: number) => `${fmtTokens(n)} tokens`
    : fmtUSD
  );

  function describeDay(date: string): string {
    const d = stats?.dias[date];
    if (!d) return "sin actividad";
    const parts = [`${d.mensajes} ${d.mensajes === 1 ? "mensaje" : "mensajes"}`];
    if (d.trabajos) parts.push(`${d.trabajos} ${d.trabajos === 1 ? "trabajo" : "trabajos"} de Claude`);
    const tokens = d.tokens_deepseek + d.tokens_claude;
    if (tokens) parts.push(`${fmtTokens(tokens)} tokens`);
    const cost = d.costo_deepseek + d.costo_claude;
    if (cost) parts.push(fmtUSD(cost));
    return parts.join(" · ");
  }

  let thisMonth = $derived(stats?.meses.find((m) => m.mes === stats!.hoy.slice(0, 7)));
  let prevMonth = $derived(stats && thisMonth ? stats.meses[stats.meses.indexOf(thisMonth) - 1] : undefined);
  const monthTokens = (m?: KorraStats["meses"][number]) =>
    m ? m.deepseek.entrada + m.deepseek.cache + m.deepseek.salida + m.claude.entrada + m.claude.cache + m.claude.salida : 0;
  let tokensDelta = $derived(
    prevMonth && monthTokens(prevMonth) > 0 ? (monthTokens(thisMonth) / monthTokens(prevMonth) - 1) * 100 : null
  );

  let tokenBars = $derived(
    (stats?.meses ?? []).map((m) => ({
      label: monthLabel(m.mes),
      detail: monthLabel(m.mes, true),
      values: [m.deepseek.entrada + m.deepseek.cache + m.deepseek.salida, m.claude.entrada + m.claude.cache + m.claude.salida],
    }))
  );
  let moneyBars = $derived(
    (stats?.meses ?? []).map((m) => ({
      label: monthLabel(m.mes),
      detail: monthLabel(m.mes, true),
      values: [m.deepseek.costo, m.claude.costo_equiv],
    }))
  );
  let hourBars = $derived(
    (stats?.horas ?? []).map((v, h) => ({ label: String(h), detail: `${h}:00 – ${h}:59`, values: [v] }))
  );
  let peakHour = $derived(stats ? stats.horas.indexOf(Math.max(...stats.horas)) : 0);
  let weekBars = $derived(
    (stats?.dias_semana ?? []).map((v, i) => ({ label: WEEKDAYS[i], detail: WEEKDAYS_LONG[i], values: [v] }))
  );
  let peakDay = $derived(stats ? stats.dias_semana.indexOf(Math.max(...stats.dias_semana)) : 0);
  let maxTool = $derived(Math.max(1, ...(stats?.herramientas ?? []).map((t) => t.veces)));
  let topModel = $derived(stats?.modelos[0]?.modelo ?? "—");
  const msgs = (n: number) => `${n} ${n === 1 ? "mensaje" : "mensajes"}`;
</script>

<div class="page wide">
  <div class="page-head">
    <span class="eyebrow">Tu asistente</span>
    <h1 class="display-title chromatic-pop">KORRA</h1>
    {#if stats}
      <span class="ink-chip">Actualizado {since(stats.generado)}</span>
    {/if}
  </div>

  {#if error}
    <div class="empty glass">
      <svg aria-hidden="true" viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3l1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9z" /></svg>
      <p>{error}</p>
    </div>
  {:else if !stats}
    <div class="empty glass"><p>Cargando estadísticas…</p></div>
  {:else}
    {@const t = stats.totales}
    <div class="stack board stagger">
      <!-- Lo del mes, de un vistazo -->
      <div class="kpis">
        <GlassCard padding="16px">
          <span class="kpi-label">DeepSeek este mes</span>
          <span class="kpi-value">{fmtUSD(thisMonth?.deepseek.costo ?? 0)}</span>
          <span class="kpi-foot">{t.deepseek_desde ? `dinero real · desde ${niceDate(t.deepseek_desde)}` : "empieza a medirse hoy"}</span>
        </GlassCard>
        <GlassCard padding="16px">
          <span class="kpi-label">Saldo DeepSeek</span>
          <span class="kpi-value" class:low={stats.saldo_deepseek && stats.saldo_deepseek.total < 2}>
            {stats.saldo_deepseek ? `$${stats.saldo_deepseek.total.toFixed(2)}` : "—"}
          </span>
          <span class="kpi-foot">{stats.saldo_deepseek && stats.saldo_deepseek.total < 2 ? "¡recarga pronto!" : "lo que te queda"}</span>
        </GlassCard>
        <GlassCard padding="16px">
          <span class="kpi-label">Claude este mes</span>
          <span class="kpi-value">{fmtUSD(thisMonth?.claude.costo_equiv ?? 0)}</span>
          <span class="kpi-foot">equivalente · va en tu plan Pro</span>
        </GlassCard>
        <GlassCard padding="16px">
          <span class="kpi-label">Tokens este mes</span>
          <span class="kpi-value">{fmtTokens(monthTokens(thisMonth))}</span>
          <span class="kpi-foot">
            {#if tokensDelta !== null}
              <b class:up={tokensDelta > 0}>{tokensDelta > 0 ? "▲" : "▼"} {Math.abs(tokensDelta).toFixed(0)}%</b> vs mes pasado
            {:else}
              {thisMonth?.claude.trabajos ?? 0} trabajos de Claude
            {/if}
          </span>
        </GlassCard>
      </div>

      <!-- Mapa de actividad -->
      <GlassCard padding="20px">
        <div class="card-head">
          <h2>Actividad</h2>
          <div class="seg" role="radiogroup" aria-label="Qué medir">
            {#each [["mensajes", "Mensajes"], ["tokens", "Tokens"], ["dinero", "$"]] as [k, l] (k)}
              <button role="radio" aria-checked={metric === k} onclick={() => (metric = k as Metric)}>{l}</button>
            {/each}
          </div>
        </div>
        <Heatmap values={heatValues} today={stats.hoy} format={heatFormat} describe={describeDay} />
        <div class="streaks">
          <div><span class="big"><CountingNumber value={t.racha_actual} /></span><span>{t.racha_actual === 1 ? "día" : "días"} de racha</span></div>
          <div><span class="big"><CountingNumber value={t.racha_max} /></span><span>racha máxima</span></div>
          <div><span class="big"><CountingNumber value={t.dias_activos} /></span><span>días activos</span></div>
        </div>
      </GlassCard>

      <!-- Tokens por mes -->
      <GlassCard padding="20px">
        <div class="card-head"><h2>Tokens por mes</h2></div>
        <Bars
          label="Tokens por mes"
          bars={tokenBars}
          series={[{ name: "DeepSeek", color: PINK }, { name: "Claude", color: INK }]}
          format={fmtTokens}
        />
        <p class="note">Claude relee mucho contexto de caché en cada trabajo, por eso sus tokens se disparan.</p>
      </GlassCard>

      <!-- Dinero por mes -->
      <GlassCard padding="20px">
        <div class="card-head"><h2>Dinero por mes</h2><span class="total">{fmtUSD(t.costo_deepseek + t.costo_claude_equiv)} total</span></div>
        <Bars
          label="Dinero por mes en dólares"
          bars={moneyBars}
          series={[{ name: "DeepSeek (real)", color: PINK }, { name: "Claude (equiv.)", color: INK }]}
          format={fmtUSD}
        />
        <p class="note">DeepSeek se paga por uso. Claude va incluido en tu plan: su $ es lo que costaría por API, útil para ver cuánto límite se come.</p>
      </GlassCard>

      <!-- Cuándo -->
      <GlassCard padding="20px">
        <div class="card-head"><h2>¿Cuándo le hablas?</h2></div>
        <Bars label="Mensajes por hora del día" bars={hourBars} series={[{ name: "Mensajes", color: PINK }]} format={msgs} selected={peakHour} tickEvery={6} height={90} />
        <p class="note">Tu hora pico es a las <b>{peakHour}:00</b>.</p>
        <div class="spacer"></div>
        <Bars label="Mensajes por día de la semana" bars={weekBars} series={[{ name: "Mensajes", color: INK }]} format={msgs} selected={peakDay} height={70} />
        <p class="note">Le escribes más los <b>{WEEKDAYS_LONG[peakDay]}</b>.</p>
      </GlassCard>

      <!-- Herramientas -->
      {#if stats.herramientas.length}
        <GlassCard padding="20px">
          <div class="card-head"><h2>Lo que más le pides</h2></div>
          <ul class="tools">
            {#each stats.herramientas as tool (tool.nombre)}
              <li>
                <span class="tool-name">{toolName(tool.nombre)}</span>
                <span class="tool-bar"><i style:width="{(tool.veces / maxTool) * 100}%"></i></span>
                <b>{tool.veces}</b>
              </li>
            {/each}
          </ul>
        </GlassCard>
      {/if}

      <!-- Datos curiosos -->
      <div class="facts">
        <div class="fact glass"><b><CountingNumber value={t.mensajes} /></b><span>mensajes tuyos</span></div>
        <div class="fact glass"><b><CountingNumber value={t.respuestas} /></b><span>respuestas de KORRA</span></div>
        <div class="fact glass"><b>{t.dia_record.mensajes}</b><span>récord en un día · {niceDate(t.dia_record.fecha)}</span></div>
        <div class="fact glass"><b><CountingNumber value={t.recuerdos} /></b><span>cosas que recuerda de ti</span></div>
        <div class="fact glass"><b><CountingNumber value={t.tareas_hechas} /></b><span>tareas entregadas · {t.tareas_pendientes} pendientes</span></div>
        <div class="fact glass"><b><CountingNumber value={t.trabajos_claude} /></b><span>trabajos de Claude · {t.trabajos_claude_error} fallaron</span></div>
        <div class="fact glass wide"><b class="model">{topModel}</b><span>el modelo que más tokens ha movido</span></div>
        <div class="fact glass wide"><b>{niceDate(t.desde)}</b><span>primer mensaje guardado</span></div>
      </div>
    </div>
  {/if}
</div>

<style>
  .stack {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  /* iPad horizontal / compu: KPIs y el mapa de actividad a todo lo ancho, las gráficas en parejas. */
  .board > :global(:nth-child(-n + 2)),
  .board > .facts {
    grid-column: 1 / -1;
  }

  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 40px 24px;
    text-align: center;
    color: var(--text-secondary);
  }

  .empty svg {
    color: var(--c-action);
  }

  .kpis {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .kpis :global(.card) {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .kpi-label {
    font-size: 11px;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--text-secondary);
  }

  .kpi-value {
    font-family: var(--font-display);
    font-size: 30px;
    font-weight: 800;
    letter-spacing: -0.03em;
    font-variant-numeric: tabular-nums;
    line-height: 1.1;
  }

  .kpi-value.low {
    color: var(--c-action);
  }

  .kpi-foot {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .kpi-foot b {
    color: var(--text-secondary);
  }

  .kpi-foot b.up {
    color: var(--c-action);
  }

  .card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 14px;
  }

  h2 {
    margin: 0;
    font-size: 17px;
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .total {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-secondary);
    font-variant-numeric: tabular-nums;
  }

  .seg {
    display: flex;
    padding: 3px;
    border-radius: 99em;
    border: 1.5px solid var(--border-strong);
  }

  .seg button {
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
    touch-action: manipulation;
    transition: background-color var(--dur-fast) ease, color var(--dur-fast) ease;
  }

  .seg button[aria-checked="true"] {
    background: var(--text);
    color: var(--bg);
  }

  .seg button:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 1px;
  }

  .streaks {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
    margin-top: 16px;
    padding-top: 14px;
    border-top: 1px solid var(--border);
  }

  .streaks div {
    display: flex;
    flex-direction: column;
    font-size: 12px;
    color: var(--text-secondary);
  }

  .streaks .big {
    font-family: var(--font-display);
    font-size: 24px;
    font-weight: 800;
    color: var(--text);
    font-variant-numeric: tabular-nums;
  }

  .note {
    margin: 12px 0 0;
    font-size: 12px;
    line-height: 1.45;
    color: var(--text-tertiary);
  }

  .note b {
    color: var(--text);
  }

  .spacer {
    height: 22px;
  }

  .tools {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .tools li {
    display: grid;
    grid-template-columns: minmax(0, 9.5em) 1fr auto;
    align-items: center;
    gap: 10px;
    font-size: 13px;
  }

  .tool-name {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--text-secondary);
  }

  .tool-bar {
    height: 10px;
  }

  .tool-bar i {
    display: block;
    height: 100%;
    min-width: 4px;
    border-radius: 0 4px 4px 0;
    background: var(--c-action);
  }

  .tools b {
    font-variant-numeric: tabular-nums;
  }

  .facts {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .fact {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 14px 16px;
    border-radius: 20px;
    animation: rise var(--dur-slow) var(--ease-out) both;
  }

  .fact.wide {
    grid-column: span 2;
  }

  .fact b {
    font-family: var(--font-display);
    font-size: 24px;
    font-weight: 800;
    letter-spacing: -0.02em;
    font-variant-numeric: tabular-nums;
  }

  .fact b.model {
    font-size: 18px;
    overflow-wrap: anywhere;
  }

  .fact span {
    font-size: 12px;
    color: var(--text-secondary);
  }

  @media (min-width: 640px) {
    .kpis {
      grid-template-columns: repeat(4, 1fr);
    }
    .facts {
      grid-template-columns: repeat(4, 1fr);
    }
  }
</style>
