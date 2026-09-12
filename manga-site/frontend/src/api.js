const BASE = import.meta.env.VITE_API_URL || '';

async function req(path, options = {}) {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  });
  if (res.status === 204) return null;
  const data = await res.json().catch(() => null);
  if (!res.ok) throw new Error((data && data.error) || `HTTP ${res.status}`);
  return data;
}

export const STATUSES = [
  { value: 'reading', label: 'Lendo' },
  { value: 'plan_to_read', label: 'Quero ler' },
  { value: 'finished', label: 'Concluído' },
  { value: 'dropped', label: 'Abandonado' },
];

export function pickText(localized, fallback = 'Sem título') {
  if (!localized) return fallback;
  if (typeof localized === 'string') return localized;
  return localized.en || localized['pt-br'] || Object.values(localized)[0] || fallback;
}

export function coverUrl(manga, size = 256) {
  const rel = (manga.relationships || []).find((r) => r.type === 'cover_art');
  const file = rel && rel.attributes && rel.attributes.fileName;
  if (!file) return '';
  return `https://uploads.mangadex.org/covers/${manga.id}/${file}.${size}.jpg`;
}

export function authors(manga) {
  return (manga.relationships || [])
    .filter((r) => r.type === 'author' && r.attributes && r.attributes.name)
    .map((r) => r.attributes.name)
    .slice(0, 2)
    .join(', ');
}

export const api = {
  search: (title, offset = 0) =>
    req(`/api/mangadex/search?title=${encodeURIComponent(title)}&limit=12&offset=${offset}`),
  detail: (id) => req(`/api/mangadex/manga/${id}`),
  chapters: (id, offset = 0, lang = 'en') =>
    req(`/api/mangadex/manga/${id}/chapters?limit=20&offset=${offset}&lang=${lang}`),
  library: (status = '') => req(`/api/library${status ? `?status=${status}` : ''}`),
  add: (entry) => req('/api/library', { method: 'POST', body: JSON.stringify(entry) }),
  update: (id, entry) => req(`/api/library/${id}`, { method: 'PUT', body: JSON.stringify(entry) }),
  remove: (id) => req(`/api/library/${id}`, { method: 'DELETE' }),
};
