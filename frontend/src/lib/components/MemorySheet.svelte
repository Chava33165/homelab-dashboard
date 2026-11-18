<script lang="ts">
  import { onDestroy, tick } from "svelte";
  import { fly, fade } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { api } from "../api/client";
  import { latestMetrics } from "../stores/metrics";
  import { services, prettyName } from "../stores/services";
  import { formatBytes } from "../format";
  import PowerSwitch from "./PowerSwitch.svelte";

  let { open = $bindable(false) }: { open?: boolean } = $props();

  interface HostProcess {
    name: string;
    count: number;
    mem_bytes: number;
  }

  // Nombres de proceso del kernel (max 15 chars) -> algo que Chava reconozca.
  const FRIENDLY: Record<string, { label: string; hint: string }> = {
    claude: { label: "Claude Code", hint: "Sesiones de Claude abiertas en la terminal o VS Code" },
    dockerd: { label: "Docker", hint: "El motor que corre todos los contenedores" },
    containerd: { label: "containerd", hint: "Parte de Docker: administra los contenedores" },
    "containerd-shim": { label: "Docker shims", hint: "Un vigilante pequeño por cada contenedor" },
    tailscaled: { label: "Tailscale", hint: "Tu VPN para entrar desde fuera" },
    node: { label: "Node.js", hint: "Normalmente el servidor de VS Code remoto" },
    "caveman-proxy": { label: "Proxy caveman", hint: "Proxy local que usa Claude Code" },
    systemd: { label: "systemd", hint: "Sistema base de la Pi" },
    "systemd-journal": { label: "Registros del sistema", hint: "journald, guarda los logs" },
    sshd: { label: "SSH", hint: "Conexiones remotas a la terminal" },
    otros: { label: "Otros procesos", hint: "Todo lo demás del sistema" },
  };

  let procs = $state<HostProcess[] | null>(null);
  let procsError = $state(false);
  let timer: ReturnType<typeof setInterval> | undefined;

  async function loadProcs() {
    try {
      procs = await api<HostProcess[]>("/memory/processes");
      procsError = false;
    } catch {
      procsError = true;
    }
  }

  // Solo se consulta mientras el panel esta abierto: cerrado no cuesta nada.
  $effect(() => {
    if (!open) return;
    loadProcs();
    timer = setInterval(loadProcs, 5000);
    return () => clearInterval(timer);
  });

  let m = $derived($latestMetrics);
  let total = $derived(m ? m.mem.total_kb * 1024 : 0);
  let used = $derived(m ? m.mem.used_kb * 1024 : 0);

  let running = $derived(
    $services.filter((s) => s.running > 0).sort((a, b) => b.memBytes - a.memBytes)
  );
  let stopped = $derived($services.filter((s) => s.running === 0));
  let servicesBytes = $derived(running.reduce((acc, s) => acc + s.memBytes, 0));
  let systemBytes = $derived((procs ?? []).reduce((acc, p) => acc + p.mem_bytes, 0));
  // Lo que no es ni contenedor ni proceso: kernel, buffers, tablas de memoria, etc. Las
  // cifras de Docker incluyen algo de cache de archivos, asi que la suma puede pasarse
  // un poco del "usado" real: en ese caso queda en 0 y no se muestra.
  let kernelBytes = $derived(Math.max(0, used - servicesBytes - systemBytes));

  const pct = (b: number) => (total > 0 ? (b / total) * 100 : 0);

  // ---- presentacion: hoja inferior en cel, panel centrado en pantallas anchas ----
  const wide = window.matchMedia("(min-width: 640px)").matches;
  const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  const inParams = reduceMotion ? { y: 0, duration: 150 } : wide ? { y: 24, duration: 260, easing: cubicOut } : { y: 600, duration: 380, easing: cubicOut };
  const outParams = reduceMotion ? { y: 0, duration: 120 } : wide ? { y: 24, duration: 180 } : { y: 600, duration: 240 };

  let sheetEl: HTMLDivElement | undefined = $state();
  let closeBtn: HTMLButtonElement | undefined = $state();
  let returnFocus: HTMLElement | null = null;

  $effect(() => {
    if (!open) return;
    returnFocus = document.activeElement as HTMLElement | null;
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    tick().then(() => closeBtn?.focus());
    return () => {
      document.body.style.overflow = prevOverflow;
      returnFocus?.focus?.();
    };
  });

  function close() {
    open = false;
  }

  function onKeydown(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === "Escape") {
      e.preventDefault();
      close();
    } else if (e.key === "Tab" && sheetEl) {
      // Foco atrapado dentro del panel mientras esta abierto.
      const f = sheetEl.querySelectorAll<HTMLElement>('button:not([disabled]), [href], [tabindex]:not([tabindex="-1"])');
      if (f.length === 0) return;
      const first = f[0];
      const last = f[f.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    }
  }

  // ---- arrastrar para cerrar (sigue el dedo 1:1, decide por velocidad) ----
  let dragY = $state(0);
  let dragging = $state(false);
  let startY = 0;
  let samples: { y: number; t: number }[] = [];

  function onPointerDown(e: PointerEvent) {
    if (wide || e.button !== 0) return;
    dragging = true;
    startY = e.clientY;
    samples = [{ y: e.clientY, t: e.timeStamp }];
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  }

  function onPointerMove(e: PointerEvent) {
    if (!dragging) return;
    const dy = e.clientY - startY;
    // Hacia arriba no hay a donde ir: resistencia progresiva (rubber band).
    dragY = dy >= 0 ? dy : -((-dy * 0.55 * 200) / (200 + 0.55 * -dy));
    samples.push({ y: e.clientY, t: e.timeStamp });
    if (samples.length > 5) samples.shift();
  }

  function onPointerUp() {
    if (!dragging) return;
    dragging = false;
    const a = samples[0];
    const b = samples[samples.length - 1];
    const velocity = b.t > a.t ? (b.y - a.y) / (b.t - a.t) : 0; // px/ms
    const height = sheetEl?.offsetHeight ?? 600;
    if (dragY > height * 0.33 || velocity > 0.5) {
      close();
    } else {
      dragY = 0; // vuelve a su lugar con la transicion CSS
    }
  }

  $effect(() => {
    if (open) dragY = 0;
  });

  onDestroy(() => clearInterval(timer));
