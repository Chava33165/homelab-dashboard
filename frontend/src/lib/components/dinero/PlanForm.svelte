<script lang="ts">
  import type { FinTarjeta } from "../../api/types";
  import { accion, pesos } from "../../stores/finanzas";

  let { tarjeta, onDone }: { tarjeta: FinTarjeta; onDone: (mensaje: string) => void } = $props();

  const MESES = ["enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"];
  const hoy = new Date();
  const opciones = Array.from({ length: 36 }, (_, i) => {
    const d = new Date(hoy.getFullYear(), hoy.getMonth() + i, 1);
    return { v: `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`, l: `${MESES[d.getMonth()]} ${d.getFullYear()}` };
  });

  let descripcion = $state("");
  let mensualidad = $state<number | null>(null);
  let ultimo = $state(opciones[2].v);
  let incluido = $state(true);
  let error = $state<string | null>(null);
  let enviando = $state(false);

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    enviando = true;
    error = null;
    try {
      await accion("POST", "planes", { tarjeta: tarjeta.id, descripcion, mensualidad, ultimo_mes: ultimo, incluido });
      onDone(`Agregué ${descripcion} a meses en ${tarjeta.nombre}.`);
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo guardar";
    } finally {
      enviando = false;
    }
  }
</script>

<form onsubmit={guardar}>
  <label class="field">¿Qué compraste?<input required maxlength="40" bind:value={descripcion} placeholder="Celular, audífonos…" /></label>
  <div class="field-row">
    <label class="field">Mensualidad<input type="number" inputmode="decimal" min="0.01" step="0.01" required bind:value={mensualidad} placeholder="$0" /></label>
    <label class="field">Último pago en
      <select bind:value={ultimo}>{#each opciones as o (o.v)}<option value={o.v}>{o.l}</option>{/each}</select>
    </label>
  </div>
  <div class="field">
    <div class="chips"><button type="button" aria-pressed={incluido} onclick={() => (incluido = !incluido)}>{incluido ? "✓ " : ""}La de este ciclo ya va en lo que llevas</button></div>
  </div>
  <p class="hint">Cada corte se suma sola la mensualidad{mensualidad ? ` de ${pesos(mensualidad)}` : ""} a {tarjeta.nombre} hasta el último pago. Lo que falta se resta de tu crédito disponible.</p>
  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando || !mensualidad || !descripcion}>{enviando ? "Guardando…" : "Agregar compra a meses"}</button>
</form>

<style>
  .hint {
    margin: -4px 0 14px;
    font-size: 12px;
    color: var(--text-tertiary);
  }
</style>
