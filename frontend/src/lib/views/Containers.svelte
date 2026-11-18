<script lang="ts">
  import { onMount } from "svelte";
  import { containersLoaded, refreshContainers, startContainersStream } from "../stores/containers";
  import { services, pending, powerErrors, playersOnline, prettyName, type Service } from "../stores/services";
  import { navigate } from "../stores/router";
  import { formatBytes } from "../format";
  import GlassCard from "../components/GlassCard.svelte";
  import PowerSwitch from "../components/PowerSwitch.svelte";

  onMount(() => {
    startContainersStream();
    refreshContainers().catch(() => void 0);
  });

  let runningServices = $derived($services.filter((s) => s.running > 0).length);
  let expanded = $state<Record<string, boolean>>({});

  // Un servicio de un solo contenedor (ej. Terraria) entra directo a su detalle; uno
  // de varios despliega la lista para elegir cual.
  function openService(s: Service) {
    if (s.members.length === 1) navigate({ name: "container", id: s.members[0].id });
    else expanded[s.project] = !expanded[s.project];
  }

  function statusText(s: Service): string {
    if (s.project in $pending) return $pending[s.project] ? "Encendiendo…" : "Apagando…";
    if (s.running === 0) return "Apagado";
    const mem = s.memBytes > 0 ? ` · ${formatBytes(s.memBytes)}` : "";
    if (s.running < s.members.length) return `${s.running} de ${s.members.length} encendidos${mem}`;
    return `Encendido${mem}`;
  }
</script>

