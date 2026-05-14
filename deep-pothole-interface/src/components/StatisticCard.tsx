interface StatisticCardProps {
  statisticNumber: string;
  statisticLabel: string;
}

function StatisticCard({
  statisticNumber,
  statisticLabel,
}: StatisticCardProps) {
  return (
    <div className="bg-yellow-primary shadow-[0_5px_8px_rgba(0,0,0,0.25)] rounded-[24px] p-6 lg:p-10 flex flex-col items-center justify-center text-center hover:-translate-y-1 transition-transform duration-300">
      <span className="text-7xl lg:text-10xl font-black text-black mb-2 tracking-tight">
        {statisticNumber}
      </span>
      <span className="text-xl lg:text-xl font-light text-black/80">
        {statisticLabel}
      </span>
    </div>
  );
}

export default StatisticCard;
