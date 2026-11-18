<script lang="ts">
  import { onMount } from "svelte";
  import type { FinCaja, FinTarjeta } from "../api/types";
  import { accion, cargarNominas, fecha, finError, movimientos, nominas, pesos, refrescar, resumen } from "../stores/finanzas";
  import GlassCard from "../components/GlassCard.svelte";
  import Bars from "../components/Bars.svelte";
  import ConfirmDialog from "../components/ConfirmDialog.svelte";
  import Sheet from "../components/dinero/Sheet.svelte";
  import GastoForm from "../components/dinero/GastoForm.svelte";
  import PagoForm from "../components/dinero/PagoForm.svelte";
  import PagarTarjetaForm from "../components/dinero/PagarTarjetaForm.svelte";
  import TarjetaForm from "../components/dinero/TarjetaForm.svelte";
  import CajaForm from "../components/dinero/CajaForm.svelte";
  import CajaMovForm from "../components/dinero/CajaMovForm.svelte";
  import InicioForm from "../components/dinero/InicioForm.svelte";
  import AjusteForm from "../components/dinero/AjusteForm.svelte";
  import PlanForm from "../components/dinero/PlanForm.svelte";
  import ExtraForm from "../components/dinero/ExtraForm.svelte";
  import MesesCarrusel from "../components/dinero/MesesCarrusel.svelte";
  import GastosPeriodo from "../components/dinero/GastosPeriodo.svelte";
  import CashyCara, { type Gesto } from "../components/dinero/CashyCara.svelte";
  import CashyChat from "../components/dinero/CashyChat.svelte";

  type Seccion = "semana" | "tarjetas" | "ahorro" | "nomina";
  const SECCIONES: [Seccion, string][] = [["semana", "Semana"], ["tarjetas", "Tarjetas"], ["ahorro", "Ahorro"], ["nomina", "Nómina"]];
  let seccion = $state<Seccion>(leerSeccion());

  function leerSeccion(): Seccion {
    try {
      const s = localStorage.getItem("dinero-seccion") as Seccion | null;
      return s && SECCIONES.some(([k]) => k === s) ? s : "semana";
    } catch {
      return "semana";
    }
  }

  $effect(() => {
    try {
      localStorage.setItem("dinero-seccion", seccion);
    } catch {
      /* sin almacenamiento: no pasa nada */
    }
    if (seccion === "nomina" || seccion === "ahorro") cargarNominas().catch(() => void 0);
  });

  onMount(() => {
    refrescar();
    // Al volver a la app (p. ej. despues de registrar un gasto con el Atajo) se ve lo nuevo.
    const onVisible = () => document.visibilityState === "visible" && refrescar();
    document.addEventListener("visibilitychange", onVisible);
    return () => document.removeEventListener("visibilitychange", onVisible);
  });

  // ---- hojas / formularios
  type Hoja =
    | { k: "gasto" } | { k: "pago" } | { k: "inicio" } | { k: "caja-nueva" }
    | { k: "pagar"; t: FinTarjeta } | { k: "tarjeta"; t: FinTarjeta }
    | { k: "caja"; c: FinCaja } | { k: "caja-mov"; c: FinCaja }
    | { k: "ajuste"; t: FinTarjeta } | { k: "plan"; t: FinTarjeta }
    | { k: "extra"; concepto: "fondo_plexus" | "aguinaldo" | "otro"; monto: number | null };
  let hoja = $state<Hoja | null>(null);
  let hojaAbierta = $state(false);
  let aviso = $state<string | null>(null);
  let avisoTimer = 0;

  function abrir(h: Hoja) {
    hoja = h;
    hojaAbierta = true;
  }

  function listo(mensaje: string) {
    hojaAbierta = false;
    aviso = mensaje;
    clearTimeout(avisoTimer);
    avisoTimer = window.setTimeout(() => (aviso = null), 5000);
  }

  const TITULOS: Record<Hoja["k"], string> = {
    gasto: "Nuevo gasto", pago: "Recibí mi pago", inicio: "Empezar mi sistema", "caja-nueva": "Nueva caja",
    pagar: "Pagar tarjeta", tarjeta: "Datos de la tarjeta", caja: "Editar caja", "caja-mov": "Mover dinero",
    ajuste: "Mi app dice…", plan: "Compra a meses", extra: "Ingreso extra",
  };

  // ---- borrar (movimientos y compras a meses)
  let borrarPath = $state<string | null>(null);
  let confirmarBorrar = $state(false);
  function pedirBorrar(path: string) {
    borrarPath = path;
    confirmarBorrar = true;
  }
  async function borrar() {
    if (borrarPath === null) return;
    try {
      await accion("DELETE", borrarPath);
      listo("Borrado.");
    } catch (e) {
      listo(e instanceof Error ? e.message : "No se pudo borrar");
    }
  }

  // ---- derivados
  let r = $derived($resumen);
  // Cashy: preocupada si hay alertas fuertes, feliz si el fondo alcanza y vas dentro del limite.
  let cashyAbierta = $state(false);
  let gestoCashy = $derived<Gesto>(
    !r ? "neutral"
    : r.alertas.some((a) => a.nivel === "alto") ? "preocupada"
    : r.fondo >= r.a_pagar_total && r.semana.gastado <= r.semana.limite ? "feliz"
    : "neutral"
  );
  let porCobrar = $derived((r?.fondo_plexus && !r.fondo_plexus.cobrado ? r.fondo_plexus.fondo_saldo_total : 0) + (r?.aguinaldo && !r.aguinaldo.cobrado ? r.aguinaldo.estimado : 0));
  let usoSemana = $derived(r && r.semana.limite ? Math.min(r.semana.gastado / r.semana.limite, 1.2) : 0);
  let nombreTarjeta = $derived(Object.fromEntries((r?.tarjetas ?? []).map((t) => [t.id, t.nombre])));
  let nombreCaja = $derived(Object.fromEntries((r?.cajas ?? []).map((c) => [c.id, c.nombre])));
  let categoriasSemana = $derived(Object.entries(r?.categorias_semana ?? {}).sort((a, b) => b[1] - a[1]));
  let categoriasMes = $derived(Object.entries(r?.categorias_mes ?? {}).sort((a, b) => b[1] - a[1]));
  let maxCatMes = $derived(Math.max(1, ...categoriasMes.map(([, v]) => v)));
  let semanasBars = $derived(
    (r?.semanas ?? []).map((s) => ({ label: fecha(s.inicio).split(" ")[0], detail: `semana del ${fecha(s.inicio)}`, values: [s.gastado] }))
  );
  let filtro = $state("");
  let movsFiltrados = $derived($movimientos.filter((m) => !filtro || m.metodo === filtro || m.categoria === filtro || m.tipo === filtro));

  function describir(m: (typeof $movimientos)[number]): { titulo: string; sub: string; signo: string } {
    switch (m.tipo) {
      case "gasto":
        return { titulo: m.nota || m.categoria, sub: `${m.categoria} · ${m.metodo === "efectivo" ? "Efectivo" : nombreTarjeta[m.metodo] ?? m.metodo}`, signo: "−" };
      case "pago_recibido":
        return { titulo: "Pago recibido", sub: `retiraste ${pesos(m.retirar ?? 0)}`, signo: "+" };
      case "pago_tarjeta":
        return { titulo: `Pago a ${nombreTarjeta[m.tarjeta] ?? m.tarjeta}`, sub: "desde el fondo", signo: "" };
      case "deposito_caja":
        return { titulo: `Depósito a ${nombreCaja[m.caja ?? 0] ?? "caja"}`, sub: m.nota, signo: "+" };
      case "retiro_caja":
        return { titulo: `Retiro de ${nombreCaja[m.caja ?? 0] ?? "caja"}`, sub: m.nota, signo: "−" };
      case "transferencia_caja":
        return { titulo: "Movimiento entre cajas", sub: `${nombreCaja[m.caja ?? 0] ?? "?"} → ${nombreCaja[m.caja_destino ?? 0] ?? "?"}`, signo: "" };
      case "compra_meses":
        return { titulo: m.nota || m.categoria, sub: `${m.categoria} · ${nombreTarjeta[m.metodo] ?? m.metodo} a meses`, signo: "−" };
      case "ajuste_tarjeta":
        return { titulo: `Ajuste ${nombreTarjeta[m.metodo] ?? m.metodo}`, sub: "para cuadrar con tu app", signo: m.monto < 0 ? "−" : "" };
      case "ingreso_extra":
        return { titulo: m.nota || "Ingreso extra", sub: m.metodo === "caja" ? `a ${nombreCaja[m.caja ?? 0] ?? "ahorro"}` : m.metodo === "fondo" ? "al fondo de tarjetas" : "a efectivo", signo: "+" };
      case "inicial":
        return { titulo: "Saldo inicial", sub: m.metodo === "caja" ? "ahorro" : nombreTarjeta[m.metodo] ? `deuda ${nombreTarjeta[m.metodo]}` : m.metodo, signo: "" };
      default:
        return { titulo: m.tipo, sub: "", signo: "" };
    }
  }

  function cuando(ts: number): string {
    const d = new Date(ts * 1000);
    const hoy = new Date();
    const hora = d.toLocaleTimeString("es-MX", { hour: "numeric", minute: "2-digit" });
    if (d.toDateString() === hoy.toDateString()) return `hoy ${hora}`;
    return `${d.toLocaleDateString("es-MX", { weekday: "short", day: "numeric", month: "short" })}`;
  }

  // ---- nomina
  let ultima = $derived($nominas[0]);
  let nominaOrden = $derived([...$nominas].reverse());
  let nominaBars = $derived(nominaOrden.map((n) => ({ label: String(n.semana), detail: `Semana ${n.semana} · pagada ${fecha(n.fecha_pago)}`, values: [n.neto] })));
  let fondoBars = $derived(nominaOrden.map((n) => ({ label: String(n.semana), detail: `Semana ${n.semana}`, values: [n.fondo_tuyo, n.fondo_empresa] })));
  let promedioNeto = $derived($nominas.length ? $nominas.reduce((a, n) => a + n.neto, 0) / $nominas.length : 0);

  function tramos(n: NonNullable<typeof ultima>) {
    const otros = Math.max(n.deducciones - n.isr - n.imss - n.fondo_ahorro - n.cafeteria, 0);
    return [
      { k: "neto", label: "A tu cuenta", v: n.neto },
      { k: "fondo", label: "Fondo de ahorro", v: n.fondo_ahorro },
      { k: "isr", label: "Impuesto (ISR)", v: n.isr },
      { k: "imss", label: "IMSS", v: n.imss },
      { k: "cafe", label: "Cafetería", v: n.cafeteria },
      ...(otros > 0.5 ? [{ k: "otros", label: "Otros", v: otros }] : []),
    ];
  }
