import { Link } from "react-router";

function Navbar() {
  return (
    // O container 'fixed' prende a navbar no topo da tela.
    // O pointer-events-none garante que você possa clicar na tela ao redor dela.
    <div className="fixed top-4 z-50 lg:top-6 left-0 w-full flex justify-center px-6 pointer-events-none">
      {/* A tag <nav> volta a ter pointer-events-auto para os cliques funcionarem nela */}
      <nav className="pointer-events-auto flex items-center justify-between w-full max-w-[480px] lg:max-w-6xl px-6 py-3 lg:px-8 lg:py-4 rounded-full bg-[#1B2237]/80 backdrop-blur-lg border border-white/10 shadow-[0_8px_32px_rgba(0,0,0,0.2)] transition-all duration-300">
        {/* LOGO */}
        <Link to="/">
          <div className="flex flex-col font-bold font-['Krona_One'] text-lg lg:text-xl leading-[1.15] tracking-wide cursor-pointer group">
            <span className="flex items-center gap-1.5 text-white">
              Deep
              {/* Pontinho amarelo moderno que reage ao hover */}
              <span className="w-1.5 h-1.5 lg:w-2 lg:h-2 rounded-full bg-[#F3C54F] transition-transform duration-300 group-hover:scale-150"></span>
            </span>
            <span className="text-white/70 group-hover:text-white transition-colors duration-300">
              Pothole
            </span>
          </div>
        </Link>

        {/* LINK */}
        <Link to="/complaint">
          <div className="relative group text-white/90 text-base lg:text-lg font-medium cursor-pointer hover:text-[#F3C54F] transition-colors duration-300 px-2 py-1">
            Denúncia
            {/* Linha amarela animada que aparece no hover */}
            <span className="absolute -bottom-1 left-0 w-0 h-[2px] bg-[#F3C54F] transition-all duration-300 group-hover:w-full rounded-full"></span>
          </div>
        </Link>
      </nav>
    </div>
  );
}

export default Navbar;
