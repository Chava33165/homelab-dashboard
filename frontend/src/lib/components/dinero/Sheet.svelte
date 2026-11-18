<script lang="ts">
  import type { Snippet } from "svelte";

  // Formulario modal sobre <dialog> nativo: el navegador ya atrapa el foco, cierra con Esc y
  // oscurece el fondo. Hoja inferior en el cel, panel centrado en pantallas anchas.
  let {
    open = $bindable(false),
    title,
    children,
  }: { open?: boolean; title: string; children: Snippet } = $props();

  let dialog: HTMLDialogElement;

  $effect(() => {
    if (open && !dialog.open) dialog.showModal();
    else if (!open && dialog.open) dialog.close();
  });
</script>

<dialog
  bind:this={dialog}
  class="sheet glass"
  aria-label={title}
  onclose={() => (open = false)}
  onclick={(e) => e.target === dialog && (open = false)}
>
  <div class="sheet-body">
    <header>
      <h2>{title}</h2>
      <button class="x" onclick={() => (open = false)} aria-label="Cerrar">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg>
      </button>
    </header>
    {#if open}
      {@render children()}
    {/if}
  </div>
</dialog>

<style>
  .sheet {
    width: 100%;
    max-width: 520px;
    max-height: min(88dvh, 760px);
    margin: auto auto 0;
    padding: 0;
    border-radius: 26px 26px 0 0;
    background: var(--bg-elevated);
    color: var(--text);
    overflow: auto;
    overscroll-behavior: contain;
  }

  .sheet[open] {
    animation: sheet-in var(--dur-slow) var(--ease-drawer);
  }

  .sheet::backdrop {
    background: var(--scrim);
  }

  .sheet-body {
    padding: 18px 20px calc(22px + var(--safe-bottom));
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
  }

  h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 21px;
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .x {
    width: 36px;
    height: 36px;
    display: grid;
    place-items: center;
    border: none;
    border-radius: 50%;
    background: color-mix(in srgb, var(--text) 8%, transparent);
    color: var(--text-secondary);
    cursor: pointer;
  }

  @keyframes sheet-in {
    from {
      transform: translateY(40px);
      opacity: 0;
    }
  }

  @media (min-width: 640px) {
    .sheet {
      margin: auto;
      border-radius: 26px;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .sheet[open] {
      animation: none;
    }
  }
</style>
