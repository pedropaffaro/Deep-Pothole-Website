interface StateTagProps {
  state: string;
}

const CORES: Record<string, string> = {
  novo: "bg-red-500/20 border-red-400/60 text-red-100",
  "em análise": "bg-orange-400/20 border-orange-300/60 text-orange-100",
  analisado: "bg-sky-400/20 border-sky-300/60 text-sky-100",
  concertando: "bg-emerald-400/20 border-emerald-300/60 text-emerald-100",
};

const PADRAO = "bg-white/10 border-white/30 text-white/80";

function StateTag({ state }: StateTagProps) {
  return (
    <span
      className={`inline-flex items-center shrink-0 px-3 py-1 rounded-full border text-sm font-bold tracking-[0.1em] uppercase ${
        CORES[state] ?? PADRAO
      }`}
    >
      {state || "sem estado"}
    </span>
  );
}

export default StateTag;
