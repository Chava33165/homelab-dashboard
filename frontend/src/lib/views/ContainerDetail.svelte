<script lang="ts">
  import { onMount, onDestroy, tick } from "svelte";
  import { containers, refreshContainers, startContainersStream } from "../stores/containers";
  import { openLogsStream } from "../api/sse";
  import { api, ApiError } from "../api/client";
  import { navigate } from "../stores/router";
  import { formatBytes, formatCreated } from "../format";
  import GlassCard from "../components/GlassCard.svelte";
  import MetricBar from "../components/MetricBar.svelte";
  import ConfirmDialog from "../components/ConfirmDialog.svelte";
  import { adapterRegistry } from "../adapters/adapterRegistry";

  let { id }: { id: string } = $props();

  let container = $derived($containers.find((c) => c.id === id));
  let AdapterPanel = $derived(container ? adapterRegistry[container.adapter] : undefined);

  let lines = $state<string[]>([]);
  let search = $state("");
  let logEl: HTMLDivElement | undefined = $state();
  let autoscroll = $state(true);
  let closeLogs: (() => void) | null = null;

  let filteredLines = $derived(
    search.trim() ? lines.filter((l) => l.toLowerCase().includes(search.toLowerCase())) : lines
  );

  const MAX_LINES = 500;

  onMount(() => {
    // Idempotente (guardada por un flag interno) -- necesaria aqui tambien, no solo en
    // Containers.svelte, para que esta vista reciba actualizaciones en vivo (stats,
    // cambios de estado) incluso si se entra directo a esta URL sin pasar antes por la
    // pestaña de Contenedores.
    startContainersStream();
    if (!container) refreshContainers().catch(() => void 0);
    closeLogs = openLogsStream(
      id,
      async (line) => {
        lines = [...lines, line].slice(-MAX_LINES);
        if (autoscroll) {
          await tick();
          if (logEl) logEl.scrollTo({ top: logEl.scrollHeight });
        }
      },
      () => (lines = [])
    );
  });

  onDestroy(() => {
    closeLogs?.();
  });

  function onLogScroll() {
    if (!logEl) return;
    const atBottom = logEl.scrollHeight - logEl.scrollTop - logEl.clientHeight < 40;
    autoscroll = atBottom;
  }

  function jumpToBottom() {
    if (!logEl) return;
    autoscroll = true;
    logEl.scrollTo({ top: logEl.scrollHeight, behavior: "smooth" });
  }

  type Action = "start" | "stop" | "restart";
  // El dialogo cierra solo al cancelar (open es $bindable); por eso va enlazado con
  // bind:open -- con open={confirmAction === x} el estado de aqui no se enteraba del
  // cierre y ese mismo boton ya no volvia a abrir el dialogo.
  let confirmAction = $state<Action>("stop");
  let confirmOpen = $state(false);

  function ask(action: Action) {
    confirmAction = action;
    confirmOpen = true;
  }
  let acting = $state(false);
  let actionError = $state("");

  async function runAction(action: Action) {
    acting = true;
    actionError = "";
    try {
      await api(`/containers/${id}/${action}`, { method: "POST" });
      await refreshContainers();
    } catch (err) {
      actionError = err instanceof ApiError ? err.message : "fallo la acción";
    } finally {
      acting = false;
    }
  }

  const actionLabels: Record<Action, { title: string; message: string; confirmLabel: string; danger: boolean }> = {
    start: { title: "¿Iniciar contenedor?", message: "Se va a iniciar este contenedor.", confirmLabel: "Iniciar", danger: false },
    stop: { title: "¿Detener contenedor?", message: "Se va a detener este contenedor. Se puede volver a iniciar despues.", confirmLabel: "Detener", danger: true },
    restart: { title: "¿Reiniciar contenedor?", message: "Se va a reiniciar este contenedor.", confirmLabel: "Reiniciar", danger: false },
  };
</script>

