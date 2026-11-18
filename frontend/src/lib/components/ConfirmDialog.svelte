<script lang="ts">
  let {
    open = $bindable(false),
    title,
    message,
    confirmLabel = "Confirmar",
    danger = false,
    requireDoubleConfirm = false,
    onConfirm,
  }: {
    open?: boolean;
    title: string;
    message: string;
    confirmLabel?: string;
    danger?: boolean;
    requireDoubleConfirm?: boolean;
    onConfirm: () => void;
  } = $props();

  let armed = $state(false);

  function handleConfirm() {
    if (requireDoubleConfirm && !armed) {
      armed = true;
      return;
    }
    open = false;
    armed = false;
    onConfirm();
  }

  function close() {
    open = false;
    armed = false;
  }
</script>

{#if open}
  <div
    class="backdrop"
    role="button"
    tabindex="0"
    aria-label="Cerrar modal"
    onclick={close}
    onkeydown={(e) => e.key === "Escape" && close()}
  >
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div class="glass dialog" role="dialog" aria-modal="true" tabindex="-1" onclick={(e) => e.stopPropagation()}>
      <div class="dialog-icon" class:danger>
        {#if danger}
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M12 9v4M12 17h.01M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
          </svg>
        {:else}
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <circle cx="12" cy="12" r="10" />
            <path d="M12 16v-4M12 8h.01" />
          </svg>
        {/if}
      </div>

      <h3>{title}</h3>
      <p>{message}</p>

      <div class="actions">
        <button class="liquid-btn cancel-btn" onclick={close}>Cancelar</button>
        <button class="liquid-btn confirm-btn" class:danger onclick={handleConfirm}>
          {armed ? "⚠️ ¿Seguro? Confirma" : confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: var(--scrim);
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding: 16px;
    padding-bottom: calc(16px + var(--safe-bottom));
    z-index: 100;
    animation: fade 200ms var(--ease-out);
  }

  @media (min-width: 640px) {
    .backdrop {
      align-items: center;
    }
  }

  @keyframes fade {
    from { opacity: 0; }
  }

  .dialog {
    width: 100%;
    max-width: 400px;
    padding: 28px 24px;
    border-radius: 28px;
    background: var(--bg-elevated);
    box-shadow:
      5px 5px 0 var(--panel-offset),
      0 20px 50px rgba(0, 0, 0, 0.35);
    animation: comicRise var(--dur-base) var(--ease-out);
    text-align: center;
  }

  @keyframes comicRise {
    from {
      transform: translateY(16px) scale(0.97);
      opacity: 0;
    }
  }

  .dialog-icon {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    margin: 0 auto 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: color-mix(in srgb, var(--c-glass) 16%, transparent);
    border: 1px solid var(--border-strong);
    color: var(--c-action);
  }

  .dialog-icon.danger {
    color: var(--danger);
    background: color-mix(in srgb, var(--danger) 16%, transparent);
    border-color: color-mix(in srgb, var(--danger) 30%, transparent);
  }

  h3 {
    font-size: 22px;
    font-weight: 800;
    margin-bottom: 8px;
  }

  p {
    color: var(--text-secondary);
    font-size: 14px;
    line-height: 1.5;
    margin: 0 0 24px;
  }

  .actions {
    display: flex;
    gap: 12px;
  }

  .liquid-btn {
    flex: 1;
    padding: 13px;
    border-radius: 16px;
    border: 1px solid var(--border-strong);
    background: color-mix(in srgb, var(--c-glass) 12%, transparent);
    color: var(--text);
    font-size: 14px;
    font-weight: 700;
    cursor: pointer;
    box-shadow: inset 0 1px 0 var(--glass-specular);
    transition: transform var(--dur-fast) var(--ease-out), background var(--dur-fast) ease;
  }

  .liquid-btn:hover {
    transform: translateY(-2px);
  }

  .liquid-btn:active {
    transform: scale(0.96);
  }

  .confirm-btn.danger {
    background: linear-gradient(135deg, var(--danger), #ff6b8b);
    border-color: transparent;
    color: white;
    box-shadow: 0 4px 16px color-mix(in srgb, var(--danger) 40%, transparent);
  }
</style>
