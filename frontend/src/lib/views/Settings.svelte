<script lang="ts">
  import { theme, type Theme } from "../stores/theme";
  import { logout } from "../stores/auth";
  import { latestMetrics } from "../stores/metrics";
  import { api, ApiError } from "../api/client";
  import GlassCard from "../components/GlassCard.svelte";
  import LiquidSwitcher from "../components/LiquidSwitcher.svelte";
  import ConfirmDialog from "../components/ConfirmDialog.svelte";
  import AlertBadge from "../components/AlertBadge.svelte";
  import CashyCara from "../components/dinero/CashyCara.svelte";

  // Cashy (pestaña Dinero): que modelo de Claude usa. Se guarda en task-hub.
  const MODELOS = [
    { id: "sonnet", nombre: "Sonnet", desc: "Recomendado: buenos consejos, gasto moderado de tu límite." },
    { id: "haiku", nombre: "Haiku", desc: "Rápido y casi no gasta límite; razona menos." },
    { id: "opus", nombre: "Opus", desc: "El más listo, pero el que más límite gasta." },
  ];
  let cashyModelo = $state<string | null>(null);
  let cashyError = $state<string | null>(null);
  let cashyBorrada = $state(false);
  let confirmarCashy = $state(false);

  api<{ modelo: string }>("/finanzas/cashy/config")
    .then((c) => (cashyModelo = c.modelo))
    .catch((e) => (cashyError = e instanceof Error ? e.message : "No se pudo leer la configuración de Cashy"));

  async function ponerModelo(m: string) {
    const antes = cashyModelo;
    cashyModelo = m;
    cashyError = null;
    try {
      await api("/finanzas/cashy/config", { method: "PUT", body: JSON.stringify({ modelo: m }) });
    } catch (e) {
      cashyModelo = antes;
      cashyError = e instanceof Error ? e.message : "No se pudo guardar";
    }
  }

  async function borrarCashy() {
    await api("/finanzas/cashy", { method: "DELETE" });
    cashyBorrada = true;
  }

  let m = $derived($latestMetrics);
  let throttled = $derived(!!(m?.throttle.available && m.throttle.raw !== "0x0"));

  type HostAction = "reboot" | "poweroff" | null;
  let confirmAction = $state<HostAction>(null);
  let acting = $state(false);
  let hostError = $state("");
  let hostDone = $state(false);

  async function runHostAction(action: "reboot" | "poweroff") {
    acting = true;
    hostError = "";
    try {
      await api(`/host/${action}`, { method: "POST" });
      hostDone = true;
    } catch (err) {
      hostError = err instanceof ApiError ? err.message : "no se pudo contactar a hostctl";
    } finally {
      acting = false;
    }
  }

  let currentPassword = $state("");
  let newPassword = $state("");
  let confirmPassword = $state("");
  let pwdSaving = $state(false);
  let pwdError = $state("");
  let pwdDone = $state(false);

  async function changePassword(e: SubmitEvent) {
    e.preventDefault();
    pwdError = "";
    pwdDone = false;

    if (newPassword.length < 8) {
      pwdError = "la contraseña nueva debe tener al menos 8 caracteres";
      return;
    }
    if (newPassword !== confirmPassword) {
      pwdError = "la confirmación no coincide con la contraseña nueva";
      return;
    }

    pwdSaving = true;
    try {
      await api("/auth/change-password", {
        method: "POST",
        body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
      });
      pwdDone = true;
      currentPassword = "";
      newPassword = "";
      confirmPassword = "";
    } catch (err) {
      pwdError = err instanceof ApiError ? err.message : "no se pudo cambiar la contraseña";
    } finally {
      pwdSaving = false;
    }
  }
</script>

