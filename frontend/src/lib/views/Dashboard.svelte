<script lang="ts">
  import { onMount } from "svelte";
  import { latestMetrics, cpuHistory, tempHistory, netRxHistory, startMetricsStream } from "../stores/metrics";
  import { startContainersStream } from "../stores/containers";
  import { logout } from "../stores/auth";
  import { theme, toggleTheme } from "../stores/theme";
  import { api, ApiError } from "../api/client";
  import type { SpeedtestResult } from "../api/types";
  import { formatBytes, formatRate, formatUptime } from "../format";
  import GlassCard from "../components/GlassCard.svelte";
  import Sparkline from "../components/Sparkline.svelte";
  import CountingNumber from "../components/CountingNumber.svelte";
  import MetricBar from "../components/MetricBar.svelte";
  import AlertBadge from "../components/AlertBadge.svelte";
  import MemorySheet from "../components/MemorySheet.svelte";

  onMount(() => {
    startMetricsStream();
    // El panel de RAM reparte la memoria por servicio con las stats de contenedores.
    startContainersStream();
  });

  let m = $derived($latestMetrics);
  let memOpen = $state(false);

  let speedtestRunning = $state(false);
  let speedtestResult = $state<SpeedtestResult | null>(null);
  let speedtestError = $state("");

  async function runSpeedtest() {
    speedtestRunning = true;
    speedtestError = "";
    speedtestResult = null;
    try {
      speedtestResult = await api<SpeedtestResult>("/speedtest", { method: "POST" });
    } catch (err) {
      speedtestError = err instanceof ApiError ? err.message : "falló el speedtest";
    } finally {
      speedtestRunning = false;
    }
  }

  // El disco "/" es el que importa; los demas (ej. /boot/firmware) van como linea chica.
  let mainDisk = $derived(m?.disks.find((d) => d.mount_point === "/") ?? m?.disks[0]);
  let otherDisks = $derived(m ? m.disks.filter((d) => d !== mainDisk) : []);

  const REAL_IFACE = /^(eth|wlan|en|tailscale)/;
  function sumReal(bytesByIface: Record<string, number>): number {
    let total = 0;
    for (const [iface, v] of Object.entries(bytesByIface)) {
      if (REAL_IFACE.test(iface)) total += v;
    }
    return total;
  }
</script>

<!-- En el cel el Panel es UNA pantalla de widgets (como la de inicio del iPhone): la
     cuadricula se estira al alto disponible y todo cabe sin hacer scroll. En pantallas
     muy bajitas o en compu vuelve a fluir normal. -->
