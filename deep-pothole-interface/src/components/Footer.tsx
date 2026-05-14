function Footer() {
  return (
    <footer
      className="w-full text-white pt-12 pb-6 px-6 lg:px-12"
      style={{
        background:
          "conic-gradient(from -210deg at 0% 0%, #1B2237 0%, #202D53 40%, #2A48A3 72%, #1B2237 96%)",
      }}
    >
      <div className="max-w-[480px] lg:max-w-6xl mx-auto flex flex-col">
        <div className="flex flex-col lg:flex-row justify-between gap-10 mb-8">
          <div className="flex flex-col">
            <h2 className="text-3xl font-['Krona_One'] mb-4 tracking-wide">
              Deep Pothole
            </h2>
            <p className="text-gray-300 text-sm max-w-[300px] leading-relaxed">
              Lorem ipsum dolor sit amet, consectetur adipiscing elit.
            </p>
          </div>

          <div className="flex flex-col">
            <h3 className="font-bold text-lg mb-1">Contato</h3>
            <p className="text-gray-300 text-sm mb-4">
              Email: contato@gruporaia.com
            </p>

            {/* Ícones das Redes Sociais */}
            <div className="flex items-center gap-4">
              {/* Substitua os src pelos seus SVGs locais */}
              <a href="#" className="hover:opacity-70 transition-opacity">
                <img
                  src="/icon-instagram.png"
                  alt="Instagram"
                  className="w-6 h-6"
                />
              </a>
              <a href="#" className="hover:opacity-70 transition-opacity">
                <img
                  src="/icon-linkedin.png"
                  alt="LinkedIn"
                  className="w-6 h-6"
                />
              </a>
              <a href="#" className="hover:opacity-70 transition-opacity">
                <img
                  src="/icon-twitter.png"
                  alt="X (Twitter)"
                  className="w-6 h-6"
                />
              </a>
              <a href="#" className="hover:opacity-70 transition-opacity">
                <img src="/icon-github.png" alt="GitHub" className="w-6 h-6" />
              </a>
            </div>
          </div>
        </div>

        {/* === LINHA DIVISÓRIA SUTIL === */}
        <hr className="border-t border-white/10 w-full mb-6" />

        {/* === PARTE INFERIOR: Direitos e Créditos === */}
        <div className="flex flex-col lg:flex-row justify-between items-center gap-4 text-lg font-bold text-gray-200 tracking-wide">
          <p>© 2026 Grupo RAIA</p>
          <p>Feito por USPCodeLab Sanca</p>
        </div>
      </div>
    </footer>
  );
}

export default Footer;
