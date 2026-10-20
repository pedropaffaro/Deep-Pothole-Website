import { useState } from "react";

import type { Complaint } from "../lib/api";
import StateTag from "./StateTag";

interface ComplaintDetailsProps {
  complaint: Complaint;
  onClose: () => void;
  onDelete: (complaint: Complaint) => Promise<void>;
}

interface InfoFieldProps {
  label: string;
  value: string;
}

interface ComplaintImageProps {
  label: string;
  src: string;
  alt: string;
  emptyMessage: string;
}

function InfoField({ label, value }: InfoFieldProps) {
  return (
    <div className="flex flex-col">
      <span className="text-sm font-bold tracking-[0.2em] text-white/60 uppercase mb-1">
        {label}
      </span>
      <span className="text-white text-lg break-words">{value || "—"}</span>
    </div>
  );
}

// A versão processada só existe para denúncias enviadas depois do modelo entrar
// no fluxo, então o erro de carregamento vira uma mensagem no lugar da imagem.
function ComplaintImage({ label, src, alt, emptyMessage }: ComplaintImageProps) {
  const [falhou, setFalhou] = useState<boolean>(false);

  return (
    <div className="flex-1 flex flex-col">
      <span className="text-sm font-bold tracking-[0.2em] text-white/60 uppercase mb-2">
        {label}
      </span>
      <div className="bg-white rounded-xl overflow-hidden h-56 lg:h-72 flex items-center justify-center">
        {src && !falhou ? (
          <img
            src={src}
            alt={alt}
            onError={() => setFalhou(true)}
            className="w-full h-full object-cover"
          />
        ) : (
          <span className="text-gray-500 text-sm px-4 text-center">
            {emptyMessage}
          </span>
        )}
      </div>
    </div>
  );
}

function ComplaintDetails({
  complaint,
  onClose,
  onDelete,
}: ComplaintDetailsProps) {
  // Exclusão apaga registro e imagens e não tem volta, então pede confirmação
  const [confirmando, setConfirmando] = useState<boolean>(false);
  const [apagando, setApagando] = useState<boolean>(false);

  const handleDelete = async () => {
    setApagando(true);
    try {
      await onDelete(complaint);
    } finally {
      setApagando(false);
      setConfirmando(false);
    }
  };

  return (
    <section className="w-full bg-[#1B2237]/60 backdrop-blur-lg border border-white/10 rounded-[24px] shadow-[0_15px_35px_rgba(0,0,0,0.3)] p-6 lg:p-8">
      <div className="flex items-start justify-between gap-4 mb-6">
        <div className="flex flex-col">
          <span className="text-md font-bold tracking-[0.2em] text-white/60 uppercase">
            Denúncia
          </span>
          <div className="flex items-center gap-3">
            <h2 className="text-white text-3xl lg:text-4xl font-bold tracking-wide">
              #{complaint.id}
            </h2>
            <StateTag state={complaint.state} />
          </div>
        </div>

        <div className="shrink-0 flex items-center gap-2">
          {confirmando ? (
            <>
              <span className="hidden sm:inline text-white/70 text-sm mr-1">
                Apagar denúncia e imagens?
              </span>
              <button
                type="button"
                onClick={handleDelete}
                disabled={apagando}
                className="px-4 py-2 bg-red-500 text-white font-semibold rounded-lg hover:bg-red-600 transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {apagando ? "Apagando..." : "Confirmar"}
              </button>
              <button
                type="button"
                onClick={() => setConfirmando(false)}
                disabled={apagando}
                className="px-4 py-2 bg-[#2a3c6b] border border-white/20 text-white font-semibold rounded-lg hover:bg-white/10 transition-colors cursor-pointer disabled:opacity-50"
              >
                Cancelar
              </button>
            </>
          ) : (
            <>
              <button
                type="button"
                onClick={() => setConfirmando(true)}
                className="px-4 py-2 bg-red-500/90 text-white font-semibold rounded-lg hover:bg-red-600 transition-colors cursor-pointer"
              >
                Apagar
              </button>
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 bg-[#2a3c6b] border border-white/20 text-white font-semibold rounded-lg hover:bg-white/10 transition-colors cursor-pointer"
              >
                Fechar
              </button>
            </>
          )}
        </div>
      </div>

      <div className="flex flex-col lg:flex-row gap-6 lg:gap-8">
        <div className="lg:flex-[2] w-full flex flex-col sm:flex-row gap-4">
          <ComplaintImage
            key={complaint.photo_url}
            label="Original"
            src={complaint.photo_url}
            alt={`Foto da denúncia ${complaint.id}`}
            emptyMessage="Sem imagem"
          />
          <ComplaintImage
            key={complaint.detect_url}
            label="Detecção"
            src={complaint.detect_url}
            alt={`Detecção da denúncia ${complaint.id}`}
            emptyMessage="Sem imagem processada pelo modelo"
          />
        </div>

        <div className="lg:flex-[1] w-full flex flex-col gap-5">
          <InfoField label="Cidade" value={complaint.city} />
          <InfoField label="Rua" value={complaint.street} />
          <InfoField
            label="Probabilidade"
            value={
              complaint.pothole_probability === null
                ? ""
                : `${(complaint.pothole_probability * 100).toFixed(1)}%`
            }
          />

          <a
            href={`https://www.google.com/maps/search/?api=1&query=${complaint.latitude},${complaint.longitude}`}
            target="_blank"
            rel="noreferrer"
            className="w-full mt-1 py-2 text-center bg-[#2a3c6b] border border-white/20 text-white font-semibold rounded-lg hover:bg-white/10 transition-colors"
          >
            Abrir no Google Maps
          </a>
        </div>
      </div>
    </section>
  );
}

export default ComplaintDetails;
