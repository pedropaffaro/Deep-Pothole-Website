export interface Complaint {
  id: number;
  city: string;
  street: string;
  latitude: number;
  longitude: number;
  photo_key: string;
  photo_url: string;
  detect_key: string;
  detect_url: string;
  pothole_probability: number | null;
  pothole_count: number | null;
  state: string;
  created_at: string | null;
}

const API_BASE_URL =
  import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1";

export async function fetchComplaints(): Promise<Complaint[]> {
  const response = await fetch(`${API_BASE_URL}/complaints`);

  if (!response.ok) {
    throw new Error(`Erro ${response.status} ao buscar denúncias`);
  }

  const data = await response.json();
  return Array.isArray(data) ? data : [];
}

export async function deleteComplaint(id: number): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/complaints/${id}`, {
    method: "DELETE",
  });

  if (!response.ok) {
    throw new Error(`Erro ${response.status} ao apagar denúncia ${id}`);
  }
}
