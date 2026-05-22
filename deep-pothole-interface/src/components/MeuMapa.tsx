import { useEffect, useRef, useMemo } from "react";
import {MapContainer, TileLayer, Marker, useMap, useMapEvents} from 'react-leaflet'
import 'leaflet/dist/leaflet.css'

interface cordProps{
    lon: number
    lat: number
    onLocationSelect?: (lat: number, lon: number) => void
}

const AtualizadorDeMapa = ({ center }: { center: {lat: number, lon: number} }) => {
    const map = useMap()

    useEffect(() => {
        if(center && center.lat !== -1 && center.lon !== -1){
            map.setView([center.lat, center.lon], map.getZoom())
        }
    }, [center.lat, center.lon, map])

    return null
}

const MapEvents = ({ onLocationSelect }: { onLocationSelect?: (lat: number, lon: number) => void }) => {
    useMapEvents({
        click(e) {
            if(onLocationSelect) {
                onLocationSelect(e.latlng.lat, e.latlng.lng);
            }
        }
    })
    return null;
}

const MeuMapa = (coords: cordProps) => {
    const position: [number, number] = [coords.lat, coords.lon]
    
    const markerRef = useRef<any>(null)
    const eventHandlers = useMemo(
        () => ({
            dragend() {
                const marker = markerRef.current
                if (marker != null && coords.onLocationSelect) {
                    const pos = marker.getLatLng()
                    coords.onLocationSelect(pos.lat, pos.lng)
                }
            },
        }),
        [coords.onLocationSelect]
    )

    return(
        <div style={{ height: "100%", width: "100%", position: "relative", zIndex: 0 }}>
            <MapContainer
                center={position}
                zoom = {15}
                style={{ height: "100%", width: "100%"}}
            >
                <TileLayer
                    attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                    url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                />
                <Marker 
                    position={position} 
                    draggable={true}
                    eventHandlers={eventHandlers}
                    ref={markerRef}
                />
                <MapEvents onLocationSelect={coords.onLocationSelect} />
                
                <AtualizadorDeMapa center={{lat: coords.lat, lon: coords.lon}} />
            </MapContainer>
        </div>
    )
}

export default MeuMapa