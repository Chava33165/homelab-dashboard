<script lang="ts">
  import { accion, fecha, resumen } from "../../stores/finanzas";

  // Con lo que tienes HOY arranca el sistema. Si ya se capturo, volver a guardar lo reemplaza.
  let { onDone }: { onDone: (mensaje: string) => void } = $props();

  // Prellenado con lo ya capturado: guardar reemplaza todo, asi corregir un dato no borra los demas.
  const ini = $resumen?.iniciales;
  const v = (n: number | undefined) => (n ? n : null);
  let efectivo = $state<number | null>(v(ini?.efectivo));
  let fondo = $state<number | null>(v(ini?.fondo));
  let ahorro = $state<number | null>(v(ini?.ahorro));
  let deudas = $state<Record<string, { corte: number | null; ciclo: number | null }>>(
    Object.fromEntries(($resumen?.tarjetas ?? []).map((t) => [t.id, { corte: v(ini?.deudas[t.id]?.corte), ciclo: v(ini?.deudas[t.id]?.ciclo) }]))
  );
  let error = $state<string | null>(null);
  let enviando = $state(false);

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    enviando = true;
    error = null;
    try {
      await accion("POST", "saldos-iniciales", {
        efectivo: efectivo || 0, fondo: fondo || 0, ahorro: ahorro || 0,
        deudas: Object.fromEntries(Object.entries(deudas).map(([k, v]) => [k, { corte: v.corte || 0, ciclo: v.ciclo || 0 }])),
      });
      onDone("¡Listo! Tu sistema ya arrancó.");
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo guardar";
    } finally {
      enviando = false;
    }
  }
</script>

<form onsubmit={guardar}>
  <p class="hint">{ini && $resumen?.iniciado ? "Estos son tus saldos de arranque. Cambia solo lo que esté mal; lo demás se queda igual." : "Pon cuánto tienes hoy. Deja en blanco lo que esté en $0."}</p>
  <div class="field-row">
    <label class="field">Efectivo en tu cartera<input type="number" inputmode="decimal" min="0" step="0.01" bind:value={efectivo} /></label>
    <label class="field">Ya apartado para tarjetas<input type="number" inputmode="decimal" min="0" step="0.01" bind:value={fondo} /></label>
  </div>
  <label class="field">Ahorro que ya tienes (sin el fondo Plexus)<input type="number" inputmode="decimal" min="0" step="0.01" bind:value={ahorro} /></label>
  {#each $resumen?.tarjetas ?? [] as t (t.id)}
    <fieldset class="card">
      <legend>{t.nombre}</legend>
      <div class="field-row">
        <label class="field">Pagar antes del {fecha(t.fecha_limite)}<input type="number" inputmode="decimal" min="0" step="0.01" bind:value={deudas[t.id].corte} placeholder="$0" /></label>
        <label class="field">Del ciclo que cierra el {fecha(t.proximo_corte)}<input type="number" inputmode="decimal" min="0" step="0.01" bind:value={deudas[t.id].ciclo} placeholder="$0" /></label>
      </div>
    </fieldset>
  {/each}
  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando}>{enviando ? "Guardando…" : "Empezar"}</button>
</form>

<style>
  .card {
    margin: 0 0 6px;
    padding: 10px 12px 0;
    border: 1.5px solid var(--border);
    border-radius: 16px;
  }

  legend {
    padding: 0 6px;
    font-weight: 800;
  }

  .hint {
    margin: 0 0 14px;
    font-size: 13px;
    color: var(--text-secondary);
  }
</style>
