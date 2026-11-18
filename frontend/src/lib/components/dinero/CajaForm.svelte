<script lang="ts">
  import type { FinCaja } from "../../api/types";
  import { accion } from "../../stores/finanzas";

  let { caja, onDone }: { caja?: FinCaja; onDone: (mensaje: string) => void } = $props();

  // svelte-ignore state_referenced_locally -- copia editable
  let c = $state({
    nombre: caja?.nombre ?? "", meta_nombre: caja?.meta_nombre ?? "", meta_monto: caja?.meta_monto ?? 0,
    rendimiento_anual: caja?.rendimiento_anual ?? 0, es_destino: caja?.es_destino ?? false,
  });
  let error = $state<string | null>(null);
  let enviando = $state(false);

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    enviando = true;
    error = null;
    try {
      const body = { ...c, meta_monto: c.meta_monto || 0, rendimiento_anual: c.rendimiento_anual || 0 };
      await accion(caja ? "PUT" : "POST", caja ? `cajas/${caja.id}` : "cajas", body);
      onDone(caja ? `Guardé ${c.nombre}.` : `Creé la caja ${c.nombre}.`);
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo guardar";
    } finally {
      enviando = false;
    }
  }

  async function borrar() {
    if (!caja) return;
    error = null;
    try {
      await accion("DELETE", `cajas/${caja.id}`);
      onDone(`Quité la caja ${caja.nombre}.`);
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo quitar";
    }
  }
</script>

<form onsubmit={guardar}>
  <label class="field">Nombre de la caja<input required maxlength="40" bind:value={c.nombre} placeholder="Carro, Emergencias, Viaje…" /></label>
  <div class="field-row">
    <label class="field">Meta (opcional)<input maxlength="40" bind:value={c.meta_nombre} placeholder="Enganche" /></label>
    <label class="field">Monto de la meta<input type="number" inputmode="decimal" min="0" step="1" bind:value={c.meta_monto} /></label>
  </div>
  <label class="field">Rendimiento anual % (si la caja paga intereses)<input type="number" inputmode="decimal" min="0" max="100" step="0.01" bind:value={c.rendimiento_anual} /></label>
  <div class="field">
    <div class="chips"><button type="button" aria-pressed={c.es_destino} onclick={() => (c.es_destino = !c.es_destino)}>{c.es_destino ? "✓ " : ""}Aquí cae mi ahorro semanal</button></div>
  </div>
  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando}>{enviando ? "Guardando…" : caja ? "Guardar" : "Crear caja"}</button>
  {#if caja}<button type="button" class="btn block quitar" onclick={borrar}>Quitar caja</button>{/if}
</form>

<style>
  .quitar {
    margin-top: 10px;
    color: var(--danger);
  }
</style>