<div class="page">
  <div class="page-head">
    <span class="eyebrow">Tu homelab</span>
    <h1 class="display-title chromatic-pop">Servicios</h1>
    {#if $containersLoaded}
      <span class="ink-chip">{runningServices} de {$services.length} encendidos</span>
    {/if}
  </div>

  {#if !$containersLoaded}
    <div class="loading-box glass">
      <p>Cargando servicios…</p>
    </div>
  {:else}
    <div class="service-list stagger">
      {#each $services as s, i (s.project)}
        {@const partial = !(s.project in $pending) && s.running > 0 && s.running < s.members.length}
        {@const players = playersOnline(s)}
        <GlassCard padding="6px 18px 6px 6px">
          <div class="service-row">
            <button
              class="service-main"
              onclick={() => openService(s)}
              aria-expanded={s.members.length > 1 ? !!expanded[s.project] : undefined}
            >
              <span class="state-dot" class:on={s.running > 0} class:partial></span>
              <span class="service-info">
                <span class="service-name">{prettyName(s.project)}</span>
                {#if players !== undefined && !(s.project in $pending)}
                  <span class="service-status players" class:active={players > 0}>
                    <svg aria-hidden="true" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></svg>
                    {players === 0 ? "Nadie conectado" : `${players} jugando ahora`}
                  </span>
                {:else}
                  <span class="service-status" class:partial>{statusText(s)}</span>
                {/if}
              </span>
              <svg class="chevron" class:open={expanded[s.project]} aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d={s.members.length > 1 ? "M6 9l6 6 6-6" : "M9 6l6 6-6 6"} /></svg>
            </button>
            {#if s.protected}
              <span class="system-tag" title="Aquí corre este dashboard: apagarlo cerraría esta página">SISTEMA</span>
            {:else}
              <PowerSwitch service={s} />
            {/if}
          </div>

          {#if $powerErrors[s.project]}
            <p class="service-error">{$powerErrors[s.project]}</p>
          {/if}

          {#if s.members.length > 1 && expanded[s.project]}
            <div class="members">
              <ul>
                {#each s.members as c (c.id)}
                  <li>
                    <button class="member" onclick={() => navigate({ name: "container", id: c.id })}>
                      <span class="mini-dot" class:on={c.state === "running"}></span>
                      <span class="member-name">{c.name}</span>
                      <span class="member-state">{c.state === "running" ? (c.stats ? formatBytes(c.stats.mem_usage_bytes) : "activo") : c.state}</span>
                      <svg class="member-chevron" aria-hidden="true" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M9 6l6 6-6 6" /></svg>
                    </button>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
        </GlassCard>
      {/each}
    </div>
  {/if}
</div>


<style>
  .loading-box {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    text-align: center;
    color: var(--text-secondary);
  }

  .service-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .service-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  /* Toda la parte izquierda de la tarjeta es el boton: blanco grande para el dedo,
     el switch queda aparte para no mezclar "ver" con "apagar". */
  .service-main {
    flex: 1;
    min-width: 0;
    min-height: 64px;
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 10px 12px;
    border: none;
    border-radius: 16px;
    background: none;
    color: inherit;
    font-family: inherit;
    text-align: left;
    cursor: pointer;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
    transition: background-color 160ms ease, transform 160ms var(--ease-out);
  }

  .service-main:active {
    background: color-mix(in srgb, var(--c-glass) 10%, transparent);
    transform: scale(0.98);
  }

  .service-main:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: -2px;
  }

  @media (hover: hover) and (pointer: fine) {
    .service-main:hover {
      background: color-mix(in srgb, var(--c-glass) 8%, transparent);
    }
    .service-main:hover .chevron {
      color: var(--c-action);
    }
  }

  .service-status.players {
    display: flex;
    align-items: center;
    gap: 5px;
    font-weight: 600;
  }

  .service-status.players.active {
    color: var(--c-action);
  }

  .chevron {
    flex-shrink: 0;
    color: var(--text-tertiary);
    transition: transform 200ms var(--ease-out), color 160ms ease;
  }

  .chevron.open {
    transform: rotate(180deg);
  }

  .state-dot {
    width: 10px;
    height: 10px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--text-tertiary);
    transition: background-color 200ms ease;
  }

  .state-dot.on {
    background: var(--c-action);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--c-action) 22%, transparent);
  }

  /* A medias: anillo sin relleno (sin meter otro color a la paleta). */
  .state-dot.partial {
    background: transparent;
    box-shadow: inset 0 0 0 2px var(--c-action);
  }

  .service-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .service-name {
    font-weight: 700;
    font-size: 17px;
    letter-spacing: -0.02em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .service-status {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
    color: var(--text-secondary);
    font-variant-numeric: tabular-nums;
    margin-top: 1px;
  }

  .service-status.partial {
    color: var(--c-action);
    font-weight: 600;
  }

  .system-tag {
    flex-shrink: 0;
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.06em;
    padding: 4px 9px;
    border-radius: 99em;
    color: var(--text-tertiary);
    border: 1px solid var(--border);
  }

  .service-error {
    margin: 0 0 10px 38px;
    font-size: 13px;
    color: var(--danger);
  }

  .members ul {
    list-style: none;
    margin: 0 0 8px 30px;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .member {
    width: 100%;
    min-height: 44px;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    touch-action: manipulation;
    background: none;
    border: none;
    border-radius: 10px;
    color: inherit;
    font-family: inherit;
    text-align: left;
    cursor: pointer;
    transition: background-color 160ms ease;
  }

  .member:active,
  .member:hover {
    background: color-mix(in srgb, var(--c-glass) 10%, transparent);
  }

  .mini-dot {
    width: 7px;
    height: 7px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--text-tertiary);
  }

  .mini-dot.on {
    background: var(--c-action);
  }

  .member-name {
    flex: 1;
    min-width: 0;
    font-size: 13px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .member-state {
    font-size: 11px;
    color: var(--text-tertiary);
    font-variant-numeric: tabular-nums;
  }

  .member-chevron {
    flex-shrink: 0;
    color: var(--text-tertiary);
  }

  @media (prefers-reduced-motion: reduce) {
    .chevron,
    .service-main {
      transition: none;
    }
  }
</style>
