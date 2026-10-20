function Navbar() {
  return (
    // O container 'fixed' prende a navbar no topo da tela.
    // O pointer-events-none garante que você possa clicar na tela ao redor dela.
    <div className="fixed top-4 z-[1000] lg:top-6 left-0 w-full flex justify-center px-6 pointer-events-none">
      <nav className="pointer-events-auto flex items-center justify-between w-full max-w-[480px] lg:max-w-6xl px-6 py-3 lg:px-8 lg:py-4 rounded-full bg-[#1B2237]/80 backdrop-blur-lg border border-white/10 shadow-[0_8px_32px_rgba(0,0,0,0.2)] transition-all duration-300">
        {/* LOGO */}
        <div className="flex flex-col font-bold font-['Krona_One'] text-lg lg:text-xl leading-[1.15] tracking-wide group">
          <span className="flex items-center gap-1.5 text-white">
            Deep
            {/* Pontinho amarelo moderno que reage ao hover */}
            <span className="w-1.5 h-1.5 lg:w-2 lg:h-2 rounded-full bg-[#F3C54F] transition-transform duration-300 group-hover:scale-150"></span>
          </span>
          <span className="text-white/70 group-hover:text-white transition-colors duration-300">
            Pothole
          </span>
        </div>

        {/* IDENTIFICAÇÃO DO PAINEL */}
        <span className="text-white/90 text-base lg:text-lg font-medium px-2 py-1">
          Back Office
        </span>
      </nav>
    </div>
  );
}

export default Navbar;
