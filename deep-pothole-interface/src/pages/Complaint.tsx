import Navbar from "../components/Navbar";
import Button from "../components/Button";
import Spinner from "../components/Spinner";
import MeuMapa from "../components/MeuMapa";

import { useState, useEffect, useRef } from "react";

// Interface para mapear a resposta da API de geolocalização (Nominatim)
interface NominatimResponse {
  address?: {
    city?: string;
    town?: string;
    village?: string;
    road?: string;
  };
}

function Complaint() {
  const [foto, setFoto] = useState<string | null>(null);
  const [cidade, setCidade] = useState<string>("");
  const [rua, setRua] = useState<string>("");
  const [coords, setCoords] = useState<{ lat: number; lng: number }>(
    {lat: -1, lng: -1}
  );
  const [isLoadingLocation, setIsLoadingLocation] = useState<boolean>(false);
  const [isSending, setSending] = useState<boolean>(false)

  const fileInputRef = useRef<HTMLInputElement>(null);

  const resetForm = () => {
    setFoto(null);

    setCidade("");
    setRua("");

    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const fetchAddressByCoords = async (lat: number, lng: number) => {
    try {
      const response = await fetch(
        `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}`
      );
      const data: NominatimResponse = await response.json();
      if (data && data.address) {
        setCidade(
          data.address.city ||
            data.address.town ||
            data.address.village ||
            ""
        );
        setRua(data.address.road || "");
      }
    } catch (error) {
      console.error("Erro ao buscar o endereço:", error);
    }
  };

  const handleLocationSelect = (lat: number, lng: number) => {
    setCoords({ lat, lng });
    fetchAddressByCoords(lat, lng);
  };

  const handleSearchAddress = async () => {
    if (!cidade && !rua) return;
    setIsLoadingLocation(true);
    try {
      const query = `${rua}, ${cidade}`;
      const response = await fetch(
        `https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(
          query
        )}`
      );
      const data = await response.json();
      if (data && data.length > 0) {
        setCoords({ lat: parseFloat(data[0].lat), lng: parseFloat(data[0].lon) });
      } else {
        alert("Endereço não encontrado no mapa.");
      }
    } catch (error) {
      console.error("Erro na busca de endereço:", error);
    } finally {
      setIsLoadingLocation(false);
    }
  };

  useEffect(() => {
    if ("geolocation" in navigator) {
      setIsLoadingLocation(true);
      navigator.geolocation.getCurrentPosition(
        async (position: GeolocationPosition) => {
          const { latitude, longitude } = position.coords;
          setCoords({ lat: latitude, lng: longitude });
          await fetchAddressByCoords(latitude, longitude);
          setIsLoadingLocation(false);
        },
        (error: GeolocationPositionError) => {
          console.error("Erro ao obter a localização do dispositivo:", error);
          setIsLoadingLocation(false);
        }
      );
    } else {
      console.log("Geolocalização não é suportada neste navegador.");
    }
  }, []);

  // Referenciando diretamente pelo namespace do React
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setFoto(URL.createObjectURL(e.target.files[0]));
    }
  };

  // Referenciando diretamente pelo namespace do React
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    // 1. Criamos o pacote de dados
    const formData = new FormData();
    formData.append("cidade", cidade);
    formData.append("rua", rua);
    if (coords) {
      formData.append("latitude", coords.lat.toString());
      formData.append("longitude", coords.lng.toString());
    }
    if (fileInputRef.current?.files?.[0]) {
      formData.append("foto", fileInputRef.current.files[0]);
    }

    try {
      // desativador de chamadas ate o retorno da chamada anterior da API
      setSending(true)

      // !!!!!!!!!!!!!: Remover após teste visual do spinner
      await new Promise((resolve) => setTimeout(resolve, 3000));

      // 2. Fazemos a chamada para o IP/Porta do Go
      const response = await fetch("http://localhost:8080/api/v1/complaint", {
        method: "POST",
        body: formData,
      });

      if (response.ok) {
        const data = await response.json();
        console.log("Sucesso:", data);
        alert("Denúncia salva no SQLite!");

        resetForm();
      } else {
        alert("Erro no servidor");
      }
    } catch (error) {
      console.error("Erro de conexão:", error);
      alert("O backend está rodando?");
    } finally {
      setSending(false)
    }
  };

  return (
    <div className="min-h-screen bg-linear-to-b from-[#2a3c6b] to-[#1e293b] font-sans pb-10">
      <Navbar />

      {/* Renderiza o Spinner quando o formulário está sendo enviado */}
      {isSending && <Spinner />}

      <main className="max-w-md mx-auto px-6 pt-32">
        <h1 className="text-white text-4xl font-bold mb-8 text-center">
          Faça Sua Denúncia
        </h1>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Sessão da Foto */}
          <div>
            <div className="flex justify-between items-center mb-1">
              <span className="text-md font-bold tracking-[0.2em] text-white/80 uppercase mb-1 mt-2">
                Imagem do Buraco
              </span>
            </div>
            <div
              onClick={() => fileInputRef.current?.click()}
              className="bg-white rounded-xl h-48 flex flex-col items-center justify-center cursor-pointer overflow-hidden relative"
            >
              {foto ? (
                <img
                  src={foto}
                  alt="Preview do buraco"
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="text-[#9CA3AF] flex flex-col items-center">
                  <svg
                    className="w-6 h-6 mb-2"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
                    ></path>
                  </svg>
                  <span className="text-sm">Envie sua imagem aqui</span>
                </div>
              )}
              <input
                type="file"
                accept="image/*"
                ref={fileInputRef}
                onChange={handleFileChange}
                className="hidden"
                required
              />
            </div>
          </div>

          {/* Sessão de Localização */}
          <div>
            <div className="flex justify-between items-center mb-1">
              <span className="text-md font-bold tracking-[0.2em] text-white/80 uppercase mb-1">
                Localização
              </span>
              {isLoadingLocation && (
                <span className="text-md font-bold tracking-[0.2em] text-yellow-primary uppercase mb-1">
                  Buscando...
                </span>
              )}
            </div>

            <div className="w-full h-64 bg-gray-200 rounded-xl overflow-hidden relative mb-2">
              {coords.lat !== -1 && coords.lng !== -1 ? (
                <MeuMapa lat={coords.lat} lon={coords.lng} onLocationSelect={handleLocationSelect} />
              ) : (
                <div className="w-full h-full flex items-center justify-center text-gray-500">
                  Aguardando localização...
                </div>
              )}
            </div>

            <div className="space-y-2 mt-3 rounded-xl">
              <div className="flex gap-2 w-full">
                <input
                  type="text"
                  placeholder="Cidade"
                  value={cidade}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    setCidade(e.target.value)
                  }
                  className="w-full px-4 py-3 rounded-lg bg-white text-blue-primary placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#FACC15]"
                  required
                />
              </div>
              <input
                type="text"
                placeholder="Rua"
                value={rua}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setRua(e.target.value)
                }
                className="w-full px-4 py-3 rounded-lg bg-white text-blue-primary placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#FACC15]"
                required
              />
              <button
                type="button"
                onClick={handleSearchAddress}
                className="w-full mt-2 py-2 bg-[#2a3c6b] border border-white/20 text-white font-semibold rounded-lg hover:bg-white/10 transition-colors"
              >
                Buscar endereço no mapa
              </button>
            </div>
          </div>

          <Button type="submit" label="Enviar Denúncia" className="mt-4" disabled={isSending}/>
        </form>
      </main>
    </div>
  );
}

export default Complaint;