<div class="page wide dash">
  <header class="page-head">
    <div class="title-block">
      <span class="eyebrow">
        <span class="live-dot" aria-hidden="true"></span>
        En vivo{#if m}&nbsp;· encendida {formatUptime(m.uptime_sec)}{/if}
      </span>
      <h1 class="display-title chromatic-pop">Homelab</h1>
    </div>

    <div class="header-actions">
      <button class="icon-btn" onclick={toggleTheme} aria-label={$theme === "dark" ? "Cambiar a tema Papel" : "Cambiar a tema Tinta"}>
        {#if $theme === "light"}
          <svg aria-hidden="true" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" /></svg>
        {:else}
          <svg aria-hidden="true" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="4.5" /><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" /></svg>
        {/if}
      </button>
      <button class="icon-btn danger" onclick={() => logout()} aria-label="Cerrar sesión">
        <svg aria-hidden="true" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 3v8" /><path d="M6.3 6.3a8 8 0 1 0 11.4 0" /></svg>
      </button>
    </div>
  </header>

  {#if m}
    <section class="grid stagger">
      <!-- RAM: ancho completo, es la que se toca para ver el desglose. -->
      <div class="cell wide">
        <GlassCard comic padding="0" onclick={() => (memOpen = true)} label="Memoria RAM: ver en qué se está usando">
          <div class="tile">
            <div class="metric-head">
              <span class="metric-label">Memoria RAM</span>
              <span class="tap-hint">
                Detalle
                <svg aria-hidden="true" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"><path d="M9 6l6 6-6 6" /></svg>
              </span>
            </div>
            <div class="value-row">
              <span class="metric-value"><CountingNumber value={m.mem.used_kb / 1024 / 1024} decimals={2} suffix=" GB" /></span>
              <span class="of">de {formatBytes(m.mem.total_kb * 1024)}</span>
            </div>
            <div class="push">
              <MetricBar percent={m.mem.used_percent} />
              {#if m.mem.swap_total_kb > 0}
                <p class="foot">{m.mem.used_percent.toFixed(0)}% en uso · swap {formatBytes(m.mem.swap_used_kb * 1024)}</p>
              {/if}
            </div>
          </div>
        </GlassCard>
      </div>

      <div class="cell">
        <GlassCard comic padding="0">
          <div class="tile">
            <div class="metric-head">
              <span class="metric-label">CPU</span>
              {#if m.throttle.available && m.throttle.raw !== "0x0"}
                <AlertBadge level="danger" text="Lenta" />
              {/if}
            </div>
            <span class="metric-value"><CountingNumber value={m.cpu_percent} decimals={0} suffix="%" /></span>
            <div class="spark"><Sparkline data={$cpuHistory} color="var(--c-action)" /></div>
          </div>
        </GlassCard>
      </div>

      <div class="cell">
        <GlassCard comic padding="0">
          <div class="tile">
            <div class="metric-head">
              <span class="metric-label">Temp</span>
              {#if m.temp_c > 75}
                <AlertBadge level="danger" text="Caliente" />
              {/if}
            </div>
            <span class="metric-value"><CountingNumber value={m.temp_c} decimals={0} suffix="°" /></span>
            <div class="spark"><Sparkline data={$tempHistory} color={m.temp_c > 75 ? "var(--danger)" : "var(--text)"} /></div>
          </div>
        </GlassCard>
      </div>

      <div class="cell">
        <GlassCard comic padding="0">
          <div class="tile">
            <div class="metric-head">
              <span class="metric-label">Red</span>
              <button class="mini-btn" onclick={runSpeedtest} disabled={speedtestRunning} aria-label="Medir velocidad de internet">
                <svg class:spinning={speedtestRunning} aria-hidden="true" viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2L4 14h7l-1 8 9-12h-7z" /></svg>
              </button>
            </div>
            <div class="rates">
              <span class="rate">
                <svg aria-label="Bajada" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14M6 13l6 6 6-6" /></svg>
                {formatRate(sumReal(m.net.rx_bytes_per_sec))}
              </span>
              <span class="rate">
                <svg aria-label="Subida" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"><path d="M12 19V5M6 11l6-6 6 6" /></svg>
                {formatRate(sumReal(m.net.tx_bytes_per_sec))}
              </span>
            </div>
            <p class="foot push" class:warn={!m.net.reachable || m.net.latency_ms > 150}>
              {#if speedtestRunning}
                Midiendo velocidad…
              {:else if speedtestResult}
                ↓ {speedtestResult.download_mbps.toFixed(0)} · ↑ {speedtestResult.upload_mbps.toFixed(0)} Mbps
              {:else if speedtestError}
                Speedtest falló
              {:else if !m.net.reachable}
                Sin internet
              {:else}
                Ping {m.net.latency_ms.toFixed(0)} ms
              {/if}
            </p>
          </div>
        </GlassCard>
      </div>

      {#if mainDisk}
        <div class="cell">
          <GlassCard comic padding="0">
            <div class="tile">
              <div class="metric-head">
                <span class="metric-label">Disco</span>
              </div>
              <span class="metric-value"><CountingNumber value={mainDisk.used_percent} decimals={0} suffix="%" /></span>
              <div class="push">
                <MetricBar percent={mainDisk.used_percent} />
                <p class="foot">
                  {formatBytes(mainDisk.total_bytes - mainDisk.used_bytes, 0)} libres
                  {#each otherDisks as d (d.mount_point)}<br />{d.mount_point.split("/").pop()} {d.used_percent.toFixed(0)}%{/each}
                </p>
              </div>
            </div>
          </GlassCard>
        </div>
      {/if}
    </section>
  {:else}
    <div class="loading-box glass">
      <span class="spinner" aria-hidden="true"></span>
      <p>Conectando con la Pi…</p>
    </div>
  {/if}
</div>

<MemorySheet bind:open={memOpen} />

<style>
  /* ---- Una sola pantalla en el cel ----
     La pagina mide exactamente la ventana (100dvh: se ajusta cuando aparece/desaparece
     la barra del navegador) y deja libre el espacio de la barra de pestañas. El
     encabezado toma lo que necesita y la cuadricula se estira con el resto. */
  .dash {
    height: 100dvh;
    display: flex;
    flex-direction: column;
    padding-top: calc(14px + var(--safe-top));
    padding-bottom: calc(104px + var(--safe-bottom));
  }

  .page-head {
    align-items: center;
    justify-content: space-between;
    flex-wrap: nowrap;
    margin-bottom: clamp(10px, 2dvh, 18px);
  }

  .title-block {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    width: auto;
  }

  .display-title {
    font-size: clamp(30px, 5dvh, 44px);
  }

  .live-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--c-action);
    animation: livePulse 1.6s ease-in-out infinite;
  }

  .header-actions {
    display: flex;
    gap: 10px;
  }

  .icon-btn {
    width: 44px;
    height: 44px;
    display: grid;
    place-items: center;
    border-radius: 50%;
    border: 1.5px solid var(--border-strong);
    background: var(--glass-bg);
    color: var(--text);
    cursor: pointer;
    touch-action: manipulation;
    transition:
      transform var(--dur-press) var(--ease-out),
      color var(--dur-fast) ease,
      border-color var(--dur-fast) ease;
  }

  .icon-btn:active {
    transform: scale(0.92);
  }

  .icon-btn:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 2px;
  }

  @media (hover: hover) and (pointer: fine) {
    .icon-btn:hover {
      border-color: var(--text);
    }
    .icon-btn.danger:hover {
      color: var(--danger);
      border-color: var(--danger);
    }
  }

  /* 2 columnas; RAM ocupa la fila de arriba completa y es un poco mas alta. Las filas
     se reparten el alto disponible, asi en un cel grande los widgets crecen y en uno
     chico se encogen, pero nunca hace falta bajar. */
  .grid {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-template-rows: 1.15fr 1fr 1fr;
    gap: 12px;
  }

  .cell {
    min-width: 0;
    min-height: 0;
    display: flex;
  }

  .cell > :global(.card) {
    flex: 1;
    min-height: 0;
    border-radius: 22px;
  }

  .wide {
    grid-column: 1 / -1;
  }

  .tile {
    position: relative;
    container-type: inline-size; /* los numeros crecen con el widget (iPad/Mac) */
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: clamp(12px, 1.9dvh, 18px) 16px;
    min-height: 0;
  }

  /* Empuja lo que sigue (barra, pie) al fondo del widget. */
  .tile > .push {
    margin-top: auto;
  }

  .metric-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    min-height: 26px;
  }

  .metric-label {
    font-size: 11px;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--text-secondary);
  }

  .tap-hint {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    font-size: 12px;
    font-weight: 700;
    color: var(--c-action);
  }

  .value-row {
    display: flex;
    align-items: baseline;
    gap: 8px;
  }

  .metric-value {
    font-family: var(--font-display);
    font-size: clamp(30px, min(5.4dvh, 17cqi), 76px);
    font-weight: 800;
    line-height: 1;
    letter-spacing: -0.04em;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .of {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-tertiary);
  }

  /* La grafica toma el espacio que sobre en el widget. */
  .spark {
    flex: 1;
    min-height: 18px;
    margin-top: 6px;
    display: flex;
    align-items: flex-end;
  }

  /* Widget grande (iPad/Mac): la grafica crece y llena el espacio libre del widget. */
  .spark :global(.sparkline) {
    height: min(100%, 220px) !important;
  }

  .spark :global(svg) {
    height: 100% !important;
    max-height: 70px;
    margin-top: 0;
  }

  .foot {
    margin: 6px 0 0;
    font-size: 12px;
    line-height: 1.35;
    color: var(--text-tertiary);
    font-variant-numeric: tabular-nums;
  }

  .foot.warn {
    color: var(--danger);
    font-weight: 700;
  }

  .rates {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-top: 4px;
  }

  .rate {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-family: var(--font-display);
    font-size: clamp(15px, 2.3dvh, 19px);
    font-weight: 800;
    letter-spacing: -0.02em;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .rate svg {
    flex-shrink: 0;
    color: var(--c-action);
  }

  /* Boton de speedtest chico en la esquina del widget de red (area tactil 44px). */
  .mini-btn {
    position: relative;
    width: 30px;
    height: 30px;
    display: grid;
    place-items: center;
    border-radius: 50%;
    border: 1.5px solid var(--border-strong);
    background: none;
    color: var(--c-action);
    cursor: pointer;
    touch-action: manipulation;
    transition: transform var(--dur-press) var(--ease-out);
  }

  .mini-btn::before {
    content: "";
    position: absolute;
    inset: -7px;
  }

  .mini-btn:active:not(:disabled) {
    transform: scale(0.9);
  }

  .mini-btn:disabled {
    cursor: progress;
  }

  .mini-btn:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 2px;
  }

  .spinning {
    animation: spin 0.9s linear infinite;
  }

  .loading-box {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 14px;
    padding: 48px 20px;
    color: var(--text-secondary);
  }

  .loading-box p {
    margin: 0;
  }

  .spinner {
    width: 26px;
    height: 26px;
    border-radius: 50%;
    border: 2.5px solid var(--border-strong);
    border-top-color: var(--c-action);
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  /* El Panel siempre es UNA pantalla de widgets que llena el alto (cel e iPad vertical: 2
     columnas). En iPad horizontal y compu: 4 columnas en 2 filas, todo a la vista. */
  @media (min-width: 1100px) {
    .grid {
      grid-template-columns: repeat(4, 1fr);
      grid-template-rows: 1fr 1fr;
      gap: 16px;
    }
    .wide,
    .grid > .cell:nth-child(4),
    .grid > .cell:nth-child(5) {
      grid-column: span 2;
    }
  }

  @media (min-width: 640px) {
    .dash {
      padding-bottom: calc(112px + var(--safe-bottom));
    }
  }

  /* Pantalla muy bajita (cel acostado, ventana chica): no cabe en una pantalla, fluye normal. */
  @media (max-height: 560px) {
    .dash {
      height: auto;
      padding-bottom: calc(120px + var(--safe-bottom));
    }
    .grid {
      grid-template-rows: none;
      grid-auto-rows: 170px;
    }
  }
</style>
