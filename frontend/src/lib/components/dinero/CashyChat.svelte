<script lang="ts">
  import { onDestroy, onMount, tick } from "svelte";
  import { api } from "../../api/client";
  import type { CashyHistoria } from "../../api/types";
  import CashyCara, { type Gesto } from "./CashyCara.svelte";

  // Chat con Cashy. La respuesta tarda (Claude corre en el worker de KORRA): mientras piensa se
  // consulta GET /cashy cada 2.5 s, y solo mientras este chat esta abierto.
  let historia = $state<CashyHistoria | null>(null);
  let texto = $state("");
  let error = $state<string | null>(null);
  let enviando = $state(false);
  let hablando = $state(false);
  let lista: HTMLDivElement;
  let timer = 0;
  let hablaTimer = 0;
  let ultimoId = 0;

  const SUGERENCIAS = ["¿Cómo voy esta semana?", "¿Puedo gastar $300 en ", "¿Cuánto debo y cuándo?", "¿Cómo ahorro más?"];

  let gesto = $derived<Gesto>(historia?.pensando || enviando ? "pensando" : hablando ? "hablando" : "neutral");

  async function cargar() {
    try {
      const h = await api<CashyHistoria>("/finanzas/cashy");
      const ult = h.mensajes.at(-1);
      if (ult && ult.rol === "cashy" && ult.estado === "ok" && ult.id === ultimoId && historia?.pensando) {
        hablando = true; // acaba de llegar la respuesta: mueve la boca un momento
        clearTimeout(hablaTimer);
        hablaTimer = window.setTimeout(() => (hablando = false), 2200);
      }
      ultimoId = ult?.id ?? 0;
      historia = h;
      error = null;
      await tick();
      lista?.scrollTo({ top: lista.scrollHeight, behavior: "smooth" });
    } catch (e) {
      error = e instanceof Error ? e.message : "No pude hablar con Cashy";
    }
    clearTimeout(timer);
    if (historia?.pensando) timer = window.setTimeout(cargar, 2500);
  }

  async function enviar(mensaje = texto) {
    mensaje = mensaje.trim();
    if (!mensaje || enviando || historia?.pensando) return;
    enviando = true;
    error = null;
    try {
      await api("/finanzas/cashy", { method: "POST", body: JSON.stringify({ mensaje }) });
      texto = "";
      await cargar();
    } catch (e) {
      error = e instanceof Error ? e.message : "No se pudo enviar";
    } finally {
      enviando = false;
    }
  }

  let confirmar = $state(false);
  async function nueva() {
    if (!confirmar) {
      confirmar = true;
      return;
    }
    confirmar = false;
    await api("/finanzas/cashy", { method: "DELETE" });
    await cargar();
  }

  function sugerir(s: string) {
    if (s.endsWith(" ")) texto = s; // la deja para que completes
    else enviar(s);
  }

  // Markdown minimo y seguro: se escapa todo y luego solo **negritas**, listas y saltos de linea.
  function md(t: string): string {
    const esc = t.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
    return esc
      .replace(/\*\*(.+?)\*\*/g, "<b>$1</b>")
      .split("\n")
      .map((l) => (/^\s*[-*•]\s+/.test(l) ? `<span class="li">${l.replace(/^\s*[-*•]\s+/, "")}</span>` : l))
      .join("<br>");
  }

  onMount(cargar);
  onDestroy(() => {
    clearTimeout(timer);
    clearTimeout(hablaTimer);
  });
</script>

