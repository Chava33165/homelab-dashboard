<script lang="ts">
  let {
    level = "warning",
    text,
  }: { level?: "warning" | "danger" | "success" | "info"; text: string } = $props();
</script>

<span class="badge {level}">
  <span class="dot"></span>
  <span class="text">{text}</span>
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 3px 10px;
    border-radius: 99em;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.02em;
    text-transform: uppercase;
    border: 1.5px solid currentColor;
    animation: comicBadgePop var(--dur-base) var(--ease-out);
  }

  @keyframes comicBadgePop {
    0% {
      transform: scale(0.7) rotate(-3deg);
      opacity: 0;
    }
    60% {
      transform: scale(1.08) rotate(1deg);
    }
    100% {
      transform: scale(1) rotate(0deg);
      opacity: 1;
    }
  }

  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    animation: pulseDot 1.4s var(--ease-out) infinite;
  }

  @keyframes pulseDot {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(0.65);
      opacity: 0.4;
    }
  }

  .warning {
    background: color-mix(in srgb, var(--warning) 16%, transparent);
    color: var(--warning);
    border-color: color-mix(in srgb, var(--warning) 40%, transparent);
  }
  .warning .dot {
    background: var(--warning);
    box-shadow: 0 0 6px var(--warning);
  }

  .danger {
    background: color-mix(in srgb, var(--danger) 18%, transparent);
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 40%, transparent);
  }
  .danger .dot {
    background: var(--danger);
    box-shadow: 0 0 6px var(--danger);
  }

  .success {
    background: color-mix(in srgb, var(--c-accent-secondary) 18%, transparent);
    color: var(--c-accent-secondary);
    border-color: color-mix(in srgb, var(--c-accent-secondary) 40%, transparent);
  }
  .success .dot {
    background: var(--c-accent-secondary);
    box-shadow: 0 0 6px var(--c-accent-secondary);
  }

  .info {
    background: color-mix(in srgb, var(--c-action) 18%, transparent);
    color: var(--c-action);
    border-color: color-mix(in srgb, var(--c-action) 40%, transparent);
  }
  .info .dot {
    background: var(--c-action);
    box-shadow: 0 0 6px var(--c-action);
  }
</style>
