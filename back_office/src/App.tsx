import { useEffect, useState } from "react";

import Navbar from "./components/Navbar";
import Spinner from "./components/Spinner";
import ComplaintsMap from "./components/ComplaintsMap";
import ComplaintDetails from "./components/ComplaintDetails";
import { deleteComplaint, fetchComplaints } from "./lib/api";
import type { Complaint } from "./lib/api";

function App() {
  const [complaints, setComplaints] = useState<Complaint[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [erro, setErro] = useState<string | null>(null);

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

  const handleDelete = async (complaint: Complaint) => {
    try {
      await deleteComplaint(complaint.id);
      setComplaints((atuais) => atuais.filter((c) => c.id !== complaint.id));
      setSelectedId(null);
      setErro(null);
    } catch (error) {
      console.error("Erro ao apagar denúncia:", error);
      setErro(`Não foi possível apagar a denúncia #${complaint.id}.`);
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
          />
        ) : (
          <div className="w-full bg-[#1B2237]/40 backdrop-blur-lg border border-white/10 rounded-[24px] p-8 text-center text-white/70">
            {!isLoading && complaints.length === 0
              ? "Nenhuma denúncia cadastrada até o momento."
              : "Clique em um marcador do mapa para inspecionar a denúncia."}
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
