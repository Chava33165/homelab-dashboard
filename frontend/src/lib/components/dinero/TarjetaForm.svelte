<script lang="ts">
  import type { FinTarjeta } from "../../api/types";
  import { accion } from "../../stores/finanzas";

  let { tarjeta, onDone }: { tarjeta: FinTarjeta; onDone: (mensaje: string) => void } = $props();

  // svelte-ignore state_referenced_locally -- copia editable de la tarjeta
  let t = $state({
    nombre: tarjeta.nombre, banco: tarjeta.banco, ultimos4: tarjeta.ultimos4, dia_corte: tarjeta.dia_corte,
    dia_pago: tarjeta.dia_pago, limite_banco: tarjeta.limite_banco, tope_personal: tarjeta.tope_personal,
    estado: tarjeta.estado, nota: tarjeta.nota,
  });
  let error = $state<string | null>(null);
  let enviando = $state(false);

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    enviando = true;
    error = null;
    try {
      await accion("PUT", `tarjetas/${tarjeta.id}`, { ...t, limite_banco: t.limite_banco || 0, tope_personal: t.tope_personal || 0 });
      onDone(`Guardé ${t.nombre}.`);
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo guardar";
    } finally {
      enviando = false;
    }
  }
</script>

<form onsubmit={guardar}>
  <div class="field-row">
    <label class="field">Nombre<input required maxlength="30" bind:value={t.nombre} /></label>
    <label class="field">Banco<input maxlength="30" bind:value={t.banco} /></label>
  </div>
  <div class="field-row">
    <label class="field">Día de corte<input type="number" min="1" max="31" required bind:value={t.dia_corte} /></label>
    <label class="field">Día límite de pago<input type="number" min="1" max="31" required bind:value={t.dia_pago} /></label>
  </div>
  <div class="field-row">
    <label class="field">Límite del banco<input type="number" inputmode="decimal" min="0" step="1" bind:value={t.limite_banco} /></label>
    <label class="field">Tu tope por ciclo<input type="number" inputmode="decimal" min="0" step="1" bind:value={t.tope_personal} /></label>
  </div>
  <div class="field-row">
    <label class="field">Últimos 4 dígitos<input inputmode="numeric" pattern="\d{'{4}'}" maxlength="4" bind:value={t.ultimos4} placeholder="1234" /></label>
    <div class="field">
      Estado
      <div class="chips" role="group" aria-label="Estado">
        {#each [["activa", "Activa"], ["pausada", "Pausada"], ["bloqueada", "Bloqueada"]] as [v, l] (v)}
          <button type="button" aria-pressed={t.estado === v} onclick={() => (t.estado = v as FinTarjeta["estado"])}>{l}</button>
        {/each}
      </div>
    </div>
  </div>
  <label class="field">Nota<input maxlength="120" bind:value={t.nota} placeholder="ej. solo para súper y gasolina" /></label>
  <p class="hint">El tope es lo máximo que tú te permites gastar por ciclo (aunque el banco te preste más). Te aviso al 80 %.</p>
  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando}>{enviando ? "Guardando…" : "Guardar"}</button>
</form>

<style>
  .hint {
    margin: -4px 0 14px;
    font-size: 12px;
    color: var(--text-tertiary);
  }
</style>
