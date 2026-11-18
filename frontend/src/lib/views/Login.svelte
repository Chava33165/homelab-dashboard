<script lang="ts">
  import { login } from "../stores/auth";
  import { ApiError } from "../api/client";
  import GlassCard from "../components/GlassCard.svelte";

  let password = $state("");
  let error = $state("");
  let loading = $state(false);

  async function submit(e: Event) {
    e.preventDefault();
    if (!password || loading) return;
    loading = true;
    error = "";
    try {
      await login(password);
    } catch (err) {
      if (err instanceof ApiError && err.status === 429) {
        error = "Demasiados intentos. Espera un minuto.";
      } else {
        error = "Contraseña incorrecta.";
      }
      password = "";
    } finally {
      loading = false;
    }
  }
</script>

<div class="wrap">
  <div class="ambient-spot ambient-1"></div>
  <div class="ambient-spot ambient-2"></div>

  <GlassCard padding="40px 32px" comic>
    <div class="spider-badge">
      <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 2v20M4 7l8 5 8-5M4 17l8-5 8 5M2 12h20" opacity="0.6"/>
        <circle cx="12" cy="12" r="3.5" fill="var(--c-action)" fill-opacity="0.3"/>
      </svg>
    </div>

    <h1 class="chromatic-pop">Homelab</h1>
    <p class="sub">Raspberry Pi 5 · Acceso Seguro</p>

    <form onsubmit={submit}>
      <div class="input-wrap">
        <input
          type="password"
          placeholder="Introduce la contraseña"
          autocomplete="current-password"
          bind:value={password}
          disabled={loading}
        />
      </div>

      {#if error}
        <p class="error">{error}</p>
      {/if}

      <button type="submit" class="submit-btn" disabled={loading || !password}>
        {#if loading}
          <span class="spin">⚡</span> Entrando…
        {:else}
          <span>Entrar al Panel →</span>
        {/if}
      </button>
    </form>
  </GlassCard>
</div>

<style>
  .wrap {
    min-height: 100dvh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    padding-top: calc(24px + var(--safe-top));
    padding-bottom: calc(24px + var(--safe-bottom));
    position: relative;
    overflow: hidden;
  }

  .ambient-spot {
    position: absolute;
    border-radius: 50%;
    filter: blur(80px);
    pointer-events: none;
    opacity: 0.35;
  }

  .ambient-1 {
    width: 320px;
    height: 320px;
    background: var(--c-action);
    top: 20%;
    left: 15%;
  }

  .ambient-2 {
    width: 280px;
    height: 280px;
    background: var(--c-accent-secondary);
    bottom: 20%;
    right: 15%;
  }

  :global(.wrap .card) {
    width: 100%;
    max-width: 380px;
    text-align: center;
    z-index: 2;
    box-shadow: 
      inset 0 1px 0 var(--glass-specular),
      0 16px 48px var(--glass-shadow),
      0 0 32px color-mix(in srgb, var(--c-action) 12%, transparent);
  }

  .spider-badge {
    width: 52px;
    height: 52px;
    margin: 0 auto 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: color-mix(in srgb, var(--c-glass) 16%, transparent);
    border: 1px solid var(--border-strong);
    color: var(--c-action);
    box-shadow: 
      inset 0 1px 0 var(--glass-specular),
      0 0 16px color-mix(in srgb, var(--c-action) 30%, transparent);
    animation: comicBadgePop 400ms var(--ease-out);
  }

  h1 {
    font-size: 42px;
    font-weight: 800;
    line-height: 1.1;
    color: var(--text);
  }

  .sub {
    color: var(--text-secondary);
    font-size: 13px;
    font-weight: 500;
    margin: 6px 0 28px;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .input-wrap {
    position: relative;
  }

  input {
    width: 100%;
    box-sizing: border-box;
    padding: 15px 18px;
    border-radius: 16px;
    border: 1px solid var(--border-strong);
    background: color-mix(in srgb, var(--c-glass) 10%, transparent);
    font-size: 15px;
    text-align: center;
    outline: none;
    color: var(--text);
    box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.1);
    transition: border-color var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out);
  }

  input:focus {
    border-color: var(--c-action);
    box-shadow: 
      0 0 0 3px color-mix(in srgb, var(--c-action) 25%, transparent),
      inset 0 1px 2px rgba(0, 0, 0, 0.1);
  }

  .error {
    color: var(--danger);
    font-size: 13px;
    font-weight: 600;
    margin: 0;
    animation: shake 320ms var(--ease-spring);
  }

  @keyframes shake {
    0%, 100% { transform: translateX(0); }
    25% { transform: translateX(-6px); }
    75% { transform: translateX(6px); }
  }

  .submit-btn {
    padding: 15px;
    border-radius: 16px;
    border: none;
    background: linear-gradient(135deg, var(--c-action), color-mix(in srgb, var(--c-action) 80%, var(--c-accent-secondary)));
    color: white;
    font-size: 15px;
    font-weight: 700;
    cursor: pointer;
    box-shadow: 
      0 4px 16px color-mix(in srgb, var(--c-action) 40%, transparent),
      inset 0 1px 0 rgba(255, 255, 255, 0.35);
    transition: transform var(--dur-fast) var(--ease-out), opacity var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) ease;
  }

  .submit-btn:hover:not(:disabled) {
    transform: translateY(-2px) scale(1.02);
    box-shadow: 
      0 8px 24px color-mix(in srgb, var(--c-action) 55%, transparent),
      inset 0 1px 0 rgba(255, 255, 255, 0.45);
  }

  .submit-btn:active:not(:disabled) {
    transform: scale(0.97);
  }

  .submit-btn:disabled {
    opacity: 0.5;
    cursor: default;
    box-shadow: none;
  }
</style>
