// Espejo de los tipos JSON que expone el backend (internal/metrics, internal/dockerapi).

export interface DiskUsage {
  mount_point: string;
  fs_type: string;
  total_bytes: number;
  free_bytes: number;
  used_bytes: number;
  used_percent: number;
}

export interface MemInfo {
  total_kb: number;
  available_kb: number;
  used_kb: number;
  used_percent: number;
  swap_total_kb: number;
  swap_used_kb: number;
  zram_active: boolean;
}

export interface ThrottleStatus {
  available: boolean;
  raw?: string;
}

export interface NetSnapshot {
  rx_bytes_per_sec: Record<string, number>;
  tx_bytes_per_sec: Record<string, number>;
  latency_ms: number;
  reachable: boolean;
}

export interface MetricsSnapshot {
  timestamp: string;
  cpu_percent: number;
  mem: MemInfo;
  disks: DiskUsage[];
  temp_c: number;
  throttle: ThrottleStatus;
  net: NetSnapshot;
  uptime_sec: number;
}

export interface ContainerPort {
  private_port: number;
  public_port?: number;
  type: string;
}

export interface ContainerStats {
  cpu_percent: number;
  mem_usage_bytes: number;
  mem_limit_bytes: number;
  mem_percent: number;
}

export interface ContainerSummary {
  id: string;
  name: string;
  image: string;
  state: string;
  status: string;
  created: number;
  ports: ContainerPort[];
  stats?: ContainerStats;
  project: string;
  protected: boolean;
  adapter: string;
  adapter_data?: unknown;
}

export interface DockerEvent {
  action: string;
  id: string;
  name?: string;
  time: number;
}

export interface SpeedtestResult {
  download_mbps: number;
  upload_mbps: number;
  latency_ms: number;
  server: string;
}

// Resumen que escribe el asistente KORRA cada 15 min (ver ~/asistente/bot/estadisticas.py).
export interface KorraDay {
  mensajes: number;
  tokens_deepseek: number;
  tokens_claude: number;
  costo_deepseek: number;
  costo_claude: number;
  trabajos: number;
}

export interface KorraMonth {
  mes: string; // "AAAA-MM"
  mensajes: number;
  deepseek: { entrada: number; cache: number; salida: number; costo: number; llamadas: number };
  claude: { entrada: number; cache: number; salida: number; costo_equiv: number; trabajos: number };
}

export interface KorraStats {
  version: number;
  generado: number;
  hoy: string;
  dias: Record<string, KorraDay>;
  meses: KorraMonth[];
  horas: number[];
  dias_semana: number[];
  herramientas: { nombre: string; veces: number }[];
  modelos: { modelo: string; tokens: number }[];
  totales: {
    mensajes: number;
    respuestas: number;
    dias_activos: number;
    racha_actual: number;
    racha_max: number;
    dia_record: { fecha: string | null; mensajes: number };
    recuerdos: number;
    tareas_hechas: number;
    tareas_pendientes: number;
    trabajos_claude: number;
    trabajos_claude_error: number;
    costo_claude_equiv: number;
    costo_deepseek: number;
    tokens_por_respuesta_deepseek: number;
    desde: string | null;
    deepseek_desde: string | null;
  };
  saldo_deepseek: { total: number; moneda: string } | null;
}

// API de dinero (~/task-hub) via /api/finanzas. Todo el dinero en pesos.
export interface FinTarjeta {
  id: string;
  nombre: string;
  banco: string;
  ultimos4: string;
  estado: "activa" | "pausada" | "bloqueada";
  nota: string;
  dia_corte: number;
  dia_pago: number;
  limite_banco: number;
  tope_personal: number;
  ultimo_corte: string;
  proximo_corte: string;
  fecha_limite: string;
  dias_para_pagar: number;
  a_pagar: number;
  ciclo_actual: number;
  deuda: number;
  disponible_banco: number | null;
  uso_tope: number | null;
  meses: { id: number; descripcion: string; mensualidad: number; ultimo_mes: string; pendiente: number; restante: number; pagos: string[] }[];
  meses_pendiente: number;
  meses_restante: number;
  calendario: FinEstadoMes[];
}

