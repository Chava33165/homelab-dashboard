<script lang="ts">
  import { pending, setPower, isOn, prettyName, stopWarning, type Service } from "../stores/services";
  import ConfirmDialog from "./ConfirmDialog.svelte";

  let { service }: { service: Service } = $props();

  let on = $derived(isOn(service, $pending));
  let busy = $derived(service.project in $pending);
  let confirmOpen = $state(false);

  function toggle() {
    if (busy) return;
    // Apagar pide confirmacion (un toque accidental en el cel tumbaria Immich); prender
    // no hace daño, va directo.
    if (on) confirmOpen = true;
    else setPower(service.project, true);
  }
</script>

<button
  class="switch"
  class:on
  class:busy
  role="switch"
  aria-checked={on}
  aria-label="{on ? 'Apagar' : 'Encender'} {prettyName(service.project)}"
  disabled={busy}
  onclick={toggle}
>
  <span class="knob"></span>
</button>

<ConfirmDialog
  bind:open={confirmOpen}
  title="¿Apagar {prettyName(service.project)}?"
  message={stopWarning(service)}
  confirmLabel="Apagar"
  danger
  onConfirm={() => setPower(service.project, false)}
/>

<style>
  /* 52x32 visible, area tactil ampliada a 48px con ::before. */
  .switch {
    position: relative;
    flex-shrink: 0;
    width: 52px;
    height: 32px;
    padding: 0;
    border: 1.5px solid var(--border-strong);
    border-radius: 99em;
    background: color-mix(in srgb, var(--c-glass) 10%, transparent);
    cursor: pointer;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
    transition:
      background-color var(--dur-base) ease,
      border-color var(--dur-base) ease,
      transform var(--dur-press) var(--ease-out);
  }

  .switch::before {
    content: "";
    position: absolute;
    inset: -9px -6px;
  }

  .switch:active:not(:disabled) {
    transform: scale(0.95);
  }

  .switch.on {
    background: var(--c-action);
    border-color: var(--c-action);
  }

  .switch:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 3px;
  }

  .switch:disabled {
    cursor: progress;
  }

  .knob {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 25px;
    height: 25px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.25), 0 2px 6px rgba(0, 0, 0, 0.2);
    transition: transform var(--dur-slow) var(--ease-spring);
  }

  .switch.on .knob {
    transform: translateX(20px);
  }

  /* Mientras Docker trabaja (varios segundos por contenedor) el knob respira. */
  .switch.busy .knob {
    animation: knobBusy 900ms ease-in-out infinite alternate;
  }

  @keyframes knobBusy {
    to {
      opacity: 0.4;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .switch.busy .knob {
      animation: none;
      opacity: 0.6;
    }
  }
</style>
