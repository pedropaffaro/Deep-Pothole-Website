import { useEffect, useMemo, useState } from "react";

import Navbar from "./components/Navbar";
import Spinner from "./components/Spinner";
import ComplaintsMap from "./components/ComplaintsMap";
import ComplaintDetails from "./components/ComplaintDetails";
import AnalyzedQueue from "./components/AnalyzedQueue";
import { deleteComplaint, fetchComplaints, updateComplaintState } from "./lib/api";
import type { Complaint } from "./lib/api";

const QUEUE_STORAGE_KEY = "backoffice:fila-analisados";
const ANALYZED = "analisado";

// A ordem de entrada na fila não existe no banco, então fica salva neste navegador
function loadQueue(): number[] {
  try {
    const raw = localStorage.getItem(QUEUE_STORAGE_KEY);
    const parsed = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? parsed.filter((id) => typeof id === "number") : [];
  } catch {
    return [];
  }
}

function saveQueue(queue: number[]) {
  try {
    localStorage.setItem(QUEUE_STORAGE_KEY, JSON.stringify(queue));
  } catch {
    // sem armazenamento: a fila só vale até recarregar a página
  }
}

function App() {
  const [complaints, setComplaints] = useState<Complaint[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [erro, setErro] = useState<string | null>(null);
  const [queueOrder, setQueueOrder] = useState<number[]>(loadQueue);

  useEffect(() => {
    let ativo = true;

    const carregar = async () => {
      try {
        const data = await fetchComplaints();
        if (!ativo) return;
        setComplaints(data);
        setErro(null);
      } catch (error) {
        console.error("Erro ao carregar denúncias:", error);
        if (!ativo) return;
        setErro("Não foi possível carregar as denúncias. O backend está rodando?");
      } finally {
        if (ativo) setIsLoading(false);
      }
    };

    carregar();

    return () => {
      ativo = false;
    };
  }, []);

  // Só existe uma denúncia inspecionada por vez: clicar em outra substitui a anterior
  const selected = complaints.find((c) => c.id === selectedId) ?? null;

  // Analisados na ordem de chegada; os que a ordem salva não conhece (por exemplo,
  // analisados em outro navegador) entram no fim, do mais antigo para o mais novo
  const queue = useMemo(() => {
    const analisados = complaints.filter((c) => c.state === ANALYZED);
    const porId = new Map(analisados.map((c) => [c.id, c]));

    const ordenados = queueOrder
      .map((id) => porId.get(id))
      .filter((c): c is Complaint => c !== undefined);

    const naFila = new Set(ordenados.map((c) => c.id));
    const restantes = analisados
      .filter((c) => !naFila.has(c.id))
      .sort((a, b) => a.id - b.id);

    return [...ordenados, ...restantes];
  }, [complaints, queueOrder]);

  const updateQueueOrder = (atualizar: (atual: number[]) => number[]) => {
    setQueueOrder((atual) => {
      const novo = atualizar(atual);
      saveQueue(novo);
      return novo;
    });
  };

  // O arrasto devolve a ordem exibida inteira, inclusive os itens sem ordem salva
  const handleReorderQueue = (ids: number[]) => updateQueueOrder(() => ids);

  const handleDelete = async (complaint: Complaint) => {
    try {
      await deleteComplaint(complaint.id);
      setComplaints((atuais) => atuais.filter((c) => c.id !== complaint.id));
      setSelectedId(null);
      updateQueueOrder((atual) => atual.filter((id) => id !== complaint.id));
      setErro(null);
    } catch (error) {
      console.error("Erro ao apagar denúncia:", error);
      setErro(`Não foi possível apagar a denúncia #${complaint.id}.`);
    }
  };

  const handleChangeState = async (complaint: Complaint, state: string) => {
    try {
      const atualizada = await updateComplaintState(complaint.id, state);
      setComplaints((atuais) =>
        atuais.map((c) => (c.id === atualizada.id ? atualizada : c))
      );
      // Virar analisado coloca no fim da fila; sair de analisado tira da fila
      updateQueueOrder((atual) => {
        const semEste = atual.filter((id) => id !== atualizada.id);
        return atualizada.state === ANALYZED ? [...semEste, atualizada.id] : semEste;
      });
      setErro(null);
    } catch (error) {
      console.error("Erro ao atualizar estado:", error);
      setErro(`Não foi possível atualizar o estado da denúncia #${complaint.id}.`);
    }
  };

  return (
    <div className="min-h-screen bg-linear-to-b from-[#2a3c6b] to-[#1e293b] font-sans pb-10">
      <Navbar />

      {isLoading && <Spinner />}

      <main className="max-w-[480px] lg:max-w-6xl mx-auto px-6 pt-32">
        <div className="flex flex-col lg:flex-row lg:items-end lg:justify-between gap-2 mb-8">
          <div className="flex flex-col">
            <span className="text-lg font-bold tracking-[0.2em] text-white/80 uppercase mb-2">
              Painel Interno
            </span>
            <h1 className="text-white text-4xl lg:text-5xl font-bold tracking-wide">
              Mapa de{" "}
              <span className="inline px-1 bg-[linear-gradient(transparent_70%,#F3C54F_70%)] leading-[1.4]">
                Denúncias
              </span>
            </h1>
          </div>

          <span className="text-white/60 text-lg">
            {complaints.length} denúncia{complaints.length === 1 ? "" : "s"}{" "}
            carregada{complaints.length === 1 ? "" : "s"}
          </span>
        </div>

        {erro && (
          <div className="mb-6 px-4 py-3 rounded-xl bg-red-500/20 border border-red-400/40 text-white">
            {erro}
          </div>
        )}

        <div className="w-full h-[60vh] min-h-[360px] bg-gray-200 rounded-[24px] overflow-hidden relative mb-6 shadow-[0_15px_35px_rgba(0,0,0,0.3)]">
          <ComplaintsMap
            complaints={complaints}
            selectedId={selectedId}
            onSelect={(complaint) => setSelectedId(complaint.id)}
          />
        </div>

        {selected ? (
          <ComplaintDetails
            complaint={selected}
            onClose={() => setSelectedId(null)}
            onDelete={handleDelete}
            onChangeState={handleChangeState}
            queue={queue}
            onSelectQueued={(c) => setSelectedId(c.id)}
            onReorderQueue={handleReorderQueue}
          />
        ) : (
          <div className="w-full bg-[#1B2237]/40 backdrop-blur-lg border border-white/10 rounded-[24px] p-8 text-center text-white/70">
            {!isLoading && complaints.length === 0
              ? "Nenhuma denúncia cadastrada até o momento."
              : "Clique em um marcador do mapa para inspecionar a denúncia."}
            {queue.length > 0 && (
              <div className="mt-4 flex justify-center">
                <AnalyzedQueue
                  complaints={queue}
                  selectedId={selectedId}
                  onSelect={(c) => setSelectedId(c.id)}
                  onReorder={handleReorderQueue}
                />
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