<div class="chat">
  <div class="top">
    <div class="avatar"><CashyCara {gesto} size={88} /></div>
    <div>
      <b>Cashy</b>
      <span>{historia?.pensando || enviando ? "pensando…" : "tu asesora de dinero"}</span>
    </div>
  </div>

  <div class="msgs" bind:this={lista} aria-live="polite">
    {#if historia && !historia.mensajes.length}
      <p class="vacio">¡Hola, Chava! Pregúntame lo que quieras de tu dinero: si te alcanza para algo, cómo vas con tus tarjetas o cómo ahorrar más. Veo todos tus datos, pero no muevo nada.</p>
    {/if}
    {#each historia?.mensajes ?? [] as m (m.id)}
      <div class="m {m.rol}" class:err={m.estado === "error"}>
        {#if m.estado === "pensando"}
          <span class="dots" aria-label="Cashy está pensando"><i></i><i></i><i></i></span>
        {:else}
          <!-- eslint-disable-next-line svelte/no-at-html-tags -- md() escapa todo el HTML -->
          {@html md(m.texto)}
        {/if}
      </div>
    {/each}
  </div>

  {#if error}<p class="form-error">{error}</p>{/if}

  <div class="chips sug">
    {#each SUGERENCIAS as s (s)}<button type="button" onclick={() => sugerir(s)} disabled={historia?.pensando}>{s.trim()}…</button>{/each}
  </div>

  <form class="enviar" onsubmit={(e) => { e.preventDefault(); enviar(); }}>
    <input bind:value={texto} maxlength="1000" placeholder="Pregúntale a Cashy…" aria-label="Mensaje para Cashy" />
    <button class="btn primary" disabled={!texto.trim() || enviando || historia?.pensando} aria-label="Enviar">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M13 6l6 6-6 6" /></svg>
    </button>
  </form>

  <div class="pie">
    <span>Usa tu plan de Claude ({historia?.modelo ?? "sonnet"}): cada pregunta gasta un poco de tu límite.</span>
    <button class="link" onclick={nueva}>{confirmar ? "¿Seguro? Toca otra vez" : "Nueva conversación"}</button>
  </div>
</div>

<style>
  .chat {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .top {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .avatar {
    width: 76px;
    height: 76px;
    display: grid;
    place-items: center;
    border-radius: 50%;
    border: 2.5px solid #141317;
    background: radial-gradient(circle at 50% 40%, #ffd3e4, #ff4f9a);
    box-shadow: 3px 3px 0 var(--panel-offset);
    flex-shrink: 0;
  }

  .top b {
    display: block;
    font-family: var(--font-display);
    font-size: 22px;
  }

  .top span {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .msgs {
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-height: min(46dvh, 460px);
    min-height: 120px;
    overflow-y: auto;
    overscroll-behavior: contain;
    padding: 4px 2px;
  }

  .vacio {
    margin: 0;
    font-size: 14px;
    line-height: 1.5;
    color: var(--text-secondary);
  }

  .m {
    max-width: 86%;
    padding: 10px 14px;
    border-radius: 18px;
    font-size: 14px;
    line-height: 1.45;
    overflow-wrap: anywhere;
  }

  .m.chava {
    align-self: flex-end;
    background: var(--c-action);
    color: #fff;
    border-bottom-right-radius: 6px;
  }

  .m.cashy {
    align-self: flex-start;
    background: color-mix(in srgb, var(--text) 8%, transparent);
    border-bottom-left-radius: 6px;
  }

  .m.err {
    border: 1.5px solid var(--danger);
  }

  .m :global(.li) {
    display: block;
    padding-left: 14px;
    position: relative;
  }

  .m :global(.li)::before {
    content: "•";
    position: absolute;
    left: 2px;
  }

  .dots {
    display: inline-flex;
    gap: 4px;
  }

  .dots i {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--text-secondary);
    animation: dot 1s infinite ease-in-out;
  }

  .dots i:nth-child(2) {
    animation-delay: 0.15s;
  }

  .dots i:nth-child(3) {
    animation-delay: 0.3s;
  }

  @keyframes dot {
    0%,
    100% {
      opacity: 0.3;
      transform: translateY(0);
    }
    50% {
      opacity: 1;
      transform: translateY(-3px);
    }
  }

  .sug {
    flex-wrap: nowrap;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .sug button {
    flex-shrink: 0;
    min-height: 34px;
    font-size: 12px;
  }

  .enviar {
    display: flex;
    gap: 8px;
  }

  .enviar input {
    flex: 1;
    min-width: 0;
    min-height: 46px;
    padding: 0 14px;
    border-radius: 16px;
    border: 1.5px solid var(--border-strong);
    background: var(--bg-elevated);
    color: var(--text);
    font: inherit;
    font-size: 16px;
  }

  .enviar input:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 1px;
  }

  .enviar .btn {
    width: 52px;
    padding: 0;
    display: grid;
    place-items: center;
  }

  .pie {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 10px;
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .link {
    border: none;
    background: none;
    padding: 6px 0;
    font: inherit;
    font-size: 12px;
    font-weight: 700;
    color: var(--c-action);
    cursor: pointer;
    white-space: nowrap;
  }

  @media (prefers-reduced-motion: reduce) {
    .dots i {
      animation: none;
    }
  }
</style>