</script>

<div class="page wide">
  <div class="page-head">
    <span class="eyebrow">Tu sistema</span>
    <h1 class="display-title chromatic-pop">Dinero</h1>
    {#if r?.iniciado}<span class="ink-chip">{r.semana.dias_para_pago === 7 ? "Hoy es día de pago" : `Pago en ${r.semana.dias_para_pago} ${r.semana.dias_para_pago === 1 ? "día" : "días"}`}</span>{/if}
  </div>

  {#if $finError && !r}
    <div class="empty glass"><p>{$finError}</p><button class="btn" onclick={refrescar}>Reintentar</button></div>
  {:else if !r}
    <div class="empty glass"><p>Cargando…</p></div>
  {:else if !r.iniciado}
    <GlassCard padding="24px">
      <h2 class="big-h">Arranquemos tu sistema</h2>
      <p class="lead">Cada jueves tu pago se divide en <b>32 % gastos</b>, <b>32 % fondo de tarjetas</b> y <b>36 % ahorro</b>. Primero dime con cuánto empiezas hoy.</p>
      <button class="btn primary block" onclick={() => abrir({ k: "inicio" })}>Capturar mis saldos de hoy</button>
    </GlassCard>
  {:else}
    {#if aviso}<div class="toast" role="status">{aviso}</div>{/if}

    {#if r.alertas.length}
      <ul class="alertas">
        {#each r.alertas as a, i (i)}
          <li class:alto={a.nivel === "alto"}>
            <svg aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><path d="M12 9v4M12 17h.01M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" /></svg>
            {a.texto}
          </li>
        {/each}
      </ul>
    {/if}

    <div class="seg-wrap">
      <div class="seg" role="tablist" aria-label="Sección">
        {#each SECCIONES as [k, l] (k)}
          <button role="tab" aria-selected={seccion === k} onclick={() => (seccion = k)}>{l}</button>
        {/each}
      </div>
    </div>

    <!-- ============================ SEMANA ============================ -->
    {#if seccion === "semana"}
      <div class="stack board stagger">
        <GlassCard padding="22px">
          <span class="label">Te quedan esta semana</span>
          <div class="hero" class:over={usoSemana >= 0.8}>
            <span class="hero-num">{pesos(r.efectivo)}</span>
            {#if r.semana.limite}<span class="hero-of">de {pesos(r.semana.limite)}</span>{/if}
          </div>
          <div class="meter" role="meter" aria-valuemin="0" aria-valuemax={r.semana.limite} aria-valuenow={r.semana.gastado} aria-label="Gastado en efectivo esta semana">
            <i style:width="{Math.min(usoSemana, 1) * 100}%"></i>
          </div>
          <p class="foot">Gastaste {pesos(r.semana.gastado)} en efectivo · tu pago llega el jueves {fecha(r.semana.proximo_pago)}</p>
          <div class="actions">
            <button class="btn primary" onclick={() => abrir({ k: "gasto" })}>+ Gasto</button>
            <button class="btn" class:pulse={!!r.nomina_pendiente} onclick={() => abrir({ k: "pago" })}>{r.nomina_pendiente ? `Llegó tu pago · ${pesos(r.nomina_pendiente.neto)}` : "Recibí mi pago"}</button>
          </div>
        </GlassCard>

        <div class="full">
          <GlassCard padding="20px"><GastosPeriodo /></GlassCard>
        </div>

        <div class="tri">
          <div class="box glass"><span class="label">Efectivo</span><b>{pesos(r.efectivo)}</b><span class="sub">para gastar</span></div>
          <div class="box glass"><span class="label">Fondo tarjetas</span><b>{pesos(r.fondo)}</b><span class="sub" class:warn={r.fondo < r.a_pagar_total}>debes pagar {pesos(r.a_pagar_total)}</span></div>
          <div class="box glass"><span class="label">Ahorro</span><b>{pesos(r.ahorro_total)}</b><span class="sub">{r.racha_ahorro} {r.racha_ahorro === 1 ? "semana" : "semanas"} seguidas</span></div>
        </div>

        <GlassCard padding="20px">
          <div class="card-head"><h2>Tu patrimonio</h2><b class="total" class:neg={r.patrimonio < 0}>{pesos(r.patrimonio)}</b></div>
          <p class="note">Efectivo + fondo + ahorro − lo que debes en tarjetas ({pesos(r.deuda_total)}).{porCobrar ? ` Aparte tienes ~${pesos(porCobrar, 0)} por cobrar (fondo Plexus y aguinaldo).` : ""}</p>
        </GlassCard>

        <GlassCard padding="20px">
          <div class="card-head"><h2>Gasto por semana</h2><span class="total">límite {pesos(r.semana.limite)}</span></div>
          <Bars label="Gasto total por semana" bars={semanasBars} series={[{ name: "Gastado", color: "var(--c-action)" }]} format={(n) => pesos(n)} height={100} />
          <p class="note">Incluye efectivo y tarjetas. La semana va de jueves a miércoles.</p>
        </GlassCard>

        {#if categoriasMes.length}
          <GlassCard padding="20px">
            <div class="card-head"><h2>¿En qué se va?</h2><span class="total">este mes</span></div>
            <ul class="hbars">
              {#each categoriasMes as [cat, v] (cat)}
                {@const sem = categoriasSemana.find(([c]) => c === cat)?.[1] ?? 0}
                <li>
                  <span class="hb-name">{cat}</span>
                  <span class="hb-bar"><i style:width="{(v / maxCatMes) * 100}%"></i></span>
                  <b>{pesos(v)}</b>
                  {#if sem}<span class="hb-sub">{pesos(sem)} esta semana</span>{/if}
                </li>
              {/each}
            </ul>
          </GlassCard>
        {/if}

        <GlassCard padding="20px">
          <div class="card-head"><h2>Movimientos</h2><button class="link" onclick={() => abrir({ k: "inicio" })}>Corregir saldos iniciales</button></div>
          <div class="chips filtro" role="group" aria-label="Filtrar">
            <button aria-pressed={filtro === ""} onclick={() => (filtro = "")}>Todo</button>
            <button aria-pressed={filtro === "efectivo"} onclick={() => (filtro = "efectivo")}>Efectivo</button>
            {#each r.tarjetas as t (t.id)}<button aria-pressed={filtro === t.id} onclick={() => (filtro = t.id)}>{t.nombre}</button>{/each}
            {#each r.categorias as c (c)}<button aria-pressed={filtro === c} onclick={() => (filtro = c)}>{c}</button>{/each}
          </div>
          {#if movsFiltrados.length}
            <ul class="movs">
              {#each movsFiltrados as m (m.id)}
                {@const d = describir(m)}
                <li>
                  <span class="mv-main"><b>{d.titulo}</b><span>{d.sub} · {cuando(m.ts)}</span></span>
                  <span class="mv-amt" class:neg={d.signo === "−"} class:pos={d.signo === "+"}>{d.signo}{pesos(Math.abs(m.monto))}</span>
                  <button class="mv-del" aria-label="Borrar {d.titulo}" onclick={() => pedirBorrar(`movimientos/${m.id}`)}>
                    <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg>
                  </button>
                </li>
              {/each}
            </ul>
          {:else}
            <p class="note">Nada por aquí todavía. Registra gastos con el botón o con tu Atajo "Gasté" del iPhone.</p>
          {/if}
        </GlassCard>
      </div>

    <!-- ============================ TARJETAS ============================ -->
    {:else if seccion === "tarjetas"}
      <div class="stack board stagger">
        <div class="fondo-line glass">
          <span>Fondo para tarjetas</span><b>{pesos(r.fondo)}</b>
          <span class="sub" class:warn={r.fondo < r.a_pagar_total}>{r.fondo >= r.a_pagar_total ? `alcanza para lo del corte (${pesos(r.a_pagar_total)})` : `te faltan ${pesos(r.a_pagar_total - r.fondo)} para lo del corte`}</span>
        </div>
        {#each r.tarjetas as t (t.id)}
          <article class="cc" class:off={t.estado !== "activa"} data-card={t.id}>
            <div class="cc-face">
              <div class="cc-top">
                <span class="cc-name">{t.nombre}</span>
                <span class="cc-tag">{t.estado === "activa" ? t.banco || "Crédito" : t.estado.toUpperCase()}</span>
              </div>
              <span class="cc-label">A pagar del corte del {fecha(t.ultimo_corte)}</span>
              <span class="cc-amount">{pesos(t.a_pagar)}</span>
              <div class="cc-bottom">
                <span>{t.a_pagar > 0 ? (t.dias_para_pagar >= 0 ? `antes del ${fecha(t.fecha_limite)} · ${t.dias_para_pagar === 0 ? "¡hoy!" : `faltan ${t.dias_para_pagar} d`}` : `¡venció el ${fecha(t.fecha_limite)}!`) : "nada pendiente del corte ✓"}</span>
                <span class="cc-num">•••• {t.ultimos4 || "····"}</span>
              </div>
            </div>
            <div class="cc-info">
              <div class="cc-row">
                <span>Este ciclo (cierra {fecha(t.proximo_corte)})</span>
                <b>{pesos(t.ciclo_actual)}{t.tope_personal ? ` de ${pesos(t.tope_personal)}` : ""}</b>
              </div>
              {#if t.tope_personal}
                <div class="meter small" class:over={(t.uso_tope ?? 0) >= 0.8}><i style:width="{Math.min(t.uso_tope ?? 0, 1) * 100}%"></i></div>
              {:else}
                <p class="note tight">Ponle un tope personal para que te avise antes de pasarte.</p>
              {/if}
              <dl class="cc-facts">
                <div><dt>Deuda total</dt><dd>{pesos(t.deuda)}</dd></div>
                <div><dt>Límite del banco</dt><dd>{t.limite_banco ? pesos(t.limite_banco) : "—"}</dd></div>
                <div><dt>Disponible</dt><dd>{t.disponible_banco !== null ? pesos(t.disponible_banco) : "—"}</dd></div>
                <div><dt>Corte · pago</dt><dd>día {t.dia_corte} · {t.dia_pago}</dd></div>
              </dl>
              {#if t.nota}<p class="note tight">{t.nota}</p>{/if}

              <MesesCarrusel meses={t.calendario} tarjeta={t.nombre} />

              <div class="meses">
                <div class="meses-head">
                  <span>Compras a meses</span>
                  {#if t.meses_restante}<b>debes {pesos(t.meses_restante)} a meses</b>{/if}
                </div>
                {#each t.meses as p (p.id)}
                  <div class="plan">
                    <span class="plan-main"><b>{p.descripcion}</b><span>{pesos(p.mensualidad)}/mes · {p.pagos.length ? `${p.pagos.length === 1 ? "último pago" : `faltan ${p.pagos.length}, el último`} el ${fecha(p.pagos[p.pagos.length - 1])} · resta ${pesos(p.restante)}` : "ya terminó ✓"}</span></span>
                    <button class="mv-del" aria-label="Quitar {p.descripcion}" onclick={() => pedirBorrar(`planes/${p.id}`)}>
                      <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg>
                    </button>
                  </div>
                {/each}
                <button class="link" onclick={() => abrir({ k: "plan", t })}>+ Agregar compra a meses</button>
              </div>

              <div class="actions">
                <button class="btn primary" disabled={t.deuda <= 0} onclick={() => abrir({ k: "pagar", t })}>Pagar con fondo</button>
                <button class="btn" onclick={() => abrir({ k: "ajuste", t })}>Mi app dice…</button>
              </div>
              <button class="link block-link" onclick={() => abrir({ k: "tarjeta", t })}>Datos de la tarjeta (límite, tope, corte, pago)</button>
            </div>
          </article>
        {/each}
        <p class="note center">Regla: el fondo solo paga tarjetas y el ahorro nunca paga tarjetas. Paga siempre lo del corte completo antes de la fecha límite para no generar intereses.</p>
      </div>

    <!-- ============================ AHORRO ============================ -->
    {:else if seccion === "ahorro"}
      <div class="stack board stagger">
        <GlassCard padding="22px">
          <span class="label">Ahorro total</span>
          <div class="hero"><span class="hero-num">{pesos(r.ahorro_total)}</span></div>
          <p class="foot">en tus cajas · racha de {r.racha_ahorro} {r.racha_ahorro === 1 ? "semana" : "semanas"}{porCobrar ? ` · y ${pesos(porCobrar)} por cobrar` : ""}</p>
          <div class="actions"><button class="btn" onclick={() => abrir({ k: "extra", concepto: "otro", monto: null })}>+ Ingreso extra</button></div>
        </GlassCard>

        {#each r.cajas as c (c.id)}
          <GlassCard padding="20px">
            <div class="card-head">
              <h2>{c.nombre}{#if c.es_destino}<span class="mini-tag">aquí cae tu ahorro</span>{/if}</h2>
              <b class="total">{pesos(c.saldo)}</b>
            </div>
            {#if c.meta_monto}
              <div class="meter"><i style:width="{(c.progreso ?? 0) * 100}%"></i></div>
              <p class="foot">{Math.round((c.progreso ?? 0) * 100)} % de {c.meta_nombre || "tu meta"} ({pesos(c.meta_monto)}){c.meta_fecha ? ` · a este ritmo la alcanzas el ${fecha(c.meta_fecha)}` : ""}</p>
            {:else}
              <p class="foot">Ponle una meta para ver cuándo la alcanzas.</p>
            {/if}
            <p class="note tight">{c.ritmo_semanal ? `Ahorras ~${pesos(c.ritmo_semanal)} por semana aquí.` : "Todavía sin ritmo semanal."}{c.rendimiento_anual ? ` Rinde ${c.rendimiento_anual} % anual (~${pesos((c.saldo * c.rendimiento_anual) / 100 / 12)} al mes).` : ""}</p>
            <div class="actions">
              <button class="btn" onclick={() => abrir({ k: "caja-mov", c })}>Mover dinero</button>
              <button class="btn" onclick={() => abrir({ k: "caja", c })}>Editar</button>
            </div>
          </GlassCard>
        {/each}
        <button class="btn block dashed" onclick={() => abrir({ k: "caja-nueva" })}>+ Nueva caja de ahorro</button>

        {#if r.fondo_plexus}
          <GlassCard padding="20px">
            <div class="card-head">
              <h2>Fondo de ahorro Plexus<span class="mini-tag">{r.fondo_plexus.cobrado ? "cobrado ✓" : "por cobrar"}</span></h2>
              <b class="total">{pesos(r.fondo_plexus.fondo_saldo_total)}</b>
            </div>
            <p class="foot">Tú {pesos(r.fondo_plexus.fondo_tuyo)} + empresa {pesos(r.fondo_plexus.fondo_empresa)} · crece ~{pesos(r.fondo_plexus.aporte_semanal)} por semana (la empresa pone lo mismo que tú).</p>
            {#if !r.fondo_plexus.cobrado}
              <p class="note tight">Se va acumulando; RH dice que lo entregan en noviembre. Hasta entonces no cuenta como dinero tuyo disponible.</p>
              <div class="actions"><button class="btn primary" onclick={() => abrir({ k: "extra", concepto: "fondo_plexus", monto: r.fondo_plexus?.fondo_saldo_total ?? null })}>Ya me lo pagaron</button></div>
            {/if}
            {#if fondoBars.length}
              <Bars label="Fondo de ahorro acumulado por semana" bars={fondoBars} series={[{ name: "Tuyo", color: "var(--c-action)" }, { name: "Empresa", color: "var(--text)" }]} format={(n) => pesos(n)} height={110} tickEvery={3} />
            {/if}
            <p class="note">Se descuenta de tu nómina antes de que te llegue, así que no cuenta en tu reparto semanal. Se lee solo de tus recibos.</p>
          </GlassCard>
        {/if}

        {#if r.aguinaldo}
          <GlassCard padding="20px">
            <div class="card-head">
              <h2>Aguinaldo<span class="mini-tag">{r.aguinaldo.cobrado ? "cobrado ✓" : `antes del ${fecha(r.aguinaldo.fecha_limite)}`}</span></h2>
              <b class="total">≈ {pesos(r.aguinaldo.estimado, 0)}</b>
            </div>
            <p class="foot">{r.aguinaldo.dias} días de salario ({pesos(r.aguinaldo.salario_diario)}/día) proporcional a lo que llevas desde el {fecha(r.aguinaldo.desde)}. Es un estimado antes de impuestos; puede variar.</p>
            {#if !r.aguinaldo.cobrado}
              <div class="actions"><button class="btn" onclick={() => abrir({ k: "extra", concepto: "aguinaldo", monto: null })}>Ya me lo pagaron</button></div>
            {/if}
          </GlassCard>
        {/if}
      </div>

    <!-- ============================ NOMINA ============================ -->
    {:else}
      <div class="stack board stagger">
        {#if ultima}
          <GlassCard padding="22px">
            <span class="label">Semana {ultima.semana} · pagada {fecha(ultima.fecha_pago)}</span>
            <div class="hero"><span class="hero-num">{pesos(ultima.neto)}</span><span class="hero-of">de {pesos(ultima.percepciones)} brutos</span></div>
            <div class="waterfall" aria-hidden="true">
              {#each tramos(ultima) as t (t.k)}<i class="wf-{t.k}" style:flex-grow={t.v}></i>{/each}
            </div>
            <ul class="wf-legend">
              {#each tramos(ultima) as t (t.k)}
                <li><i class="wf-{t.k}"></i><span>{t.label}</span><b>{pesos(t.v)}</b></li>
              {/each}
              <li class="desp"><i></i><span>Vales de despensa → casa</span><b>{pesos(ultima.despensa)}</b></li>
            </ul>
            {#if ultima.nota}<p class="note">{ultima.nota}</p>{/if}
          </GlassCard>

          <GlassCard padding="20px">
            <div class="card-head"><h2>Neto por semana</h2><span class="total">prom. {pesos(promedioNeto, 0)}</span></div>
            <Bars label="Neto pagado por semana" bars={nominaBars} series={[{ name: "Neto", color: "var(--c-action)" }]} format={(n) => pesos(n)} height={110} tickEvery={3} />
          </GlassCard>

          <GlassCard padding="20px">
            <div class="card-head"><h2>Recibos</h2><span class="total">{$nominas.length}</span></div>
            <ul class="movs">
              {#each $nominas as n (n.uuid)}
                <li>
                  <span class="mv-main"><b>Semana {n.semana}</b><span>{fecha(n.periodo_ini)} – {fecha(n.periodo_fin)}{n.nota ? ` · ${n.nota.split(";")[0].slice(0, 60)}` : ""}</span></span>
                  <span class="mv-amt">{pesos(n.neto)}</span>
                </li>
              {/each}
            </ul>
            <p class="note">Mándale tus recibos a KORRA por Telegram: los acomoda en PLEXUS y aquí aparecen solos (revisa cada hora).</p>
          </GlassCard>
        {:else}
          <div class="empty glass"><p>Todavía no hay recibos de nómina.</p></div>
        {/if}
      </div>
    {/if}

    {#if seccion === "semana"}
      <button class="fab" aria-label="Nuevo gasto" onclick={() => abrir({ k: "gasto" })}>
        <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="currentColor" stroke-width="2.6" stroke-linecap="round"><path d="M12 5v14M5 12h14" /></svg>
      </button>
    {/if}
  {/if}
</div>

{#if r?.iniciado}
  <button class="cashy-btn" aria-label="Hablar con Cashy, tu asesora de dinero" onclick={() => (cashyAbierta = true)}>
    <CashyCara gesto={gestoCashy} size={78} />
  </button>
{/if}

<Sheet bind:open={cashyAbierta} title="Pregúntale a Cashy">
  <CashyChat />
</Sheet>

<Sheet bind:open={hojaAbierta} title={hoja ? TITULOS[hoja.k] : ""}>
  {#if hoja?.k === "gasto"}<GastoForm onDone={listo} />
  {:else if hoja?.k === "pago"}<PagoForm onDone={listo} />
  {:else if hoja?.k === "inicio"}<InicioForm onDone={listo} />
  {:else if hoja?.k === "pagar"}<PagarTarjetaForm tarjeta={hoja.t} onDone={listo} />
  {:else if hoja?.k === "tarjeta"}<TarjetaForm tarjeta={hoja.t} onDone={listo} />
  {:else if hoja?.k === "caja"}<CajaForm caja={hoja.c} onDone={listo} />
  {:else if hoja?.k === "caja-nueva"}<CajaForm onDone={listo} />
  {:else if hoja?.k === "caja-mov"}<CajaMovForm caja={hoja.c} onDone={listo} />
  {:else if hoja?.k === "ajuste"}<AjusteForm tarjeta={hoja.t} onDone={listo} />
  {:else if hoja?.k === "plan"}<PlanForm tarjeta={hoja.t} onDone={listo} />
  {:else if hoja?.k === "extra"}<ExtraForm concepto={hoja.concepto} montoInicial={hoja.monto} onDone={listo} />
  {/if}
</Sheet>

<ConfirmDialog bind:open={confirmarBorrar} title="¿Borrar?" message="Se quita y se recalculan tus saldos." confirmLabel="Borrar" danger onConfirm={borrar} />

<style>
  .stack {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  /* iPad horizontal / compu: 2 columnas (clase global .board en app.css). Lo principal de
     cada seccion (el primer bloque), los 3 saldos y los botones/notas ocupan todo el ancho. */
  .board > :global(:first-child),
  .board > .tri,
  .board > .note,
  .board > .dashed {
    grid-column: 1 / -1;
  }

  @media (min-width: 1000px) {
    .seg-wrap,
    .alertas {
      max-width: 640px;
    }
    .alertas {
      margin-left: 0;
    }
  }

  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 40px 24px;
    text-align: center;
    color: var(--text-secondary);
  }

  .big-h {
    margin: 0 0 8px;
    font-family: var(--font-display);
    font-size: 24px;
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .lead {
    margin: 0 0 18px;
    font-size: 15px;
    line-height: 1.5;
    color: var(--text-secondary);
  }

  .lead b {
    color: var(--text);
  }

  .toast {
    position: fixed;
    left: 50%;
    top: calc(14px + var(--safe-top));
    z-index: 60;
    transform: translateX(-50%);
    width: max-content;
    max-width: calc(100% - 32px);
    padding: 12px 18px;
    border-radius: 16px;
    background: var(--text);
    color: var(--bg);
    font-size: 14px;
    font-weight: 700;
    box-shadow: 3px 3px 0 var(--c-action);
    animation: rise var(--dur-base) var(--ease-out) both;
  }

  .alertas {
    list-style: none;
    margin: 0 0 14px;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .alertas li {
    display: flex;
    gap: 10px;
    align-items: flex-start;
    padding: 12px 14px;
    border-radius: 16px;
    border: 1.5px solid var(--border-strong);
    font-size: 13px;
    font-weight: 600;
    line-height: 1.4;
  }

  .alertas li svg {
    flex-shrink: 0;
    margin-top: 1px;
    color: var(--text-secondary);
  }

  .alertas li.alto {
    border-color: var(--c-action);
    background: color-mix(in srgb, var(--c-action) 12%, transparent);
  }

  .alertas li.alto svg {
    color: var(--c-action);
  }

  .seg-wrap {
    margin-bottom: 14px;
  }

  .seg {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    padding: 4px;
    border-radius: 99em;
    border: 1.5px solid var(--border-strong);
  }

  .seg button {
    min-height: 38px;
    border: none;
    border-radius: 99em;
    background: none;
    font: inherit;
    font-size: 13px;
    font-weight: 800;
    color: var(--text-secondary);
    cursor: pointer;
    touch-action: manipulation;
    transition: background-color var(--dur-fast) ease, color var(--dur-fast) ease;
  }

  .seg button[aria-selected="true"] {
    background: var(--text);
    color: var(--bg);
  }

  .seg button:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 1px;
  }

  .label {
    display: block;
    font-size: 11px;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--text-secondary);
  }

  .hero {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 4px 10px;
    margin: 4px 0 12px;
  }

  .hero-num {
    font-family: var(--font-display);
    font-size: clamp(44px, 14vw, 60px);
    font-weight: 800;
    letter-spacing: -0.04em;
    line-height: 1;
    font-variant-numeric: tabular-nums;
  }

  .hero.over .hero-num {
    color: var(--c-action);
  }

  .hero-of {
    font-size: 15px;
    font-weight: 700;
    color: var(--text-secondary);
  }

  .meter {
    height: 12px;
    border-radius: 99em;
    background: color-mix(in srgb, var(--text) 9%, transparent);
    overflow: hidden;
  }

  .meter.small {
    height: 8px;
    margin: 8px 0 4px;
  }

  .meter i {
    display: block;
    height: 100%;
    border-radius: 99em;
    background: var(--text);
    transform-origin: left;
    animation: grow-x var(--dur-slow) var(--ease-out) both;
  }

  .hero.over + .meter i,
  .meter.over i {
    background: var(--c-action);
  }

  @keyframes grow-x {
    from {
      transform: scaleX(0);
    }
  }

  .foot {
    margin: 10px 0 0;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .actions {
    display: flex;
    gap: 10px;
    margin-top: 16px;
  }

  .actions .btn {
    flex: 1;
  }

  .pulse {
    border-color: var(--c-action);
    color: var(--c-action);
  }

  .tri {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
  }

  .box {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
    padding: 14px 12px;
    border-radius: 20px;
    animation: rise var(--dur-slow) var(--ease-out) both;
  }

  .box .label {
    font-size: 10px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .box b {
    font-family: var(--font-display);
    font-size: 20px;
    font-weight: 800;
    letter-spacing: -0.02em;
    font-variant-numeric: tabular-nums;
    overflow-wrap: anywhere;
  }

  .sub {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .sub.warn {
    color: var(--c-action);
    font-weight: 700;
  }

  .card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 12px;
  }

  h2 {
    margin: 0;
    font-size: 17px;
    font-weight: 800;
    letter-spacing: -0.02em;
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
  }

  .total {
    font-size: 14px;
    font-weight: 700;
    color: var(--text-secondary);
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  b.total {
    font-family: var(--font-display);
    font-size: 20px;
    color: var(--text);
  }

  b.total.neg {
    color: var(--c-action);
  }

  .link {
    border: none;
    background: none;
    padding: 6px 0;
    font: inherit;
    font-size: 12px;
    font-weight: 700;
    color: var(--c-action);
    cursor: pointer;
  }

  .block-link {
    display: block;
    width: 100%;
    margin-top: 8px;
    text-align: center;
  }

  .meses {
    margin-top: 16px;
    padding-top: 12px;
    border-top: 1px solid var(--border);
  }

  .meses-head {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    font-weight: 800;
    color: var(--text-secondary);
  }

  .meses-head b {
    color: var(--text);
  }

  .plan {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 0;
  }

  .plan-main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    font-size: 13px;
  }

  .plan-main span {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .mini-tag {
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    padding: 3px 8px;
    border-radius: 99em;
    color: var(--c-action);
    border: 1px solid var(--c-action);
  }

  .note {
    margin: 12px 0 0;
    font-size: 12px;
    line-height: 1.45;
    color: var(--text-tertiary);
  }

  .note.tight {
    margin-top: 6px;
  }

  .note.center {
    text-align: center;
    padding: 0 8px;
  }

  .hbars {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .hbars li {
    display: grid;
    grid-template-columns: 6.5em 1fr auto;
    align-items: center;
    gap: 4px 10px;
    font-size: 13px;
  }

  .hb-name {
    color: var(--text-secondary);
  }

  .hb-bar {
    height: 10px;
  }

  .hb-bar i {
    display: block;
    height: 100%;
    min-width: 4px;
    border-radius: 0 4px 4px 0;
    background: var(--c-action);
  }

  .hbars b {
    font-variant-numeric: tabular-nums;
  }

  .hb-sub {
    grid-column: 2 / -1;
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .filtro {
    flex-wrap: nowrap;
    overflow-x: auto;
    scrollbar-width: none;
    margin: 0 -4px 10px;
    padding: 0 4px 2px;
  }

  .filtro::-webkit-scrollbar {
    display: none;
  }

  .filtro button {
    flex-shrink: 0;
    min-height: 32px;
    font-size: 12px;
  }

  .movs {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .movs li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 0;
    border-top: 1px solid var(--border);
  }

  .movs li:first-child {
    border-top: none;
  }

  .mv-main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .mv-main b {
    font-size: 14px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .mv-main span {
    font-size: 12px;
    color: var(--text-tertiary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .mv-amt {
    font-weight: 800;
    font-variant-numeric: tabular-nums;
  }

  .mv-amt.pos {
    color: var(--c-action);
  }

  .mv-del {
    width: 34px;
    height: 34px;
    display: grid;
    place-items: center;
    flex-shrink: 0;
    border: none;
    border-radius: 50%;
    background: none;
    color: var(--text-tertiary);
    cursor: pointer;
  }

  .mv-del:hover,
  .mv-del:focus-visible {
    background: color-mix(in srgb, var(--text) 8%, transparent);
    color: var(--text);
  }

  /* 72 px; la cara (78 px) se sale un poco del circulo: el pelo esponjado "rebasa" el boton. */
  .cashy-btn {
    position: fixed;
    left: 16px;
    bottom: calc(94px + var(--safe-bottom));
    z-index: 35;
    width: 72px;
    height: 72px;
    display: grid;
    place-items: center;
    padding: 0;
    border: 2.5px solid #141317;
    border-radius: 50%;
    background: radial-gradient(circle at 50% 40%, #ffd3e4, #ff4f9a);
    color: #141317;
    animation: rebote 3.2s var(--ease-out) infinite;
    box-shadow: 3px 3px 0 var(--c-action), 0 10px 24px rgba(0, 0, 0, 0.3);
    cursor: pointer;
    touch-action: manipulation;
    transition: transform var(--dur-press) var(--ease-out);
  }

  .cashy-btn:active {
    transform: scale(0.94);
  }

  /* un saltito cada tanto para que se note que esta "viva" */
  @keyframes rebote {
    0%,
    82%,
    100% {
      transform: translateY(0);
    }
    88% {
      transform: translateY(-6px);
    }
    94% {
      transform: translateY(0);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .cashy-btn {
      animation: none;
    }
  }

  .cashy-btn:focus-visible {
    outline: 2px solid var(--c-action);
    outline-offset: 3px;
  }

  @media (min-width: 640px) {
    .cashy-btn {
      bottom: 32px;
    }
  }

  .fab {
    position: fixed;
    right: 18px;
    bottom: calc(96px + var(--safe-bottom));
    z-index: 35;
    width: 58px;
    height: 58px;
    display: grid;
    place-items: center;
    border: none;
    border-radius: 50%;
    background: var(--c-action);
    color: #fff;
    box-shadow: 3px 3px 0 var(--panel-offset), 0 10px 24px color-mix(in srgb, var(--c-action) 35%, transparent);
    cursor: pointer;
    touch-action: manipulation;
    transition: transform var(--dur-press) var(--ease-out);
  }

  .fab:active {
    transform: scale(0.94);
  }

  .fab:focus-visible {
    outline: 2px solid var(--text);
    outline-offset: 3px;
  }

  /* ---- tarjetas ---- */
  .fondo-line {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 2px 10px;
    padding: 14px 16px;
    border-radius: 20px;
    font-size: 13px;
    font-weight: 700;
    color: var(--text-secondary);
    animation: rise var(--dur-slow) var(--ease-out) both;
  }

  .fondo-line b {
    font-family: var(--font-display);
    font-size: 20px;
    color: var(--text);
  }

  .fondo-line .sub {
    grid-column: 1 / -1;
  }

  .cc {
    border-radius: 24px;
    overflow: hidden;
    border: 1.5px solid var(--glass-border);
    background: var(--bg-elevated);
    box-shadow: 4px 4px 0 var(--panel-offset);
    animation: rise var(--dur-slow) var(--ease-out) both;
  }

  .cc-face {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 18px 20px;
    min-height: clamp(170px, 46vw, 220px); /* llena el ancho de su columna, sin pasar de alto */
    color: #fff;
    background: #0b0b0c;
  }

  .cc[data-card="nu"] .cc-face {
    background: linear-gradient(135deg, #820ad1, #4a0679);
  }

  .cc[data-card="plata"] .cc-face {
    background: linear-gradient(135deg, #2b2b2e, #0b0b0c);
  }

  .cc.off .cc-face {
    filter: grayscale(0.85);
  }

  .cc-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: auto;
  }

  .cc-name {
    font-family: var(--font-display);
    font-size: 22px;
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .cc-tag {
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    padding: 4px 9px;
    border-radius: 99em;
    background: rgba(255, 255, 255, 0.16);
  }

  .cc-label {
    font-size: 11px;
    font-weight: 700;
    opacity: 0.75;
  }

  .cc-amount {
    font-family: var(--font-display);
    font-size: 38px;
    font-weight: 800;
    letter-spacing: -0.03em;
    line-height: 1.05;
    font-variant-numeric: tabular-nums;
  }

  .cc-bottom {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    margin-top: 6px;
    font-size: 12px;
    font-weight: 600;
    opacity: 0.85;
  }

  .cc-num {
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.08em;
    white-space: nowrap;
  }

  .cc-info {
    padding: 16px 20px 18px;
  }

  .cc-row {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .cc-row b {
    color: var(--text);
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .cc-facts {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    margin: 14px 0 0;
  }

  .cc-facts dt {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .cc-facts dd {
    margin: 1px 0 0;
    font-weight: 800;
    font-variant-numeric: tabular-nums;
  }

  .dashed {
    border-style: dashed;
    color: var(--text-secondary);
  }

  /* ---- nomina ---- */
  .waterfall {
    display: flex;
    gap: 2px;
    height: 18px;
    margin: 4px 0 14px;
    border-radius: 6px;
    overflow: hidden;
  }

  .waterfall i {
    flex-basis: 0;
    min-width: 3px;
  }

  .wf-neto {
    background: var(--c-action);
  }
  .wf-fondo {
    background: var(--text);
  }
  .wf-isr {
    background: color-mix(in srgb, var(--text) 55%, var(--bg-elevated));
  }
  .wf-imss {
    background: color-mix(in srgb, var(--text) 38%, var(--bg-elevated));
  }
  .wf-cafe {
    background: color-mix(in srgb, var(--text) 24%, var(--bg-elevated));
  }
  .wf-otros {
    background: color-mix(in srgb, var(--text) 14%, var(--bg-elevated));
  }

  .wf-legend {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .wf-legend li {
    display: grid;
    grid-template-columns: 10px 1fr auto;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .wf-legend li i {
    width: 10px;
    height: 10px;
    border-radius: 3px;
  }

  .wf-legend li b {
    color: var(--text);
    font-variant-numeric: tabular-nums;
  }

  .wf-legend li.desp {
    padding-top: 8px;
    border-top: 1px dashed var(--border-strong);
  }

  .wf-legend li.desp i {
    border: 1.5px dashed var(--text-tertiary);
  }

  @media (min-width: 640px) {
    .fab {
      bottom: 32px;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .meter i {
      animation: none;
    }
  }
</style>
