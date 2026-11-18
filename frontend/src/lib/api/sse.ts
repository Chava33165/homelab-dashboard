// Un unico EventSource multiplexado (event: metrics | containers | docker_event | alert)
// contra /api/stream. Los navegadores limitan a 6 conexiones HTTP/1.1 concurrentes por
// origen -- multiplexar todo en un solo stream evita agotarlas. Los logs de un
// contenedor (api/logs.ts) son la unica conexion SSE aparte, y solo mientras esa vista
// de detalle esta montada.

type Handler<T> = (data: T) => void;

class MainStream {
  private source: EventSource | null = null;
  private handlers = new Map<string, Set<Handler<any>>>();
  private reconnectDelay = 1000;

  connect() {
    if (this.source) return;
    this.source = new EventSource("/api/stream");
    for (const [event, set] of this.handlers) {
      this.attach(event, set);
    }
    this.source.onopen = () => {
      this.reconnectDelay = 1000;
    };
    this.source.onerror = () => {
      // EventSource ya reintenta solo, pero si el server tumbo la sesion (401 no se
      // puede detectar por EventSource, llega como error generico) forzamos un ciclo
      // de reconexion con backoff simple para no martillar al server.
      this.source?.close();
      this.source = null;
      setTimeout(() => this.connect(), this.reconnectDelay);
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, 15000);
    };
  }

  disconnect() {
    this.source?.close();
    this.source = null;
  }

  on<T>(event: string, handler: Handler<T>) {
    let set = this.handlers.get(event);
    if (!set) {
      set = new Set();
      this.handlers.set(event, set);
      // Un solo listener por evento y por conexion: recorre el Set en vivo, asi que los
      // handlers que se agreguen despues ya quedan cubiertos. Volver a hacer attach aqui
      // en cada on() duplicaba las llamadas (ej. cada "X se unio" de Terraria llegaba
      // N veces tras entrar N veces a la vista).
      if (this.source) this.attach(event, set);
    }
    set.add(handler);
    return () => set!.delete(handler);
  }

  private attach(event: string, set: Set<Handler<any>>) {
    this.source?.addEventListener(event, (e: MessageEvent) => {
      let data: unknown;
      try {
        data = JSON.parse(e.data);
      } catch {
        return;
      }
      for (const h of set) h(data);
    });
  }
}

export const mainStream = new MainStream();

// Stream de logs de UN contenedor -- conexion aparte, se abre/cierra junto con la vista
// de detalle (ver views/ContainerDetail.svelte).
//
// onOpen corre en cada (re)conexion: el server siempre reenvia las ultimas 200 lineas al
// conectar, y el proxy de Docker corta el stream cada ~10 min -- sin limpiar ahi, cada
// reconexion repetia esas 200 lineas debajo de las que ya se veian.
export function openLogsStream(containerId: string, onLine: (line: string) => void, onOpen: () => void): () => void {
  const source = new EventSource(`/api/containers/${containerId}/logs/stream`);
  source.onopen = onOpen;
  source.addEventListener("log", (e: MessageEvent) => {
    try {
      onLine(JSON.parse(e.data));
    } catch {
      /* ignorar linea malformada */
    }
  });
  return () => source.close();
}
