import { useCallback, useEffect, useState } from 'react';
import { TYPES, api, money, today, typeLabel } from './api.js';

function AccountsView({ accounts, onChange }) {
  const [name, setName] = useState('');
  const [currency, setCurrency] = useState('BRL');
  const [error, setError] = useState('');

  const create = async (e) => {
    e.preventDefault();
    setError('');
    try {
      await api.createAccount({ name, currency });
      setName('');
      onChange();
    } catch (err) { setError(err.message); }
  };

  const del = async (a) => {
    if (!confirm(`Excluir "${a.name}" e seus lançamentos?`)) return;
    try {
      await api.deleteAccount(a.id);
      onChange();
    } catch (err) { setError(err.message); }
  };

  return (
    <>
      <div className="section-title">Contas ({accounts.length})</div>
      {error && <p className="error">{error}</p>}
      <div className="cards">
        {accounts.map((a) => (
          <div className="card" key={a.id}>
            <h3>{a.name}</h3>
            <div className="sub mono">{a.currency}</div>
            <div className="row">
              <button className="btn danger" onClick={() => del(a)}>Excluir</button>
            </div>
          </div>
        ))}
      </div>
      <form className="form" onSubmit={create}>
        <div className="row2">
          <label>Nome <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Corretora BR / Broker US" /></label>
          <label>Moeda (3 letras) <input value={currency} onChange={(e) => setCurrency(e.target.value.toUpperCase())} placeholder="BRL" maxLength={3} /></label>
        </div>
        <div><button className="btn primary" type="submit">Criar conta</button></div>
      </form>
      <p className="hint">Uma conta por moeda/corretora. Lançamentos usam a moeda da conta para o caixa.</p>
    </>
  );
}

const emptyTxn = { accountId: '', type: 'deposit', symbol: '', market: 'BR', qty: '1', price: '', currency: 'BRL', date: today(), note: '' };

