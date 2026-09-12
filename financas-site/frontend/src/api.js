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

export const TYPES = [
  { value: 'deposit', label: 'Aporte' },
  { value: 'withdraw', label: 'Retirada' },
  { value: 'buy', label: 'Compra' },
  { value: 'sell', label: 'Venda' },
  { value: 'dividend', label: 'Dividendo' },
];

const typeLabel = (v) => (TYPES.find((t) => t.value === v) || {}).label || v;

export function money(v, cur = 'BRL') {
  try {
    return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: cur }).format(v || 0);
  } catch {
    return `${(v || 0).toFixed(2)} ${cur}`;
  }
}

export function today() {
  return new Date().toISOString().slice(0, 10);
}

export const api = {
  accounts: () => req('/api/accounts'),
  createAccount: (a) => req('/api/accounts', { method: 'POST', body: JSON.stringify(a) }),
  deleteAccount: (id) => req(`/api/accounts/${id}`, { method: 'DELETE' }),
  transactions: (accountId = 0) => req(`/api/transactions${accountId ? `?accountId=${accountId}` : ''}`),
  createTxn: (t) => req('/api/transactions', { method: 'POST', body: JSON.stringify(t) }),
  deleteTxn: (id) => req(`/api/transactions/${id}`, { method: 'DELETE' }),
  quotes: (br, us) => req(`/api/quotes?br=${encodeURIComponent(br)}&us=${encodeURIComponent(us)}`),
  portfolio: (accountId = 0) => req(`/api/portfolio${accountId ? `?accountId=${accountId}` : ''}`),
};

export { typeLabel };
