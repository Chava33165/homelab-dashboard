<script lang="ts">
  import { accion, pesos, resumen } from "../../stores/finanzas";

  let { onDone }: { onDone: (mensaje: string) => void } = $props();

  let monto = $state<number | null>(null);
  let metodo = $state("efectivo");
  let categoria = $state("Comida");
  let nota = $state("");
  let meses = $state(0);
  let error = $state<string | null>(null);
  let enviando = $state(false);

  $effect(() => {
    if (metodo === "efectivo") meses = 0;
  });

  let nombres = $derived(Object.fromEntries(($resumen?.tarjetas ?? []).map((t) => [t.id, t.nombre])));

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    if (!monto || monto <= 0) return;
    enviando = true;
    error = null;
    try {
      const r = await accion<{ mensaje: string }>("POST", "gastos", { monto, metodo, categoria, nota, meses });
      onDone(r.mensaje);
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo guardar";
    } finally {
      enviando = false;
    }
  }
</script>

<form onsubmit={guardar}>
  <label class="field">
    ¿Cuánto?
    <!-- svelte-ignore a11y_autofocus -->
    <input class="money" type="number" inputmode="decimal" min="0.01" step="0.01" max="100000" required autofocus bind:value={monto} placeholder="$0" />
  </label>

  <div class="field">
    Con qué pagaste
    <div class="chips" role="group" aria-label="Método">
      {#each $resumen?.metodos ?? ["efectivo"] as m (m)}
        <button type="button" aria-pressed={metodo === m} onclick={() => (metodo = m)}>{m === "efectivo" ? "Efectivo" : nombres[m] ?? m}</button>
      {/each}
    </div>
  </div>

  {#if metodo !== "efectivo"}
    <div class="field">
      ¿De contado o a meses?
      <div class="chips" role="group" aria-label="Meses sin intereses">
        {#each [0, 3, 6, 9, 12] as m (m)}
          <button type="button" aria-pressed={meses === m} onclick={() => (meses = m)}>{m ? `${m} MSI` : "Contado"}</button>
        {/each}
      </div>
      {#if meses && monto}<span class="msi">{pesos(Math.round((monto / meses) * 100) / 100)} al mes durante {meses} meses</span>{/if}
    </div>
  {/if}

  <div class="field">
    Categoría
    <div class="chips" role="group" aria-label="Categoría">
      {#each $resumen?.categorias ?? [] as c (c)}
        <button type="button" aria-pressed={categoria === c} onclick={() => (categoria = c)}>{c}</button>
      {/each}
    </div>
  </div>

  <label class="field">
    Nota (opcional)
    <input maxlength="80" bind:value={nota} placeholder="tacos, camión…" />
  </label>

  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando || !monto}>{enviando ? "Guardando…" : "Guardar gasto"}</button>
</form>

<style>
  .msi {
    font-size: 13px;
    font-weight: 800;
    color: var(--c-action);
  }
</style>
