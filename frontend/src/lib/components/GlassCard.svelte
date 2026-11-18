<script lang="ts">
  import type { Snippet } from "svelte";

  // Vineta de comic. Con onclick se vuelve un <button> real (teclado, foco, lector de
  // pantalla) y responde al presionar; sin onclick es una superficie estatica -- antes
  // varias tarjetas tenian cursor de "clic" sin hacer nada al tocarlas.
  let {
    padding = "22px",
    comic = false,
    onclick,
    label,
    children,
  }: {
    padding?: string;
    comic?: boolean;
    onclick?: () => void;
    label?: string;
    children: Snippet;
  } = $props();

</script>

{#if onclick}
  <button
    class="glass card pressable"
    class:comic-halftone-pattern={comic}
    style:padding
    {onclick}
    aria-label={label}
  >
    {@render children()}
  </button>
{:else}
  <div class="glass card" class:comic-halftone-pattern={comic} style:padding>
    {@render children()}
  </div>
{/if}

<style>
  .card {
    position: relative;
    display: block;
    width: 100%;
    border-radius: 24px;
    animation: rise var(--dur-slow) var(--ease-out) both;
  }

  .pressable {
    font: inherit;
    color: inherit;
    text-align: left;
    cursor: pointer;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
    transition:
      transform var(--dur-fast) var(--ease-out),
      box-shadow var(--dur-fast) var(--ease-out);
  }

  /* Al presionar la vineta "se hunde" sobre su sombra dura: feedback inmediato en
     pointer-down, como un boton fisico. */
  .pressable:active {
    transform: translate(2px, 2px) scale(0.99);
    box-shadow:
      inset 0 1px 0 var(--glass-specular),
      2px 2px 0 var(--panel-offset);
    transition-duration: var(--dur-press);
  }

  .pressable:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 3px;
  }

  @media (hover: hover) and (pointer: fine) {
    .pressable:hover {
      transform: translate(-1px, -1px);
      box-shadow:
        inset 0 1px 0 var(--glass-specular),
        6px 6px 0 var(--c-action);
    }
  }
</style>
