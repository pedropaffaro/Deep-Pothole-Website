import {
  DndContext,
  MouseSensor,
  TouchSensor,
  closestCenter,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import type { DragEndEvent } from "@dnd-kit/core";
import {
  SortableContext,
  arrayMove,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";
import { CSS } from "@dnd-kit/utilities";

import type { Complaint } from "../lib/api";
import { pinColor } from "../lib/states";

interface AnalyzedQueueProps {
  // Já ordenada: o primeiro da lista é o que está há mais tempo na fila
  complaints: Complaint[];
  selectedId: number | null;
  onSelect: (complaint: Complaint) => void;
  // Recebe os ids na nova ordem depois de um arrasto
  onReorder: (ids: number[]) => void;
}

interface QueueItemProps {
  complaint: Complaint;
  ativo: boolean;
  compacta: boolean;
  onSelect: (complaint: Complaint) => void;
}

// Acima disso os retângulos perdem o texto e viram só faixas de cor
const MAX_COM_TEXTO = 4;

function QueueItem({ complaint, ativo, compacta, onSelect }: QueueItemProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: complaint.id });

  const nome = complaint.street || "sem rua";

  return (
    <li
      ref={setNodeRef}
      className="shrink-0"
      style={{
        transform: CSS.Transform.toString(transform),
        transition,
        zIndex: isDragging ? 10 : undefined,
        position: "relative",
      }}
    >
      <button
        type="button"
        {...attributes}
        {...listeners}
        onClick={() => onSelect(complaint)}
        title={`${nome}, ${complaint.city}`}
        aria-label={`Selecionar denúncia em ${nome}`}
        className={`w-full rounded-md border text-left transition-colors ${
          isDragging ? "cursor-grabbing shadow-lg" : "cursor-grab"
        } ${compacta ? "h-4" : "px-2 py-1 text-xs truncate"} ${
          ativo
            ? "bg-[#FFCA3A] border-[#1B2237] text-[#1B2237]"
            : "border-white/20 text-white hover:brightness-125"
        }`}
        style={ativo ? undefined : { background: `${pinColor(complaint.state)}99` }}
      >
        {compacta ? null : nome}
      </button>
    </li>
  );
}

function AnalyzedQueue({
  complaints,
  selectedId,
  onSelect,
  onReorder,
}: AnalyzedQueueProps) {
  // Distância mínima para o clique continuar selecionando; no toque, o arrasto
  // só começa após segurar, para não brigar com a rolagem da lista
  const sensors = useSensors(
    useSensor(MouseSensor, { activationConstraint: { distance: 5 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 200, tolerance: 6 } })
  );

  if (complaints.length === 0) return null;

  const compacta = complaints.length > MAX_COM_TEXTO;
  const ids = complaints.map((c) => c.id);

  const handleDragEnd = ({ active, over }: DragEndEvent) => {
    if (!over || active.id === over.id) return;
    const from = ids.indexOf(Number(active.id));
    const to = ids.indexOf(Number(over.id));
    if (from < 0 || to < 0) return;
    onReorder(arrayMove(ids, from, to));
  };

  return (
    <div className="shrink-0 w-28 flex flex-col">
      <span className="mb-1 text-xs font-bold leading-tight tracking-[0.1em] text-white/60 uppercase">
        Prioridade de atendimento
      </span>

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        modifiers={[restrictToVerticalAxis]}
        onDragEnd={handleDragEnd}
      >
        <SortableContext items={ids} strategy={verticalListSortingStrategy}>
          <ul className="flex flex-col gap-1 overflow-y-auto max-h-32 pr-1">
            {complaints.map((complaint) => (
              <QueueItem
                key={complaint.id}
                complaint={complaint}
                ativo={complaint.id === selectedId}
                compacta={compacta}
                onSelect={onSelect}
              />
            ))}
          </ul>
        </SortableContext>
      </DndContext>
    </div>
  );
}

export default AnalyzedQueue;
