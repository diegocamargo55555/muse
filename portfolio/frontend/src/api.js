const BASE = import.meta.env.VITE_API_URL || '';

async function req(path, options = {}) {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  });
  if (res.status === 204) return null;
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error((data && data.error) || `HTTP ${res.status}`);
  }
  return data;
}

export const api = {
  list: () => req('/api/projects'),
  get: (id) => req(`/api/projects/${id}`),
  create: (project, token) =>
    req('/api/projects', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(project),
    }),
  update: (id, project, token) =>
    req(`/api/projects/${id}`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(project),
    }),
  remove: (id, token) =>
    req(`/api/projects/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` },
    }),
};

export const samples = [
  {
    id: 'site-investimentos',
    title: 'Site de Investimentos',
    description: 'Acompanhamento de carteiras, aportes e rentabilidade.',
    tech: ['Go', 'React'],
    featured: true,
  },
  {
    id: 'site-manga',
    title: 'Site de Mangá',
    description: 'Catálogo e leitura com favoritos e progresso.',
    tech: ['React', 'Go'],
    featured: true,
  },
];
