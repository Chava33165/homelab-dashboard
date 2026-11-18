<script lang="ts">
  import type { FinEstadoMes } from "../../api/types";
  import { pesos } from "../../stores/finanzas";

  // Carrusel de estados de cuenta, mes por mes: lo cerrado, lo que va en curso y lo que ya tienes
  // comprometido a meses. Desliza con el dedo (scroll-snap nativo, sin JS de arrastre).
  let { meses, tarjeta }: { meses: FinEstadoMes[]; tarjeta: string } = $props();

  const MES = ["ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"];
  const parte = (iso: string) => iso.split("-").map(Number);
  const ETIQUETA = { cerrado: "a pagar", en_curso: "va en curso", futuro: "solo a meses" } as const;
</script>

{#if meses.length}
  <div class="wrap">
    <span class="head">Tus pagos mes a mes</span>
    <ol class="rail" aria-label="Estados de cuenta de {tarjeta}">
      {#each meses as m, i (m.fecha_limite)}
        {@const [y, mo, d] = parte(m.fecha_limite)}
        <li class="mes {m.estado}" class:vacio={m.estado === "cerrado" && m.monto === 0}>
          <span class="tag">{ETIQUETA[m.estado]}</span>
          <span class="nombre">{MES[mo - 1]}{y !== new Date().getFullYear() ? ` ${String(y).slice(2)}` : ""}</span>
          <span class="dia">antes del {d}</span>
          <b class="monto">{m.estado === "cerrado" && m.monto === 0 ? "✓ pagado" : pesos(m.monto)}</b>
          {#if m.estado === "en_curso"}
            <span class="sub">{m.monto_meses ? `incluye ${pesos(m.monto_meses)} a meses · ` : ""}sube con lo que gastes</span>
          {:else if m.detalle.length}
            <ul class="detalle">
              {#each m.detalle as x, k (k)}
                <li><span>{x.descripcion}</span><span>{pesos(x.mensualidad)}{x.ultimo ? " · último" : ` · faltan ${x.faltan}`}</span></li>
              {/each}
            </ul>
          {:else if m.estado === "futuro"}
            <span class="sub">sin meses aquí</span>
          {/if}
          {#if i === 0 && meses.length > 1}<span class="hint" aria-hidden="true">desliza →</span>{/if}
        </li>
      {/each}
    </ol>
  </div>
{/if}

<style>
  .wrap {
    margin-top: 16px;
  }

  .head {
    display: block;
    margin-bottom: 8px;
    font-size: 12px;
    font-weight: 800;
    color: var(--text-secondary);
  }

  /* Riel horizontal: cada mes se "pega" al borde al soltar (scroll-snap). */
  .rail {
    list-style: none;
    margin: 0 -20px;
    padding: 2px 20px 6px;
    display: flex;
    gap: 10px;
    overflow-x: auto;
    scroll-snap-type: x mandatory;
    scroll-padding-inline: 20px;
    overscroll-behavior-x: contain;
    scrollbar-width: none;
  }

  .rail::-webkit-scrollbar {
    display: none;
  }

  .mes {
    position: relative;
    flex: 0 0 44%;
    min-width: 132px;
    scroll-snap-align: start;
    display: flex;
    flex-direction: column;
    gap: 1px;
    padding: 12px 12px 14px;
    border-radius: 18px;
    border: 1.5px solid var(--border-strong);
    background: color-mix(in srgb, var(--text) 4%, transparent);
  }

  /* Lo proximo a pagar resalta en rosa; lo en curso en tinta; lo futuro discreto (borde punteado). */
  .mes.cerrado:not(.vacio) {
    border-color: var(--c-action);
    background: color-mix(in srgb, var(--c-action) 13%, transparent);
  }

  .mes.en_curso {
    border-color: var(--text);
  }

  .mes.futuro {
    border-style: dashed;
  }

  .tag {
    align-self: flex-start;
    margin-bottom: 6px;
    padding: 2px 7px;
    border-radius: 99em;
    font-size: 9px;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    background: color-mix(in srgb, var(--text) 10%, transparent);
    color: var(--text-secondary);
  }

  .cerrado:not(.vacio) .tag {
    background: var(--c-action);
    color: #fff;
  }

  .nombre {
    font-family: var(--font-display);
    font-size: 22px;
    font-weight: 800;
    letter-spacing: -0.02em;
    text-transform: capitalize;
    line-height: 1;
  }

  .dia {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .monto {
    margin-top: 8px;
    font-family: var(--font-display);
    font-size: 20px;
    font-weight: 800;
    font-variant-numeric: tabular-nums;
  }

  .sub {
    font-size: 11px;
    line-height: 1.35;
    color: var(--text-tertiary);
  }

  .detalle {
    list-style: none;
    margin: 4px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .detalle li {
    display: flex;
    flex-direction: column;
    font-size: 11px;
    line-height: 1.3;
  }

  .detalle li span:first-child {
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .detalle li span:last-child {
    color: var(--text-tertiary);
  }

  .hint {
    position: absolute;
    top: 12px;
    right: 12px;
    font-size: 10px;
    font-weight: 700;
    color: var(--text-tertiary);
  }

  @media (min-width: 640px) {
    .mes {
      flex-basis: 30%;
    }
  }
</style>