</script>

<svelte:window onkeydown={onKeydown} />

{#if open}
  <div class="scrim" transition:fade={{ duration: reduceMotion ? 120 : 220 }} onclick={close} aria-hidden="true"></div>

  <div
    class="sheet"
    class:dragging
    bind:this={sheetEl}
    role="dialog"
    aria-modal="true"
    aria-labelledby="mem-title"
    in:fly={inParams}
    out:fly={outParams}
    style:transform={dragY ? `translateY(${dragY}px)` : undefined}
  >
    <!-- Arrastrar es un atajo tactil; el boton Cerrar y Esc cubren lo mismo con teclado. -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="drag-zone"
      onpointerdown={onPointerDown}
      onpointermove={onPointerMove}
      onpointerup={onPointerUp}
      onpointercancel={onPointerUp}
    >
      <span class="grabber" aria-hidden="true"></span>
      <div class="sheet-head">
        <div>
          <span class="eyebrow">Memoria RAM</span>
          <h2 id="mem-title" class="big">{formatBytes(used)}</h2>
          <p class="of">de {formatBytes(total)} · {pct(used).toFixed(0)}% en uso</p>
        </div>
        <button class="close" bind:this={closeBtn} onclick={close} aria-label="Cerrar">
          <svg aria-hidden="true" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg>
        </button>
      </div>
    </div>

    <div class="sheet-body">
      <!-- Barra apilada: que parte de los 4 GB es de quien. -->
      <div class="stack" role="img" aria-label="Servicios {formatBytes(servicesBytes)}, sistema {formatBytes(systemBytes)}, kernel y caché {formatBytes(kernelBytes)}">
        <span class="seg seg-services" style:flex-grow={pct(servicesBytes)}></span>
        <span class="seg seg-system" style:flex-grow={pct(systemBytes)}></span>
        <span class="seg seg-kernel" style:flex-grow={pct(kernelBytes)}></span>
        <span class="seg seg-free" style:flex-grow={Math.max(0, 100 - pct(used))}></span>
      </div>
      <ul class="legend">
        <li><span class="sw seg-services"></span>Servicios <b>{formatBytes(servicesBytes)}</b></li>
        <li><span class="sw seg-system"></span>Sistema <b>{procs ? formatBytes(systemBytes) : "…"}</b></li>
        {#if kernelBytes > 1024 * 1024}
          <li><span class="sw seg-kernel"></span>Kernel <b>{formatBytes(kernelBytes)}</b></li>
        {/if}
      </ul>

      <section>
        <h3>Servicios <span class="hint">puedes apagarlos aquí</span></h3>
        <ul class="rows">
          {#each running as s (s.project)}
            <li class="row">
              <div class="row-main">
                <div class="row-top">
                  <span class="name">{prettyName(s.project)}</span>
                  <span class="mem">{formatBytes(s.memBytes)}</span>
                </div>
                <div class="mini-track"><span class="mini-fill" style:transform="scaleX({used > 0 ? s.memBytes / used : 0})"></span></div>
                {#if s.members.length > 1}
                  <p class="members">
                    {#each [...s.members].filter((c) => c.stats).sort((a, b) => (b.stats?.mem_usage_bytes ?? 0) - (a.stats?.mem_usage_bytes ?? 0)) as c, i (c.id)}
                      {i > 0 ? " · " : ""}{c.name.replace(`${s.project}-`, "").replace(/-1$/, "")} {formatBytes(c.stats?.mem_usage_bytes ?? 0)}
                    {/each}
                  </p>
                {/if}
              </div>
              {#if s.protected}
                <span class="system-tag" title="Aquí corre este dashboard">SISTEMA</span>
              {:else}
                <PowerSwitch service={s} />
              {/if}
            </li>
          {/each}
          {#each stopped as s (s.project)}
            <li class="row off">
              <div class="row-main">
                <div class="row-top">
                  <span class="name">{prettyName(s.project)}</span>
                  <span class="mem">Apagado</span>
                </div>
              </div>
              <PowerSwitch service={s} />
            </li>
          {/each}
        </ul>
      </section>

      <section>
        <h3>Sistema <span class="hint">fuera de Docker</span></h3>
        {#if procsError}
          <p class="note">No se pudo leer los procesos de la Pi.</p>
        {:else if !procs}
          <p class="note">Leyendo procesos…</p>
        {:else}
          <ul class="rows">
            {#each procs as p (p.name)}
              {@const f = FRIENDLY[p.name]}
              <li class="row">
                <div class="row-main">
                  <div class="row-top">
                    <span class="name">{f?.label ?? p.name}{p.count > 1 ? ` ×${p.count}` : ""}</span>
                    <span class="mem">{formatBytes(p.mem_bytes)}</span>
                  </div>
                  <div class="mini-track"><span class="mini-fill ink" style:transform="scaleX({used > 0 ? p.mem_bytes / used : 0})"></span></div>
                  {#if f}<p class="members">{f.hint}</p>{/if}
                </div>
              </li>
            {/each}
          </ul>
          <p class="note">Estos se cierran desde la terminal de la Pi; el dashboard a propósito no tiene permiso para matar procesos.</p>
        {/if}
      </section>

      <section>
        <h3>Kernel y caché {#if kernelBytes > 1024 * 1024}<span class="mem-inline">{formatBytes(kernelBytes)}</span>{/if}</h3>
        <p class="note">Lo que usa Linux para sí mismo (tablas de memoria, buffers, drivers). No se puede apagar y el kernel lo devuelve solo cuando hace falta.</p>
        {#if m && m.mem.swap_total_kb > 0}
          <p class="note">Swap: {formatBytes(m.mem.swap_used_kb * 1024)} de {formatBytes(m.mem.swap_total_kb * 1024)}{m.mem.swap_used_kb > m.mem.swap_total_kb * 0.8 ? " — casi llena: conviene apagar algo." : ""}</p>
        {/if}
      </section>
    </div>
  </div>
{/if}

<style>
  .scrim {
    position: fixed;
    inset: 0;
    z-index: 60;
    background: var(--scrim);
  }

  .sheet {
    position: fixed;
    z-index: 61;
    left: 0;
    right: 0;
    bottom: 0;
    max-height: 88dvh;
    display: flex;
    flex-direction: column;
    background: var(--bg-elevated);
    border: 1.5px solid var(--glass-border);
    border-bottom: none;
    border-radius: 28px 28px 0 0;
    box-shadow: 0 -4px 0 var(--panel-offset), 0 -20px 60px rgba(0, 0, 0, 0.35);
    padding-bottom: var(--safe-bottom);
    transition: transform var(--dur-slow) var(--ease-drawer);
    will-change: transform;
  }

  /* Mientras el dedo arrastra no hay transicion: la hoja va pegada al dedo 1:1. */
  .sheet.dragging {
    transition: none;
  }

  @media (min-width: 640px) {
    .sheet {
      left: 50%;
      right: auto;
      bottom: auto;
      top: 50%;
      width: min(560px, calc(100vw - 48px));
      max-height: min(760px, 86dvh);
      translate: -50% -50%;
      border-bottom: 1.5px solid var(--glass-border);
      border-radius: 28px;
      box-shadow: 6px 6px 0 var(--panel-offset), 0 30px 80px rgba(0, 0, 0, 0.35);
      padding-bottom: 0;
    }
  }

  .drag-zone {
    padding: 10px 22px 6px;
    touch-action: none;
    cursor: grab;
  }

  .sheet.dragging .drag-zone {
    cursor: grabbing;
  }

  .grabber {
    display: block;
    width: 40px;
    height: 5px;
    margin: 0 auto 12px;
    border-radius: 99em;
    background: var(--border-strong);
  }

  @media (min-width: 640px) {
    .grabber {
      display: none;
    }
    .drag-zone {
      padding-top: 22px;
      cursor: default;
    }
  }

  .sheet-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .big {
    font-family: var(--font-display);
    font-size: 44px;
    font-weight: 800;
    line-height: 1;
    letter-spacing: -0.035em;
    margin-top: 6px;
    font-variant-numeric: tabular-nums;
  }

  .of {
    margin: 6px 0 0;
    font-size: 14px;
    color: var(--text-secondary);
  }

  .close {
    flex-shrink: 0;
    width: 40px;
    height: 40px;
    display: grid;
    place-items: center;
    border-radius: 50%;
    border: 1.5px solid var(--border-strong);
    background: none;
    color: var(--text);
    cursor: pointer;
    transition: transform var(--dur-press) var(--ease-out);
  }

  .close:active {
    transform: scale(0.92);
  }

  .close:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 2px;
  }

  .sheet-body {
    overflow-y: auto;
    overscroll-behavior: contain;
    padding: 14px 22px 26px;
  }

  .stack {
    display: flex;
    gap: 3px;
    height: 14px;
    border-radius: 99em;
    overflow: hidden;
  }

  .seg {
    flex-basis: 0;
    min-width: 0;
    border-radius: 3px;
    transition: flex-grow var(--dur-slow) var(--ease-out);
  }

  .seg-services {
    background: var(--c-action);
  }
  .seg-system {
    background: var(--text);
  }
  .seg-kernel {
    background: var(--text-tertiary);
  }
  .seg-free {
    background: color-mix(in srgb, var(--c-glass) 10%, transparent);
  }

  .legend {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 16px;
    list-style: none;
    margin: 12px 0 0;
    padding: 0;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .legend li {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .legend b {
    color: var(--text);
    font-variant-numeric: tabular-nums;
  }

  .sw {
    width: 10px;
    height: 10px;
    border-radius: 3px;
  }

  section {
    margin-top: 26px;
  }

  h3 {
    display: flex;
    align-items: baseline;
    gap: 8px;
    font-size: 18px;
    font-weight: 800;
    letter-spacing: -0.02em;
    margin-bottom: 8px;
  }

  .hint {
    font-family: var(--font-body);
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0;
    color: var(--text-tertiary);
  }

  .mem-inline {
    margin-left: auto;
    font-size: 15px;
    font-variant-numeric: tabular-nums;
  }

  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .row {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px 0;
    border-bottom: 1px solid var(--border);
  }

  .row:last-child {
    border-bottom: none;
  }

  .row.off .name {
    color: var(--text-tertiary);
  }

  .row-main {
    flex: 1;
    min-width: 0;
  }

  .row-top {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    font-size: 15px;
  }

  .name {
    font-weight: 700;
    letter-spacing: -0.01em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .mem {
    flex-shrink: 0;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }

  .row.off .mem {
    font-weight: 500;
    color: var(--text-tertiary);
  }

  .mini-track {
    height: 4px;
    margin-top: 7px;
    border-radius: 99em;
    background: color-mix(in srgb, var(--c-glass) 8%, transparent);
    overflow: hidden;
  }

  .mini-fill {
    display: block;
    width: 100%;
    height: 100%;
    border-radius: 99em;
    background: var(--c-action);
    transform-origin: left center;
    transition: transform var(--dur-slow) var(--ease-out);
  }

  .mini-fill.ink {
    background: var(--text);
  }

  .members {
    margin: 6px 0 0;
    font-size: 12px;
    color: var(--text-tertiary);
    font-variant-numeric: tabular-nums;
  }

  .system-tag {
    flex-shrink: 0;
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.06em;
    padding: 4px 9px;
    border-radius: 99em;
    color: var(--text-tertiary);
    border: 1px solid var(--border-strong);
  }

  .note {
    margin: 8px 0 0;
    font-size: 13px;
    line-height: 1.45;
    color: var(--text-secondary);
  }

  @media (prefers-reduced-transparency: reduce) {
    .scrim {
      background: rgba(0, 0, 0, 0.75);
    }
  }
</style>