<div class="page">
  <div class="head-title">
    <span class="comic-chip">SISTEMA</span>
    <h1 class="chromatic-pop">Ajustes</h1>
  </div>

  <GlassCard comic>
    <div class="card-head">
      <span class="label">ESTILO Y APARIENCIA</span>
      <span class="status-pill">LIQUID GLASS</span>
    </div>
    <p class="desc">Tinta (negro) o Papel (blanco), los dos con el rosa de Gwen.</p>
    <LiquidSwitcher />
  </GlassCard>

  <GlassCard comic>
    <div class="card-head">
      <span class="label">CASHY</span>
      <span class="cashy-face"><CashyCara size={46} gesto="feliz" /></span>
    </div>
    <p class="desc">Tu asesora de dinero en la pestaña Dinero. Usa tu plan de Claude: elige qué modelo responde.</p>
    {#if cashyModelo}
      <div class="modelos" role="radiogroup" aria-label="Modelo de Cashy">
        {#each MODELOS as m (m.id)}
          <button role="radio" aria-checked={cashyModelo === m.id} onclick={() => ponerModelo(m.id)}>
            <b>{m.nombre}</b><span>{m.desc}</span>
          </button>
        {/each}
      </div>
    {/if}
    {#if cashyError}<p class="host-error">{cashyError}</p>{/if}
    <button type="button" class="liquid-action-btn" onclick={() => (confirmarCashy = true)}>
      <span>{cashyBorrada ? "Conversación borrada ✓" : "Borrar conversación con Cashy"}</span>
    </button>
  </GlassCard>

  <ConfirmDialog bind:open={confirmarCashy} title="¿Borrar la conversación?" message="Cashy olvida lo que han platicado y empieza una sesión nueva. Tus datos de dinero no se tocan." confirmLabel="Borrar" danger onConfirm={borrarCashy} />

  <GlassCard comic>
    <div class="card-head">
      <span class="label">CONTRASEÑA</span>
      <span class="status-pill ok">HASH LOCAL</span>
    </div>
    <p class="desc">Cámbiala cuando quieras. Se hashea aquí mismo en el servidor y se guarda de una vez — nunca viaja ni queda guardada en texto plano.</p>

    {#if pwdDone}
      <div class="glass host-done-box">
        <p class="note">Contraseña actualizada. Tu sesión actual sigue activa.</p>
      </div>
    {/if}

    <form class="pwd-form" onsubmit={changePassword}>
      <input class="pwd-input" type="password" autocomplete="current-password" placeholder="Contraseña actual" bind:value={currentPassword} required />
      <input class="pwd-input" type="password" autocomplete="new-password" placeholder="Contraseña nueva (mín. 8 caracteres)" bind:value={newPassword} required minlength="8" />
      <input class="pwd-input" type="password" autocomplete="new-password" placeholder="Confirma la contraseña nueva" bind:value={confirmPassword} required minlength="8" />
      {#if pwdError}
        <p class="host-error">{pwdError}</p>
      {/if}
      <button type="submit" class="liquid-action-btn" disabled={pwdSaving}>
        <span>{pwdSaving ? "Guardando…" : "Actualizar contraseña"}</span>
      </button>
    </form>
  </GlassCard>

  <GlassCard comic>
    <div class="card-head">
      <span class="label">CONTROL DEL HOST</span>
      {#if throttled}
        <AlertBadge level="danger" text="Throttling / bajo voltaje" />
      {:else}
        <span class="status-pill ok">RPi 5 OK</span>
      {/if}
    </div>

    {#if hostDone}
      <div class="glass host-done-box">
        <p class="note">Comando enviado con éxito. La Pi debería {confirmAction === "poweroff" ? "apagarse" : "reiniciarse"} en unos segundos — este panel se va a desconectar.</p>
      </div>
    {:else}
      <p class="desc">Requiere que <code>hostctl</code> esté activo en el host vía socket activation. Las órdenes se envían mediante RPC seguro sin conceder privilegios root al contenedor.</p>
      <div class="host-actions">
        <button class="liquid-action-btn" disabled={acting} onclick={() => (confirmAction = "reboot")}>
          <span>🔄 Reiniciar Pi</span>
        </button>
        <button class="liquid-action-btn danger" disabled={acting} onclick={() => (confirmAction = "poweroff")}>
          <span>⏻ Apagar Pi</span>
        </button>
      </div>
      {#if hostError}
        <p class="host-error">{hostError}</p>
      {/if}
    {/if}
  </GlassCard>

  <button class="liquid-logout-btn" onclick={() => logout()}>
    <span>Cerrar sesión</span>
  </button>
</div>

<ConfirmDialog
  open={confirmAction === "reboot"}
  title="¿Reiniciar la Raspberry Pi?"
  message="Todos los servicios (incluido este panel) se van a caer unos segundos mientras reinicia el sistema."
  confirmLabel="Reiniciar"
  danger={true}
  requireDoubleConfirm={true}
  onConfirm={() => runHostAction("reboot")}
/>
<ConfirmDialog
  open={confirmAction === "poweroff"}
  title="¿Apagar la Raspberry Pi?"
  message="Se apaga por completo. Vas a necesitar acceso físico (o un enchufe inteligente) para encenderla de nuevo — este panel no puede encenderla a distancia."
  confirmLabel="Apagar"
  danger={true}
  requireDoubleConfirm={true}
  onConfirm={() => runHostAction("poweroff")}
/>

<style>
  .page {
    max-width: 520px;
    margin: 0 auto;
    padding: calc(24px + var(--safe-top)) 20px calc(110px + var(--safe-bottom));
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .head-title {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 4px;
  }

  .comic-chip {
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.08em;
    color: var(--c-action);
    text-transform: uppercase;
  }

  h1 {
    font-size: 40px;
    font-weight: 800;
    color: var(--text);
  }

  .card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .label {
    font-size: 11px;
    letter-spacing: 0.05em;
    color: var(--text-secondary);
    font-weight: 700;
  }

  .status-pill {
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.05em;
    padding: 2px 8px;
    border-radius: 99em;
    color: var(--c-action);
    background: color-mix(in srgb, var(--c-action) 12%, transparent);
    border: 1px solid color-mix(in srgb, var(--c-action) 30%, transparent);
  }

  .status-pill.ok {
    color: var(--c-accent-secondary);
    background: color-mix(in srgb, var(--c-accent-secondary) 12%, transparent);
    border-color: color-mix(in srgb, var(--c-accent-secondary) 30%, transparent);
  }

  .desc {
    font-size: 13px;
    color: var(--text-secondary);
    margin: 0 0 14px;
    line-height: 1.5;
  }

  code {
    font-family: var(--font-mono);
    background: color-mix(in srgb, var(--border) 80%, transparent);
    padding: 2px 6px;
    border-radius: 6px;
    font-size: 12px;
  }

  .cashy-face {
    width: 44px;
    height: 44px;
    display: grid;
    place-items: center;
    border: 2px solid #141317;
    border-radius: 50%;
    background: radial-gradient(circle at 50% 40%, #ffd3e4, #ff4f9a);
  }

  .modelos {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 14px;
  }

  .modelos button {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
    padding: 12px 14px;
    border-radius: 16px;
    border: 1.5px solid var(--border-strong);
    background: none;
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .modelos button[aria-checked="true"] {
    border-color: var(--c-action);
    background: color-mix(in srgb, var(--c-action) 12%, transparent);
  }

  .modelos b {
    font-size: 15px;
  }

  .modelos span {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .modelos button:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 2px;
  }

  .pwd-form {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-top: 4px;
  }

  .pwd-input {
    padding: 12px 14px;
    border-radius: 14px;
    border: 1px solid var(--border);
    background: color-mix(in srgb, var(--c-glass) 8%, transparent);
    color: var(--text);
    font-size: 14px;
    outline: none;
    transition: border-color var(--dur-fast) ease, box-shadow var(--dur-fast) ease;
  }

  .pwd-input::placeholder {
    color: var(--text-tertiary);
  }

  .pwd-input:focus {
    border-color: var(--c-action);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--c-action) 18%, transparent);
  }

  .host-actions {
    display: flex;
    gap: 12px;
    margin-top: 16px;
  }

  .liquid-action-btn {
    flex: 1;
    padding: 13px;
    border-radius: 16px;
    border: 1px solid var(--border-strong);
    background: linear-gradient(135deg, color-mix(in srgb, var(--c-glass) 16%, transparent), color-mix(in srgb, var(--c-glass) 6%, transparent));
    color: var(--text);
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
    box-shadow: 
      inset 0 1px 0 var(--glass-specular),
      0 4px 12px var(--glass-shadow);
    transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) ease, border-color var(--dur-fast) ease;
  }

  .liquid-action-btn:hover:not(:disabled) {
    transform: translateY(-2px);
    border-color: var(--c-accent-secondary);
    box-shadow: 0 6px 20px color-mix(in srgb, var(--c-accent-secondary) 25%, transparent);
  }

  .liquid-action-btn:active:not(:disabled) {
    transform: scale(0.96);
  }

  .liquid-action-btn:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .liquid-action-btn.danger {
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 35%, var(--border-strong));
  }

  .liquid-action-btn.danger:hover:not(:disabled) {
    border-color: var(--danger);
    box-shadow: 0 6px 20px color-mix(in srgb, var(--danger) 25%, transparent);
  }

  .host-error {
    color: var(--danger);
    font-size: 13px;
    margin: 12px 0 0;
    font-weight: 600;
  }

  .host-done-box {
    padding: 16px;
    margin-top: 12px;
  }

  .liquid-logout-btn {
    padding: 14px;
    border-radius: 16px;
    border: 1px solid color-mix(in srgb, var(--danger) 40%, var(--border-strong));
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    color: var(--danger);
    font-size: 14px;
    font-weight: 700;
    cursor: pointer;
    margin-top: 4px;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.15);
    transition: transform var(--dur-fast) var(--ease-out), background var(--dur-fast) ease;
  }

  .liquid-logout-btn:hover {
    background: color-mix(in srgb, var(--danger) 16%, transparent);
    transform: translateY(-2px);
  }

  .liquid-logout-btn:active {
    transform: scale(0.97);
  }
</style>
