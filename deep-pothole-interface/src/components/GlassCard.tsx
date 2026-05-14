import React from "react";

interface GlassCardProps {
  children: React.ReactNode;
  className?: string;
}

const GlassCard = ({ children, className = "" }: GlassCardProps) => {
  return (
    <div
      className={`
      relative overflow-hidden w-full h-full
      rounded-[10px]
      border-[10px] border-white/10 
      
      bg-gradient-to-br from-white/10 to-transparent
      backdrop-blur-xl
      
      shadow-[0_15px_35px_rgba(0,0,0,0.3),inset_0px_10px_2px_rgba(255,255,255,0.3)]
      
      ${className}
    `}
    >
      <div className="w-full h-full p-0 rounded-[5px] flex items-center justify-center">
        {children}
      </div>
    </div>
  );
};

export default GlassCard;
