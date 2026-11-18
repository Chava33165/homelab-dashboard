<script lang="ts">
  import type { FinTarjeta } from "../../api/types";
  import { accion, fecha, pesos, resumen } from "../../stores/finanzas";

  let { tarjeta, onDone }: { tarjeta: FinTarjeta; onDone: (mensaje: string) => void } = $props();

  const fondo = $resumen?.fondo ?? 0;
  // svelte-ignore state_referenced_locally -- solo valor inicial sugerido
  let monto = $state<number | null>(Math.min(tarjeta.a_pagar || tarjeta.deuda, fondo) || null);
  let error = $state<string | null>(null);
  let enviando = $state(false);

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    enviando = true;
    error = null;
    try {
      const r = await accion<{ mensaje: string }>("POST", "pagos-tarjeta", { tarjeta: tarjeta.id, monto });
      onDone(r.mensaje);
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo registrar";
    } finally {
      enviando = false;
    }
  }
</script>

<form onsubmit={guardar}>
  <dl class="facts">
    <div><dt>Del corte ({fecha(tarjeta.ultimo_corte)})</dt><dd>{pesos(tarjeta.a_pagar)}</dd></div>
    <div><dt>Deuda total</dt><dd>{pesos(tarjeta.deuda)}</dd></div>
    <div><dt>En tu fondo</dt><dd class:low={fondo < tarjeta.a_pagar}>{pesos(fondo)}</dd></div>
  </dl>
  <label class="field">
    Cuánto pagaste
    <input class="money" type="number" inputmode="decimal" min="0.01" step="0.01" required bind:value={monto} />
  </label>
  <p class="hint">Sale de tu fondo de tarjetas. Tu ahorro no se toca.</p>
  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando || !monto}>{enviando ? "Guardando…" : `Pagar ${tarjeta.nombre}`}</button>
</form>

<style>
  .facts {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
    margin: 0 0 16px;
  }
  .facts div {
    padding: 10px;
    border-radius: 14px;
    background: color-mix(in srgb, var(--text) 5%, transparent);
  }
  dt {
    font-size: 11px;
    color: var(--text-secondary);
  }
  dd {
    margin: 2px 0 0;
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 800;
    font-variant-numeric: tabular-nums;
  }
  dd.low {
    color: var(--c-action);
  }
  .hint {
    margin: -6px 0 14px;
    font-size: 12px;
    color: var(--text-tertiary);
  }
</style>
