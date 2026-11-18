<script lang="ts">
  import { accion, resumen } from "../../stores/finanzas";

  // Dinero que no es tu nomina semanal: no pasa por el 32/32/36. Lo recomendado es ahorrarlo.
  let {
    concepto = "otro",
    montoInicial = null,
    onDone,
  }: { concepto?: "fondo_plexus" | "aguinaldo" | "otro"; montoInicial?: number | null; onDone: (mensaje: string) => void } = $props();

  // svelte-ignore state_referenced_locally -- solo valor sugerido
  let monto = $state<number | null>(montoInicial);
  let destino = $state<"caja" | "fondo" | "efectivo">("caja");
  let caja = $state<number | undefined>($resumen?.cajas.find((c) => c.es_destino)?.id);
  let nota = $state("");
  let error = $state<string | null>(null);
  let enviando = $state(false);

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    enviando = true;
    error = null;
    try {
      const r = await accion<{ mensaje: string }>("POST", "ingresos-extra", { monto, destino, caja: destino === "caja" ? caja : null, concepto, nota });
      onDone(r.mensaje);
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo guardar";
    } finally {
      enviando = false;
    }
  }
</script>

<form onsubmit={guardar}>
  <label class="field">¿Cuánto te llegó?<input class="money" type="number" inputmode="decimal" min="0.01" step="0.01" required bind:value={monto} /></label>
  <div class="field">
    ¿A dónde va?
    <div class="chips" role="group" aria-label="Destino">
      <button type="button" aria-pressed={destino === "caja"} onclick={() => (destino = "caja")}>Ahorro (recomendado)</button>
      <button type="button" aria-pressed={destino === "fondo"} onclick={() => (destino = "fondo")}>Fondo de tarjetas</button>
      <button type="button" aria-pressed={destino === "efectivo"} onclick={() => (destino = "efectivo")}>Efectivo</button>
    </div>
  </div>
  {#if destino === "caja" && ($resumen?.cajas.length ?? 0) > 1}
    <label class="field">Caja
      <select bind:value={caja}>{#each $resumen?.cajas ?? [] as c (c.id)}<option value={c.id}>{c.nombre}</option>{/each}</select>
    </label>
  {/if}
  {#if concepto === "otro"}<label class="field">¿De qué es?<input maxlength="80" bind:value={nota} placeholder="regalo, venta, bono…" /></label>{/if}
  <p class="hint">Es dinero extra: no cambia tu límite de la semana. Si lo mandas a efectivo, tu sistema te deja gastarlo — mejor guárdalo o úsalo para dejar en ceros una tarjeta.</p>
  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando || !monto}>{enviando ? "Guardando…" : "Registrar"}</button>
</form>

<style>
  .hint {
    margin: -4px 0 14px;
    font-size: 12px;
    color: var(--text-tertiary);
  }
</style>
