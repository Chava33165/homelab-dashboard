<script lang="ts">
  import { route, navigate } from "../stores/router";

  let current = $derived($route.name);
  let activeKey = $derived<"dashboard" | "containers" | "dinero" | "korra" | "settings">(
    current === "container" ? "containers" : current === "dashboard" ? "dashboard" : current
  );
</script>

<nav class="glass tabbar" data-active={activeKey}>

  <button class="tab" class:active={activeKey === "dashboard"} aria-current={activeKey === "dashboard" ? "page" : undefined} onclick={() => navigate({ name: "dashboard" })} aria-label="Panel">
    <div class="icon-wrap">
      <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="3" width="7" height="9" rx="2" />
        <rect x="14" y="3" width="7" height="5" rx="2" />
        <rect x="14" y="12" width="7" height="9" rx="2" />
        <rect x="3" y="16" width="7" height="5" rx="2" />
      </svg>
    </div>
    <span>Panel</span>
  </button>

  <button class="tab" class:active={activeKey === "containers"} aria-current={activeKey === "containers" ? "page" : undefined} onclick={() => navigate({ name: "containers" })} aria-label="Servicios">
    <div class="icon-wrap">
      <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="4" width="18" height="7" rx="2" />
        <rect x="3" y="13" width="18" height="7" rx="2" />
        <path d="M7 7.5h.01M7 16.5h.01" stroke-width="3" />
      </svg>
    </div>
    <span>Servicios</span>
  </button>

  <button class="tab" class:active={activeKey === "dinero"} aria-current={activeKey === "dinero" ? "page" : undefined} onclick={() => navigate({ name: "dinero" })} aria-label="Dinero">
    <div class="icon-wrap">
      <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="2.5" y="6" width="19" height="12" rx="2.5" />
        <circle cx="12" cy="12" r="2.6" />
        <path d="M6 9.5v.01M18 14.5v.01" stroke-width="2.6" />
      </svg>
    </div>
    <span>Dinero</span>
  </button>

  <button class="tab" class:active={activeKey === "korra"} aria-current={activeKey === "korra" ? "page" : undefined} onclick={() => navigate({ name: "korra" })} aria-label="KORRA">
    <div class="icon-wrap">
      <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 3l1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9z" />
        <path d="M19 15l.8 2.2L22 18l-2.2.8L19 21l-.8-2.2L16 18l2.2-.8z" />
      </svg>
    </div>
    <span>KORRA</span>
  </button>

  <button class="tab" class:active={activeKey === "settings"} aria-current={activeKey === "settings" ? "page" : undefined} onclick={() => navigate({ name: "settings" })} aria-label="Ajustes">
    <div class="icon-wrap">
      <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="3.2" />
        <path
          d="M19.4 13a7.6 7.6 0 0 0 0-2l2-1.5-2-3.4-2.3 1a7.6 7.6 0 0 0-1.7-1L15 3.5h-4L10.6 6a7.6 7.6 0 0 0-1.7 1l-2.3-1-2 3.4L6.6 11a7.6 7.6 0 0 0 0 2l-2 1.5 2 3.4 2.3-1a7.6 7.6 0 0 0 1.7 1l.4 2.5h4l.4-2.5a7.6 7.6 0 0 0 1.7-1l2.3 1 2-3.4z"
        />
      </svg>
    </div>
    <span>Ajustes</span>
  </button>
</nav>

<style>
  .tabbar {
    position: fixed;
    left: 50%;
    bottom: calc(16px + var(--safe-bottom));
    transform: translateX(-50%);
    display: flex;
    gap: 6px;
    padding: 6px;
    border-radius: 99em;
    z-index: 40;
    /* Unico desenfoque de la pagina (lo que flota sobre el contenido). */
    backdrop-filter: blur(18px) saturate(140%);
    -webkit-backdrop-filter: blur(18px) saturate(140%);
    background: var(--float-bg);
    border: 1.5px solid var(--glass-border);
    box-shadow:
      3px 3px 0 var(--panel-offset),
      0 12px 36px var(--glass-shadow);
    view-transition-name: navbar;
  }

  /* Los botones comparten un ancho fijo para que la barra quede pareja. */
  .tabbar {
    --tab-w: 64px; /* 5 pestañas caben en 358 px (390 - 2x16 de margen) */
  }

  .tab {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    width: var(--tab-w);
    padding: 8px 6px;
    border: none;
    background: transparent;
    border-radius: 99em;
    color: var(--text-secondary);
    cursor: pointer;
    user-select: none;
    transition:
      color var(--dur-fast) var(--ease-out),
      transform var(--dur-press) var(--ease-out);
    touch-action: manipulation;
  }

  @media (hover: hover) and (pointer: fine) {
    .tab:not(.active):hover {
      color: var(--text);
    }
  }

  .tab:active {
    transform: scale(0.94);
  }

  .tab span {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0;
  }

  /* Pestaña activa: ícono y texto en rosa + rayita de tinta abajo, sin "gota". */
  .tab.active {
    color: var(--c-action);
  }

  .tab::after {
    content: "";
    position: absolute;
    left: 50%;
    bottom: 3px;
    width: 18px;
    height: 3px;
    margin-left: -9px;
    border-radius: 99em;
    background: var(--c-action);
    transform: scaleX(0);
    transition: transform var(--dur-base) var(--ease-out);
  }

  .tab.active::after {
    transform: scaleX(1);
  }

  @media (prefers-reduced-transparency: reduce) {
    .tabbar {
      backdrop-filter: none;
      -webkit-backdrop-filter: none;
      background: var(--bg-elevated);
    }
  }

</style>