function TxnsView({ accounts, onChange }) {
  const [filter, setFilter] = useState(0);
  const [items, setItems] = useState([]);
  const [form, setForm] = useState(emptyTxn);
  const [error, setError] = useState('');

  const load = useCallback(async () => {
    try {
      setItems(await api.transactions(filter));
    } catch (e) { setError(e.message); }
  }, [filter]);

  useEffect(() => { load(); }, [load, accounts]);

  const set = (k, v) => setForm({ ...form, [k]: v });

  const submit = async (e) => {
    e.preventDefault();
    setError('');
    try {
      await api.createTxn({
        accountId: Number(form.accountId),
        type: form.type,
        symbol: form.symbol.toUpperCase().trim(),
        market: form.market,
        qty: Number(form.qty),
        price: Number(form.price),
        currency: form.currency.toUpperCase(),
        date: form.date,
        note: form.note,
      });
      setForm({ ...emptyTxn, accountId: form.accountId });
      load();
      onChange();
    } catch (err) { setError(err.message); }
  };

  const del = async (id) => {
    if (!confirm('Excluir lançamento?')) return;
    try {
      await api.deleteTxn(id);
      load();
      onChange();
    } catch (err) { setError(err.message); }
  };

  const needsSymbol = form.type === 'buy' || form.type === 'sell';

  return (
    <>
      <div className="toolbar">
        <select value={filter} onChange={(e) => setFilter(Number(e.target.value))}>
          <option value={0}>Todas as contas</option>
          {accounts.map((a) => <option key={a.id} value={a.id}>{a.name} ({a.currency})</option>)}
        </select>
      </div>
      {error && <p className="error">{error}</p>}
      <table className="table">
        <thead><tr><th>Data</th><th>Tipo</th><th>Ativo</th><th>Qtd × Preço</th><th className="num">Valor</th><th></th></tr></thead>
        <tbody>
          {items.map((t) => (
            <tr key={t.id}>
              <td className="mono">{(t.date || '').slice(0, 10)}</td>
              <td>{typeLabel(t.type)}</td>
              <td className="mono">{t.symbol ? `${t.symbol} · ${t.market}` : '—'}</td>
              <td className="mono">{t.qty} × {t.price}</td>
              <td className="num">{money(t.qty * t.price, t.currency)}</td>
              <td><button className="btn danger" onClick={() => del(t.id)}>X</button></td>
            </tr>
          ))}
        </tbody>
      </table>
      {items.length === 0 && <p className="hint">Nenhum lançamento. Registre um aporte abaixo.</p>}

      <div className="section-title">Novo lançamento</div>
      <form className="form" onSubmit={submit}>
        <div className="row3">
          <label>Conta
            <select value={form.accountId} onChange={(e) => {
              const id = e.target.value;
              const acc = accounts.find((a) => String(a.id) === String(id));
              setForm({ ...form, accountId: id, currency: acc ? acc.currency : form.currency });
            }}>
              <option value="">Selecionar…</option>
              {accounts.map((a) => <option key={a.id} value={a.id}>{a.name} ({a.currency})</option>)}
            </select>
          </label>
          <label>Tipo
            <select value={form.type} onChange={(e) => set('type', e.target.value)}>
              {TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
            </select>
          </label>
          <label>Data <input type="date" value={form.date} onChange={(e) => set('date', e.target.value)} /></label>
        </div>
        <div className="row3">
          <label>Ativo {needsSymbol && '(obrigatório)'} <input value={form.symbol} onChange={(e) => set('symbol', e.target.value.toUpperCase())} placeholder="PETR4 / AAPL" /></label>
          <label>Mercado
            <select value={form.market} onChange={(e) => set('market', e.target.value)}>
              <option value="BR">BR — brapi</option>
              <option value="US">Exterior — Yahoo</option>
            </select>
          </label>
          <label>Moeda <input value={form.currency} onChange={(e) => set('currency', e.target.value.toUpperCase())} maxLength={3} /></label>
        </div>
        <div className="row3">
          <label>Qtd <input type="number" step="any" value={form.qty} onChange={(e) => set('qty', e.target.value)} /></label>
          <label>Preço <input type="number" step="any" value={form.price} onChange={(e) => set('price', e.target.value)} /></label>
          <label>Nota <input value={form.note} onChange={(e) => set('note', e.target.value)} placeholder="opcional" /></label>
        </div>
        <div><button className="btn primary" type="submit">Salvar</button></div>
      </form>
    </>
  );
}

function PortfolioView({ accounts, refreshKey }) {
  const [filter, setFilter] = useState(0);
  const [data, setData] = useState(null);
  const [status, setStatus] = useState('idle');

  const load = useCallback(async () => {
    setStatus('loading');
    try {
      setData(await api.portfolio(filter));
      setStatus('ready');
    } catch {
      setStatus('error');
    }
  }, [filter]);

  useEffect(() => { load(); }, [load, refreshKey]);

  return (
    <>
      <div className="toolbar">
        <select value={filter} onChange={(e) => setFilter(Number(e.target.value))}>
          <option value={0}>Consolidado</option>
          {accounts.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}
        </select>
        <button className="btn" onClick={load}>Atualizar cotações</button>
      </div>
      {status === 'loading' && <p className="hint">Buscando cotações (brapi + Yahoo)…</p>}
      {status === 'error' && <p className="error">Falha ao carregar carteira.</p>}
      {status === 'ready' && data && (
        <>
          <div className="summary">
            <div className="stat"><div className="k">Total (BRL)</div><div className="v">{money(data.totalBRL, 'BRL')}</div></div>
            {Object.entries(data.cashByCurrency || {}).map(([c, v]) => (
              <div className="stat" key={`c${c}`}><div className="k">Caixa {c}</div><div className="v">{money(v, c)}</div></div>
            ))}
            {Object.entries(data.marketByCurrency || {}).map(([c, v]) => (
              <div className="stat" key={`m${c}`}><div className="k">Mercado {c}</div><div className="v">{money(v, c)}</div></div>
            ))}
          </div>
          <p className="hint">
            Câmbio USD→BRL: {data.fxLive ? `${data.fxUsed} (ao vivo)` : 'indisponível — USD contado sem conversão'}.
            {(data.missingQuotes || []).length > 0 && ` Sem cotação: ${data.missingQuotes.join(', ')}.`}
            {data.quoteErrors && Object.keys(data.quoteErrors).length > 0 && ' (Yahoo pode limitar — tente de novo em 1 min.)'}
          </p>
          <div className="section-title">Posições ({(data.positions || []).length})</div>
          <table className="table">
            <thead><tr><th>Ativo</th><th className="num">Qtd</th><th className="num">PM</th><th className="num">Preço</th><th className="num">Mercado</th><th className="num">P/L</th></tr></thead>
            <tbody>
              {(data.positions || []).map((p) => (
                <tr key={p.symbol}>
                  <td className="mono">{p.symbol} · {p.market}{!p.quoteOK && ' (s/ cotação)'}</td>
                  <td className="num">{p.qty}</td>
                  <td className="num">{money(p.avgCost, p.currency)}</td>
                  <td className="num">{p.quoteOK ? money(p.price, p.currency) : '—'}</td>
                  <td className="num">{money(p.marketValue, p.currency)}</td>
                  <td className={`num ${p.pnl >= 0 ? 'pos' : 'neg'}`}>{money(p.pnl, p.currency)}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {(data.positions || []).length === 0 && <p className="hint">Sem posições. Registre compras na aba Lançamentos.</p>}
        </>
      )}
    </>
  );
}

export default function App() {
  const [view, setView] = useState('portfolio');
  const [accounts, setAccounts] = useState([]);
  const [refreshKey, setRefreshKey] = useState(0);

  const reload = useCallback(async () => {
    try {
      setAccounts(await api.accounts());
      setRefreshKey((k) => k + 1);
    } catch { /* API offline */ }
  }, []);

  useEffect(() => { reload(); }, [reload]);

  return (
    <div className="wrap">
      <header className="topbar">
        <div className="brand">carteira<span>.</span></div>
        <nav className="nav">
          <button className={`btn${view === 'portfolio' ? ' active' : ''}`} onClick={() => setView('portfolio')}>Carteira</button>
          <button className={`btn${view === 'txns' ? ' active' : ''}`} onClick={() => setView('txns')}>Lançamentos</button>
          <button className={`btn${view === 'accounts' ? ' active' : ''}`} onClick={() => setView('accounts')}>Contas</button>
        </nav>
      </header>

      {view === 'portfolio' && <PortfolioView accounts={accounts} refreshKey={refreshKey} />}
      {view === 'txns' && <TxnsView accounts={accounts} onChange={reload} />}
      {view === 'accounts' && <AccountsView accounts={accounts} onChange={reload} />}

      <footer>
        <span>Carteira — brapi (BR) + Yahoo (exterior) · contas multi-moeda</span>
        <span>Gin + GORM + Postgres</span>
      </footer>
    </div>
  );
}
