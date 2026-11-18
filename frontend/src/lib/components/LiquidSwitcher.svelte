<script lang="ts">
  import { theme, type Theme } from "../stores/theme";

  // Control segmentado de 2 opciones. La pastilla es UN elemento que se traslada (solo
  // transform, se compone en GPU), no un fondo por boton.
  const options: { value: Theme; label: string }[] = [
    { value: "dark", label: "Tinta" },
    { value: "light", label: "Papel" },
  ];
</script>

<div class="segmented" role="radiogroup" aria-label="Tema" data-active={$theme}>
  <span class="thumb" aria-hidden="true"></span>
  {#each options as opt (opt.value)}
    <button
      class="seg"
      role="radio"
      aria-checked={$theme === opt.value}
      onclick={() => theme.set(opt.value)}
    >
      {#if opt.value === "dark"}
        <svg aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" /></svg>
      {:else}
        <svg aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" /></svg>
      {/if}
      <span>{opt.label}</span>
    </button>
  {/each}
</div>

<style>
  .segmented {
    position: relative;
    display: grid;
    grid-template-columns: 1fr 1fr;
    width: 240px;
    max-width: 100%;
    margin: 8px auto;
    padding: 4px;
    border-radius: 99em;
    background: color-mix(in srgb, var(--c-glass) 8%, transparent);
    border: 1.5px solid var(--border-strong);
  }

  .thumb {
    position: absolute;
    top: 4px;
    bottom: 4px;
    left: 4px;
    width: calc(50% - 4px);
    border-radius: 99em;
    background: var(--text);
    box-shadow: 2px 2px 0 var(--c-action);
    transition: transform var(--dur-slow) var(--ease-out);
  }

  .segmented[data-active="light"] .thumb {
    transform: translateX(100%);
  }

  .seg {
    position: relative;
    z-index: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-height: 40px;
    border: none;
    background: none;
    border-radius: 99em;
    font-size: 13px;
    font-weight: 700;
    color: var(--text-secondary);
    cursor: pointer;
    touch-action: manipulation;
    transition: color var(--dur-base) ease, transform var(--dur-press) var(--ease-out);
  }

  .seg[aria-checked="true"] {
    color: var(--bg);
  }

  .seg:active {
    transform: scale(0.96);
  }

  .seg:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 2px;
  }
</style>
