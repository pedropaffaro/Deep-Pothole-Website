import { useEffect, useMemo } from "react";
import { MapContainer, TileLayer, Marker, useMap } from "react-leaflet";
import L from "leaflet";
import "leaflet/dist/leaflet.css";

import type { Complaint } from "../lib/api";
import { pinColor } from "../lib/states";

interface ComplaintsMapProps {
  complaints: Complaint[];
  selectedId: number | null;
  onSelect: (complaint: Complaint) => void;
}

// Centro padrão: São Carlos - SP, usado enquanto não há denúncias carregadas
const DEFAULT_CENTER: [number, number] = [-22.0087, -47.8909];
const DEFAULT_ZOOM = 13;

const pinIcon = (color: string, selected: boolean) =>
  L.divIcon({
    className: "",
    html: `
      <div style="
        width: 26px;
        height: 26px;
        border-radius: 9999px 9999px 4px 9999px;
        transform: rotate(45deg);
        background: ${color};
        border: 3px solid ${selected ? "#1B2237" : "#FFFFFF"};
        ${selected ? "outline: 3px solid #FFCA3A;" : ""}
        box-shadow: 0 5px 12px rgba(0,0,0,0.35);
      "></div>
    `,
    iconSize: [26, 26],
    iconAnchor: [13, 26],
  });

// Ajusta o enquadramento do mapa sempre que a lista de denúncias muda
const AjustadorDeLimites = ({ complaints }: { complaints: Complaint[] }) => {
  const map = useMap();

  useEffect(() => {
    if (complaints.length === 0) return;

    const bounds = L.latLngBounds(
      complaints.map((c) => [c.latitude, c.longitude] as [number, number])
    );
    map.fitBounds(bounds, { padding: [48, 48], maxZoom: 16 });
  }, [complaints, map]);

  return null;
};

// Centraliza no item selecionado sem alterar o zoom escolhido pelo usuário
const FocoNaSelecao = ({ selected }: { selected: Complaint | null }) => {
  const map = useMap();

  useEffect(() => {
    if (!selected) return;
    map.panTo([selected.latitude, selected.longitude]);
  }, [selected, map]);

  return null;
};

const ComplaintsMap = ({
  complaints,
  selectedId,
  onSelect,
}: ComplaintsMapProps) => {
  const selected = useMemo(
    () => complaints.find((c) => c.id === selectedId) ?? null,
    [complaints, selectedId]
  );

  return (
    <div style={{ height: "100%", width: "100%", position: "relative", zIndex: 0 }}>
      <MapContainer
        center={DEFAULT_CENTER}
        zoom={DEFAULT_ZOOM}
        style={{ height: "100%", width: "100%" }}
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        {complaints.map((complaint) => (
          <Marker
            key={complaint.id}
            position={[complaint.latitude, complaint.longitude]}
            icon={pinIcon(pinColor(complaint.state), complaint.id === selectedId)}
            zIndexOffset={complaint.id === selectedId ? 1000 : 0}
            eventHandlers={{
              click: () => onSelect(complaint),
            }}
          />
        ))}

        <AjustadorDeLimites complaints={complaints} />
        <FocoNaSelecao selected={selected} />
      </MapContainer>
    </div>
  );
};

export default ComplaintsMap;
