import Button from "../components/Button";
import Navbar from "../components/Navbar";
import GlassCard from "../components/GlassCard";
import StatisticCard from "../components/StatisticCard";

function Home() {
  return (
    <div className="min-h-screen font-[PT_Sans]">
      <section className="w-full flex bg-blue-secondary flex-col items-center pt-8 pb-16 px-6 relative overflow-hidden">
        <div className="w-full z-50 mb-30">
          <Navbar />
        </div>

        <div className="w-full max-w-[480px] lg:max-w-6xl flex flex-col lg:flex-row lg:items-center gap-12 z-10">
          <div className="flex-1 flex flex-col">
            <span className="text-lg font-bold tracking-[0.2em] text-white/80 uppercase mb-4">
              Sobre Nós
            </span>
            <h1 className="text-6xl lg:text-7xl font-bold text-white text-center lg:text-left mb-10 tracking-wide">
              Nosso Projeto
            </h1>

            <div className="mb-8">
              <span className="relative inline-block z-10 text-4xl font-bold text-white px-1 after:content-[''] after:absolute after:left-0 after:bottom-[2px] after:w-full after:h-[30%] after:bg-[#F3C54F] after:-z-10">
                Objetivo
              </span>
              <p className="text-white/80 text-[17px] mt-4 leading-relaxed">
                Identificar buracos nas vias a partir de imagens enviadas pelos
                próprios usuários.
              </p>
            </div>

            <div className="mb-10 lg:mb-14">
              <span className="inline px-1 text-4xl font-bold text-white bg-[linear-gradient(transparent_70%,#F3C54F_70%)] leading-[1.4]">
                IA e <br className="lg:hidden" />
                DeepLearning
              </span>
              <p className="text-white/80 text-[17px] mt-4 leading-relaxed">
                Usamos inteligência artificial e técnicas de aprendizado
                profundo para analisar imagens e geolocalizar buracos em vias.
              </p>
            </div>

            <Button
              label={"Faça sua denúncia"}
              to="/complaint"
              className="hidden lg:block w-fit px-12"
            />
          </div>

          <div className="flex-1 flex flex-col">
            <div className="grid grid-cols-2 gap-4 lg:gap-6 mb-10 lg:mb-0">
              <GlassCard>
                <img
                  src="/seg1.png"
                  alt="Detecção de buraco 1"
                  className="w-full h-full object-cover rounded-[5px] aspect-[3/4] min-h-[250px] lg:min-h-[350px]"
                />
              </GlassCard>

              <GlassCard>
                <img
                  src="/seg2.png"
                  alt="Detecção de buraco 2"
                  className="w-full h-full object-cover rounded-[5px] aspect-[3/4] min-h-[250px] lg:min-h-[350px]"
                />
              </GlassCard>
            </div>

            <Button
              label={"Faça sua denúncia"}
              to="/complaint"
              className="lg:hidden"
            />
          </div>
        </div>
      </section>

      <section className="w-full bg-white relative pt-20 pb-24 px-6 overflow-hidden">
        {/* === EFEITOS DE BLUR CLAROS NO FUNDO === */}
        <div className="absolute top-[-20%] left-[-10%] w-[60%] h-[60%] rounded-full bg-[#2A48A3]/5 blur-[120px] pointer-events-none z-0" />
        <div className="absolute top-[30%] right-[-20%] w-[50%] h-[50%] rounded-full bg-[#2A48A3]/10 blur-[130px] pointer-events-none z-0" />
        <div className="absolute bottom-[-10%] left-[20%] w-[40%] h-[40%] rounded-full bg-[#F3C54F]/5 blur-[100px] pointer-events-none z-0" />

        <div className="w-full max-w-[480px] lg:max-w-6xl mx-auto flex flex-col items-center relative z-10">
          <div className="text-center mb-14 flex flex-col items-center">
            <h2 className="text-6xl lg:text-7xl font-bold text-[#1B2237] mb-5 tracking-wide">
              Nosso {/* Mantendo o efeito de marca-texto que você gostou! */}
              <span className="inline px-1 bg-[linear-gradient(transparent_70%,#F3C54F_70%)] leading-[1.4]">
                Impacto
              </span>
            </h2>
            <p className="text-[#1B2237]/60 text-lg max-w-2xl text-center leading-relaxed">
              Acompanhe como nossa comunidade está ajudando a mapear e
              transformar a infraestrutura urbana, uma denúncia por vez.
            </p>
          </div>

          <div className="w-full flex flex-col lg:flex-row gap-6 lg:gap-8">
            {/* LADO ESQUERDO: MAPA COM BADGE FLUTUANTE */}
            <div className="flex-[3] relative w-full bg-white/80 backdrop-blur-sm rounded-[32px] p-3 shadow-[0_15px_40px_rgba(42,72,163,0.06)] border border-[#2A48A3]/10">
              {/* Badge "Ao vivo" */}
              <div className="absolute top-8 left-8 z-10 bg-white/90 backdrop-blur-md px-4 py-2 rounded-full flex items-center gap-2.5 shadow-[0_4px_12px_rgba(0,0,0,0.1)] border border-gray-100">
                <span className="w-2.5 h-2.5 rounded-full bg-red-500 animate-pulse"></span>
                <span className="text-[#1B2237] text-sm font-bold tracking-wide">
                  Mapeamento em tempo real
                </span>
              </div>

              <img
                src="/map.png"
                alt="Mapa de impacto mostrando denúncias"
                className="w-full h-full object-cover rounded-[20px] min-h-[350px] lg:min-h-[500px]"
              />
            </div>

            {/* LADO DIREITO: GRID DE ESTATÍSTICAS (Agora com 4 itens) */}
            <div className="flex-[2] w-full grid grid-cols-2 gap-4 lg:gap-6">
              <StatisticCard
                statisticNumber="200+"
                statisticLabel="Denúncias realizadas"
              />
              <StatisticCard
                statisticNumber="50+"
                statisticLabel="Bairros mapeados"
              />
              <StatisticCard
                statisticNumber="98%"
                statisticLabel="Precisão da IA"
              />
              <StatisticCard
                statisticNumber="24h"
                statisticLabel="Tempo de processamento"
              />
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}

export default Home;
