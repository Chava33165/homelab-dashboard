<script lang="ts">
  import { api } from "../../api/client";
  import type { FinReparto } from "../../api/types";
  import { accion, pesos, resumen } from "../../stores/finanzas";

  // "Recibí mi pago": muestra en vivo como se reparte (lo calcula task-hub) antes de guardar.
  let { onDone }: { onDone: (mensaje: string) => void } = $props();

  const pendiente = $resumen?.nomina_pendiente ?? null;
  let monto = $state<number | null>(pendiente?.neto ?? 2505);
  let sobrante = $state<number | null>($resumen?.sobrante_sugerido ?? 0);
  let caja = $state<number | undefined>($resumen?.cajas.find((c) => c.es_destino)?.id);
  let preview = $state<FinReparto | null>(null);
  // Semana distinta (p. ej. ponerte al corriente con tarjetas): tu decides las 3 partes.
  let personalizado = $state(false);
  let g = $state<number | null>(null);
  let fo = $state<number | null>(null);
  let a = $state<number | null>(null);
  let falta = $derived(Math.round(((monto ?? 0) - (g ?? 0) - (fo ?? 0) - (a ?? 0)) * 100) / 100);

  function personalizar() {
    personalizado = !personalizado;
    if (personalizado && preview) {
      g = preview.gastos;
      fo = preview.fondo;
      a = preview.ahorro;
    }
  }

  function extra(): Record<string, number> {
    return personalizado ? { gastos: g ?? 0, fondo: fo ?? 0, ahorro: a ?? 0 } : {};
  }
  let error = $state<string | null>(null);
  let enviando = $state(false);

  $effect(() => {
    const m = monto, s = sobrante ?? 0;
    const q = new URLSearchParams({ monto: String(m ?? 0), sobrante: String(s), ...Object.fromEntries(Object.entries(extra()).map(([k, v]) => [k, String(v)])) });
    if (!m || m <= 0 || (personalizado && falta !== 0)) {
      if (!personalizado) preview = null;
      return;
    }
    const t = setTimeout(async () => {
      try {
        preview = await api<FinReparto>(`/finanzas/reparto?${q}`);
      } catch {
        preview = null;
      }
    }, 200);
    return () => clearTimeout(t);
  });

  async function guardar(e: SubmitEvent) {
    e.preventDefault();
    enviando = true;
    error = null;
    try {
      const r = await accion<{ mensaje: string }>("POST", "pagos-recibidos", {
        monto, sobrante: sobrante ?? 0, caja, nomina_uuid: pendiente?.uuid ?? null, ...extra(),
      });
      onDone(r.mensaje);
    } catch (err) {
      error = err instanceof Error ? err.message : "No se pudo guardar";
    } finally {
      enviando = false;
    }
  }
</script>

<form onsubmit={guardar}>
  {#if pendiente}
    <p class="hint">Llegó tu nómina de la semana {pendiente.semana} ({pesos(pendiente.neto)}). La despensa ({pesos(pendiente.despensa)}) va aparte, a casa.</p>
  {/if}
  <div class="field-row">
    <label class="field">
      Te llegó
      <input type="number" inputmode="decimal" min="1" step="0.01" required bind:value={monto} />
    </label>
    <label class="field">
      Efectivo que te sobró
      <input type="number" inputmode="decimal" min="0" step="0.01" bind:value={sobrante} />
    </label>
  </div>
  {#if ($resumen?.cajas.length ?? 0) > 1}
    <label class="field">
      El ahorro va a
      <select bind:value={caja}>
        {#each $resumen?.cajas ?? [] as c (c.id)}<option value={c.id}>{c.nombre}</option>{/each}
      </select>
    </label>
  {/if}

  <div class="field">
    <div class="chips"><button type="button" aria-pressed={personalizado} onclick={personalizar}>{personalizado ? "✓ " : ""}Esta semana reparto distinto</button></div>
  </div>
  {#if personalizado}
    <div class="custom">
      <label class="field">Gastos (efectivo)<input type="number" inputmode="decimal" min="0" step="0.01" bind:value={g} /></label>
      <label class="field">Fondo tarjetas<input type="number" inputmode="decimal" min="0" step="0.01" bind:value={fo} /></label>
      <label class="field">Ahorro<input type="number" inputmode="decimal" min="0" step="0.01" bind:value={a} /></label>
    </div>
    <p class="cuadra" class:mal={falta !== 0}>{falta === 0 ? "✓ Cuadra con lo que te llegó" : falta > 0 ? `Te faltan ${pesos(falta)} por repartir` : `Te pasaste ${pesos(-falta)}`}</p>
    <p class="hint">Para pagar tarjetas, mándalo al fondo y luego usa "Pagar con fondo" en cada tarjeta.</p>
  {/if}

  {#if preview && !(personalizado && falta !== 0)}
    <ol class="split" aria-live="polite">
      <li><span class="dot g"></span><span>Retira para tus gastos</span><b>{pesos(preview.retirar)}</b></li>
      {#if preview.sobrante_a_ahorro > 0}
        <li class="sub"><span></span><span>(te sobraron {pesos(preview.sobrante_a_ahorro)}: tu límite sigue en {pesos(preview.gastos)})</span><b></b></li>
      {/if}
      <li><span class="dot f"></span><span>Aparta al fondo de tarjetas</span><b>{pesos(preview.fondo)}</b></li>
      <li><span class="dot a"></span><span>Guarda en ahorro</span><b>{pesos(preview.ahorro + preview.sobrante_a_ahorro)}</b></li>
    </ol>
  {/if}

  {#if error}<p class="form-error">{error}</p>{/if}
  <button class="btn primary block" disabled={enviando || !monto || (personalizado && falta !== 0)}>{enviando ? "Guardando…" : "Ya lo separé"}</button>
</form>

<style>
  .hint {
    margin: 0 0 14px;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .custom {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 8px;
  }

  .cuadra {
    margin: -4px 0 8px;
    font-size: 13px;
    font-weight: 800;
    color: var(--text-secondary);
  }

  .cuadra.mal {
    color: var(--c-action);
  }

  .split {
    list-style: none;
    margin: 4px 0 18px;
    padding: 14px;
    border-radius: 16px;
    background: color-mix(in srgb, var(--text) 5%, transparent);
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .split li {
    display: grid;
    grid-template-columns: 12px 1fr auto;
    align-items: center;
    gap: 10px;
    font-size: 14px;
  }

  .split li.sub {
    margin-top: -6px;
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .split b {
    font-family: var(--font-display);
    font-size: 18px;
    font-variant-numeric: tabular-nums;
  }

  .dot {
    width: 10px;
    height: 10px;
    border-radius: 3px;
  }
  .dot.g {
    background: var(--c-action);
  }
  .dot.f {
    background: var(--text);
  }
  .dot.a {
    background: color-mix(in srgb, var(--c-action) 45%, var(--bg-elevated));
  }
</style>
