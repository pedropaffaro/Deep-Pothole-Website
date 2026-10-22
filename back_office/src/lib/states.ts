export interface StateInfo {
  value: string;
  // Cor do pin no mapa, alinhada com a cor da StateTag
  pin: string;
}

export const STATES: StateInfo[] = [
  { value: "novo", pin: "#EF4444" },
  { value: "em análise", pin: "#FB923C" },
  { value: "analisado", pin: "#38BDF8" },
  { value: "concertando", pin: "#34D399" },
];

export const DEFAULT_PIN_COLOR = "#2A48A3";

export function pinColor(state: string): string {
  return STATES.find((s) => s.value === state)?.pin ?? DEFAULT_PIN_COLOR;
}