/** Un estado de cuenta: el ya cerrado, el que va en curso o uno futuro (solo mensualidades a meses). */
export interface FinEstadoMes {
  fecha_limite: string;
  corte: string;
  estado: "cerrado" | "en_curso" | "futuro";
  monto: number;
  monto_meses: number;
  detalle: { descripcion: string; mensualidad: number; ultimo: boolean; faltan: number }[];
}

export interface FinCaja {
  id: number;
  nombre: string;
  saldo: number;
  meta_nombre: string;
  meta_monto: number;
  rendimiento_anual: number;
  es_destino: boolean;
  progreso: number | null;
  ritmo_semanal: number;
  meta_fecha: string | null;
}

export interface FinResumen {
  hoy: string;
  mensaje: string;
  iniciado: string | null;
  semana: { inicio: string; proximo_pago: string; dias_para_pago: number; limite: number; gastado: number };
  efectivo: number;
  fondo: number;
  ahorro_total: number;
  deuda_total: number;
  a_pagar_total: number;
  patrimonio: number;
  tarjetas: FinTarjeta[];
  cajas: FinCaja[];
  racha_ahorro: number;
  categorias_semana: Record<string, number>;
  categorias_mes: Record<string, number>;
  semanas: { inicio: string; gastado: number }[];
  alertas: { nivel: "alto" | "medio"; texto: string }[];
  nomina_pendiente: { uuid: string; semana: number; fecha_pago: string; neto: number; despensa: number } | null;
  fondo_plexus: { fondo_tuyo: number; fondo_empresa: number; fondo_saldo_total: number; fecha_pago: string; aporte_semanal: number; cobrado: boolean } | null;
  iniciales: { efectivo: number; fondo: number; ahorro: number; deudas: Record<string, { corte: number; ciclo: number }> };
  aguinaldo: { dias: number; estimado: number; salario_diario: number; desde: string; fecha_limite: string; cobrado: boolean } | null;
  categorias: string[];
  metodos: string[];
  sobrante_sugerido: number;
}

export interface FinReparto {
  monto: number;
  sobrante: number;
  gastos: number;
  fondo: number;
  ahorro: number;
  sobrante_a_ahorro: number;
  retirar: number;
}

export interface FinMovimiento {
  id: number;
  ts: number;
  tipo: string;
  monto: number;
  metodo: string;
  categoria: string;
  nota: string;
  tarjeta: string;
  caja: number | null;
  caja_destino: number | null;
  retirar: number | null;
}

export interface FinConcepto {
  clave: string;
  concepto: string;
  monto: number;
}

export interface FinNomina {
  uuid: string;
  semana: number;
  periodo_ini: string;
  periodo_fin: string;
  fecha_pago: string;
  percepciones: number;
  deducciones: number;
  neto: number;
  isr: number;
  imss: number;
  fondo_ahorro: number;
  cafeteria: number;
  despensa: number;
  fondo_tuyo: number;
  fondo_empresa: number;
  fondo_saldo_total: number;
  conceptos: { percepciones: FinConcepto[]; deducciones: FinConcepto[] };
  nota: string;
}

export interface CashyMensaje {
  id: number;
  ts: number;
  rol: "chava" | "cashy";
  texto: string;
  estado: "ok" | "pensando" | "error";
}

export interface CashyHistoria {
  mensajes: CashyMensaje[];
  pensando: boolean;
  modelo: string;
  sesion: boolean;
}

export interface FinPeriodo {
  vista: "dia" | "semana" | "mes";
  inicio: string;
  fin: string;
  total: number;
  anterior_total: number;
  promedio_diario: number;
  por_categoria: Record<string, number>;
  por_metodo: Record<string, number>;
  barras: { fecha: string; total: number }[];
  lista: { id: number; ts: number; monto: number; metodo: string; categoria: string; nota: string; tipo: string }[];
  anterior: string;
  siguiente: string | null;
}
