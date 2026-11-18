<script lang="ts">
  import type { FinTarjeta } from "../../api/types";
  import { accion, fecha } from "../../stores/finanzas";

  // "Mi app dice…": lo mas facil para que la tarjeta cuadre con el banco. Pones los 2 numeros
  // que te muestra la app y el sistema mete la diferencia como ajuste.
  let { tarjeta, onDone }: { tarjeta: FinTarjeta; onDone: (mensaje: string) => void } = $props();

  // svelte-ignore state_referenced_locally -- valores iniciales = lo que el sistema cree hoy
  let corte = $state<number | null>(tarjeta.a_pagar);
  // svelte-ignore state_referenced_locally
  let ciclo = $state<number | null>(tarjeta.ciclo_actual);
  let error = $state<string | null>(null);
  let enviando = $state(false);

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    enviando = true;
    error = null;
    try {
      const r = await accion<{ mensaje: string }>("POST", `tarjetas/${tarjeta.id}/ajustar`, { corte: corte ?? 0, ciclo: ciclo ?? 0 });
      onDone(r.mensaje);
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo ajustar";
    } finally {
      enviando = false;
    }
  }
</script>

<form onsubmit={guardar}>
  <p class="hint">Abre la app de {tarjeta.nombre} y copia estos dos números. Si ya cuadran, no cambia nada.</p>
  <label class="field">
    Pago para no generar intereses (vence {fecha(tarjeta.fecha_limite)})
    <input class="money" type="number" inputmode="decimal" min="0" step="0.01" bind:value={corte} />
  </label>
  <label class="field">
    Lo que llevas del ciclo que cierra el {fecha(tarjeta.proximo_corte)}
    <input class="money" type="number" inputmode="decimal" min="0" step="0.01" bind:value={ciclo} />
  </label>
  {#if tarjeta.meses.length}<p class="hint">Las mensualidades de tus compras a meses de este ciclo ya van incluidas en ese número.</p>{/if}
  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando}>{enviando ? "Guardando…" : "Que quede igual que mi app"}</button>
</form>

<style>
  .hint {
    margin: 0 0 14px;
    font-size: 13px;
    color: var(--text-secondary);
  }
</style>