<div class="page">
  <button class="back-pill" onclick={() => navigate({ name: "containers" })}>
    <span>← Servicios</span>
  </button>

  {#if container}
    <header>
      <div class="header-badge-row">
        <span class="comic-chip">{container.adapter.toUpperCase()}</span>
        <span class="status-pill" class:running={container.state === "running"}>
          {container.status}
        </span>
      </div>
      <h1 class="chromatic-pop">{container.name}</h1>
      <p class="image">{container.image}</p>
      <p class="meta">Creado {formatCreated(container.created)}</p>
    </header>

    <div class="grid">
      {#if container.stats}
        <GlassCard comic>
          <span class="label">CPU USAGE</span>
          <div class="value">{container.stats.cpu_percent.toFixed(1)}%</div>
        </GlassCard>
        <GlassCard comic>
          <span class="label">MEMORIA RAM</span>
          <div class="value">{formatBytes(container.stats.mem_usage_bytes)}</div>
          <MetricBar percent={container.stats.mem_percent} label="de {formatBytes(container.stats.mem_limit_bytes)}" />
        </GlassCard>
      {/if}
      {#if container.ports.length > 0}
        <GlassCard comic>
          <span class="label">PUERTOS EXPUESTOS</span>
          <div class="ports">
            {#each container.ports as p}
              <span class="port">{p.public_port ? `${p.public_port}→` : ""}{p.private_port}/{p.type}</span>
            {/each}
          </div>
        </GlassCard>
      {/if}
    </div>

    {#if AdapterPanel}
      <div class="adapter-panel">
        <AdapterPanel data={container.adapter_data} containerId={container.id} />
      </div>
    {/if}

    <div class="actions">
      {#if container.state === "running"}
        <button class="liquid-btn action-btn restart" disabled={acting} onclick={() => ask("restart")}>
          <svg aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 0 1 15.5-6.2L21 8" /><path d="M21 3v5h-5" /><path d="M21 12a9 9 0 0 1-15.5 6.2L3 16" /><path d="M3 21v-5h5" /></svg>
          <span>Reiniciar</span>
        </button>
        <button class="liquid-btn action-btn danger" disabled={acting} onclick={() => ask("stop")}>
          <svg aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><rect x="6" y="6" width="12" height="12" rx="2" /></svg>
          <span>Detener</span>
        </button>
      {:else}
        <button class="liquid-btn action-btn start" disabled={acting} onclick={() => ask("start")}>
          <svg aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M7 4.5v15l12-7.5z" /></svg>
          <span>Iniciar</span>
        </button>
      {/if}
    </div>
    {#if actionError}
      <p class="action-error">{actionError}</p>
    {/if}

    <section class="logs-section">
      <div class="logs-head">
        <h2>Registro de Logs</h2>
        <input type="search" placeholder="Filtrar logs…" bind:value={search} />
      </div>
      <div class="glass logs" bind:this={logEl} onscroll={onLogScroll}>
        {#each filteredLines as line, i (i)}
          <pre>{line}</pre>
        {/each}
        {#if filteredLines.length === 0}
          <p class="empty">Esperando líneas de logs del contenedor…</p>
        {/if}
      </div>
      {#if !autoscroll}
        <button class="jump-btn" onclick={jumpToBottom}>↓ Ir al final</button>
      {/if}
    </section>
  {:else}
    <div class="glass loading-box">
      <p>Cargando información del contenedor…</p>
    </div>
  {/if}
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title={actionLabels[confirmAction].title}
  message={actionLabels[confirmAction].message}
  confirmLabel={actionLabels[confirmAction].confirmLabel}
  danger={actionLabels[confirmAction].danger}
  onConfirm={() => runAction(confirmAction)}
/>

<style>
  .page {
    max-width: 760px;
    margin: 0 auto;
    padding: calc(24px + var(--safe-top)) 20px calc(110px + var(--safe-bottom));
  }

  .back-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: color-mix(in srgb, var(--c-glass) 12%, transparent);
    border: 1px solid var(--border-strong);
    color: var(--c-action);
    font-size: 13px;
    font-weight: 700;
    padding: 6px 14px;
    border-radius: 99em;
    margin-bottom: 22px;
    cursor: pointer;
    box-shadow: inset 0 1px 0 var(--glass-specular);
    transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast) ease;
  }

  .back-pill:hover {
    transform: translateX(-3px);
    border-color: var(--c-action);
  }

  header {
    margin-bottom: 24px;
    view-transition-name: container-header;
  }

  .header-badge-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
  }

  .comic-chip {
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.08em;
    color: var(--c-accent-secondary);
    background: color-mix(in srgb, var(--c-accent-secondary) 14%, transparent);
    border: 1px solid color-mix(in srgb, var(--c-accent-secondary) 30%, transparent);
    padding: 2px 8px;
    border-radius: 99em;
    text-transform: uppercase;
  }

  .status-pill {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-tertiary);
  }

  .status-pill.running {
    color: var(--c-accent-secondary);
  }

  h1 {
    font-size: 36px;
    font-weight: 800;
    line-height: 1.15;
    word-break: break-word;
    color: var(--text);
  }

  .image {
    color: var(--text-secondary);
    font-size: 13px;
    font-weight: 500;
    margin: 6px 0 2px;
  }

  .meta {
    color: var(--text-tertiary);
    font-size: 12px;
    margin: 0;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 14px;
    margin-bottom: 22px;
  }

  .adapter-panel {
    margin-bottom: 22px;
  }

  .label {
    font-size: 11px;
    letter-spacing: 0.05em;
    font-weight: 700;
    color: var(--text-secondary);
  }

  .value {
    font-family: var(--font-display);
    font-size: 32px;
    font-weight: 800;
    margin: 4px 0 10px;
    color: var(--text);
  }

  .ports {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 10px;
  }

  .port {
    font-size: 12px;
    font-weight: 600;
    padding: 4px 10px;
    border-radius: 10px;
    background: color-mix(in srgb, var(--c-glass) 12%, transparent);
    border: 1px solid var(--border);
    font-variant-numeric: tabular-nums;
  }

  .actions {
    display: flex;
    gap: 12px;
    margin-bottom: 12px;
  }

  .liquid-btn {
    flex: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 48px;
    touch-action: manipulation;
    padding: 13px;
    border-radius: 16px;
    border: 1px solid var(--border-strong);
    background: linear-gradient(135deg, color-mix(in srgb, var(--c-glass) 16%, transparent), color-mix(in srgb, var(--c-glass) 6%, transparent));
    color: var(--text);
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
    box-shadow: inset 0 1px 0 var(--glass-specular);
    transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) ease, border-color var(--dur-fast) ease;
  }

  .liquid-btn:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
  }

  .liquid-btn:active:not(:disabled) {
    transform: scale(0.96);
  }

  .liquid-btn:disabled {
    opacity: 0.4;
    cursor: default;
  }

  .liquid-btn.start:hover:not(:disabled) {
    border-color: var(--c-accent-secondary);
    color: var(--c-accent-secondary);
  }

  .liquid-btn.restart:hover:not(:disabled) {
    border-color: var(--c-action);
    color: var(--c-action);
  }

  .liquid-btn.danger {
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 35%, var(--border-strong));
  }

  .liquid-btn.danger:hover:not(:disabled) {
    border-color: var(--danger);
    box-shadow: 0 6px 20px color-mix(in srgb, var(--danger) 25%, transparent);
  }

  .action-error {
    color: var(--danger);
    font-size: 13px;
    font-weight: 600;
    margin: 8px 0;
  }

  .logs-section {
    margin-top: 28px;
    position: relative;
  }

  .logs-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .logs-head h2 {
    font-size: 22px;
    font-weight: 800;
  }

  .logs-head input {
    padding: 8px 14px;
    border-radius: 12px;
    border: 1px solid var(--border-strong);
    background: color-mix(in srgb, var(--c-glass) 12%, transparent);
    color: var(--text);
    font-size: 13px;
    outline: none;
    width: 170px;
    transition: border-color var(--dur-fast) ease;
  }

  .logs-head input:focus {
    border-color: var(--c-action);
  }

  .logs {
    height: 380px;
    overflow-y: auto;
    padding: 16px;
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.6;
    background: color-mix(in srgb, var(--bg-elevated) 85%, transparent);
  }

  .logs pre {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--text-secondary);
  }

  .empty {
    color: var(--text-tertiary);
    font-size: 13px;
    text-align: center;
    margin-top: 40px;
  }

  .jump-btn {
    position: absolute;
    bottom: 18px;
    right: 18px;
    padding: 8px 16px;
    border-radius: 99em;
    border: none;
    background: linear-gradient(135deg, var(--c-action), var(--c-accent-secondary));
    color: white;
    font-size: 12px;
    font-weight: 700;
    cursor: pointer;
    box-shadow: 0 4px 16px color-mix(in srgb, var(--c-action) 40%, transparent);
    animation: comicBadgePop 240ms var(--ease-out);
    transition: transform var(--dur-fast) var(--ease-out);
  }

  .jump-btn:hover {
    transform: translateY(-2px) scale(1.05);
  }

  .loading-box {
    padding: 30px;
    text-align: center;
    color: var(--text-secondary);
  }
</style>
