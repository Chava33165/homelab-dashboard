<script lang="ts">
  import type { FinCaja } from "../../api/types";
  import { accion, pesos, resumen } from "../../stores/finanzas";

  let { caja, onDone }: { caja: FinCaja; onDone: (mensaje: string) => void } = $props();

  let tipo = $state<"deposito" | "retiro" | "transferencia">("deposito");
  let monto = $state<number | null>(null);
  let destino = $state<number | undefined>(undefined);
  let nota = $state("");
  let error = $state<string | null>(null);
  let enviando = $state(false);
  let otras = $derived(($resumen?.cajas ?? []).filter((c) => c.id !== caja.id));

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    enviando = true;
    error = null;
    try {
      await accion("POST", `cajas/${caja.id}/movimientos`, { tipo, monto, destino: tipo === "transferencia" ? destino : null, nota });
      onDone(tipo === "deposito" ? `Guardaste ${pesos(monto ?? 0)} en ${caja.nombre}.` : "Listo.");
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo guardar";
    } finally {
      enviando = false;
    }
  }
</script>

<form onsubmit={guardar}>
  <p class="saldo">{caja.nombre}: <b>{pesos(caja.saldo)}</b></p>
  <div class="field">
    <div class="chips" role="group" aria-label="Tipo">
      <button type="button" aria-pressed={tipo === "deposito"} onclick={() => (tipo = "deposito")}>Depositar extra</button>
      <button type="button" aria-pressed={tipo === "retiro"} onclick={() => (tipo = "retiro")}>Sacar</button>
      {#if otras.length}<button type="button" aria-pressed={tipo === "transferencia"} onclick={() => (tipo = "transferencia")}>Mover a otra caja</button>{/if}
    </div>
  </div>
  <label class="field">Monto<input class="money" type="number" inputmode="decimal" min="0.01" step="0.01" required bind:value={monto} /></label>
  {#if tipo === "transferencia"}
    <label class="field">A la caja
      <select required bind:value={destino}>{#each otras as c (c.id)}<option value={c.id}>{c.nombre}</option>{/each}</select>
    </label>
  {/if}
  <label class="field">Nota (opcional)<input maxlength="80" bind:value={nota} /></label>
  {#if tipo === "retiro"}<p class="hint">Sacar del ahorro rompe tu racha solo si dejas de ahorrar; aun así, que sea para algo importante.</p>{/if}
  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando || !monto}>{enviando ? "Guardando…" : "Guardar"}</button>
</form>

<style>
  .saldo {
    margin: 0 0 12px;
    color: var(--text-secondary);
  }
  .saldo b {
    color: var(--text);
  }
  .hint {
    margin: -4px 0 14px;
    font-size: 12px;
    color: var(--text-tertiary);
  }
</style>
