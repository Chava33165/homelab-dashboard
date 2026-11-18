// Mapa slug -> componente de panel especifico. La vista de detalle de contenedor
// (ContainerDetail.svelte) renderiza el componente que corresponda segun el campo
// "adapter" que devuelve /api/containers -- si no hay entrada para ese slug (ej.
// "generic", o un adapter nuevo del backend sin panel de frontend todavia), no se
// muestra nada extra y el contenedor se ve con la info generica de siempre.
//
// Para agregar un adapter nuevo: crear el .svelte del panel y agregar una linea aqui.
// Ver docs/adapters.md.

import type { Component } from "svelte";
import TerrariaPanel from "./TerrariaPanel.svelte";

export const adapterRegistry: Record<string, Component<any>> = {
  terraria: TerrariaPanel,
};
