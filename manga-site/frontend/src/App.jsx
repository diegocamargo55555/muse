import { useCallback, useEffect, useState } from 'react';
import { STATUSES, api, authors, coverUrl, pickText } from './api.js';

const statusLabel = (v) => (STATUSES.find((s) => s.value === v) || {}).label || v;

function SearchView({ onOpen, savedIds }) {
  const [q, setQ] = useState('berserk');
  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [status, setStatus] = useState('idle');
  const [error, setError] = useState('');

  const run = useCallback(async (query, off) => {
    if (!query.trim()) return;
    setStatus('loading');
    setError('');
    try {
      const data = await api.search(query.trim(), off);
      setItems(data.data || []);
      setTotal(data.total || 0);
      setOffset(off);
      setStatus('ready');
    } catch (e) {
      setError(e.message);
      setStatus('error');
    }
  }, []);

  useEffect(() => { run('berserk', 0); }, [run]);

  return (
    <>
      <form className="search-row" onSubmit={(e) => { e.preventDefault(); run(q, 0); }}>
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Buscar título — ex. berserk, one piece…" />
        <button className="btn primary" type="submit" disabled={status === 'loading'}>
          {status === 'loading' ? 'Buscando…' : 'Buscar'}
        </button>
      </form>
      {error && <p className="error">{error}</p>}
      {status === 'ready' && <p className="hint">{total} resultados · MangaDex (EN).</p>}
      <div className="grid">
        {items.map((m) => (
          <article className="manga-card" key={m.id}>
            {coverUrl(m) && <img src={coverUrl(m)} alt="" loading="lazy" />}
            <div className="body">
              <h3>{pickText(m.attributes.title)}</h3>
              <div className="by">{authors(m) || 'Autor desconhecido'}</div>
              <div className="row">
                <button className="btn" onClick={() => onOpen(m.id)}>Detalhe</button>
                {savedIds.has(m.id) && <span className="badge saved">na estante</span>}
              </div>
            </div>
          </article>
        ))}
      </div>
      {status === 'ready' && total > 12 && (
        <div className="pager">
          <button className="btn" disabled={offset === 0} onClick={() => run(q, Math.max(0, offset - 12))}>← Anterior</button>
          <button className="btn" disabled={offset + 12 >= total} onClick={() => run(q, offset + 12)}>Próxima →</button>
        </div>
      )}
    </>
  );
}

function DetailView({ id, onBack, onChanged }) {
  const [manga, setManga] = useState(null);
  const [chapters, setChapters] = useState([]);
  const [chTotal, setChTotal] = useState(0);
  const [chOffset, setChOffset] = useState(0);
  const [lang, setLang] = useState('en');
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(null);
  const [pick, setPick] = useState('reading');

  useEffect(() => {
    let alive = true;
    setManga(null); setChapters([]); setChOffset(0); setError(''); setSaved(null);
    api.detail(id).then((d) => { if (alive) setManga(d.data); }).catch((e) => alive && setError(e.message));
    api.library().then((lib) => {
      if (!alive) return;
      const found = lib.find((e) => e.mangaId === id);
      if (found) { setSaved(found); setPick(found.status); }
    }).catch(() => {});
    return () => { alive = false; };
  }, [id]);

  const loadChapters = useCallback(async (off, lg) => {
    try {
      const d = await api.chapters(id, off, lg);
      setChapters(off === 0 ? (d.data || []) : [...chapters, ...(d.data || [])]);
      setChTotal(d.total || 0);
      setChOffset(off);
    } catch (e) {
      setError(e.message);
    }
  }, [id, chapters]);

  useEffect(() => { loadChapters(0, lang); }, [id, lang]); // eslint-disable-line

  const save = async () => {
    setError('');
    try {
      const entry = {
        mangaId: id,
        title: pickText(manga.attributes.title),
        coverUrl: coverUrl(manga, 256),
        status: pick,
      };
      const res = saved ? await api.update(id, entry) : await api.add(entry);
      setSaved(res);
      onChanged();
    } catch (e) {
      setError(e.message);
    }
  };

  const remove = async () => {
    if (!confirm('Remover da estante?')) return;
    try {
      await api.remove(id);
      setSaved(null);
      onChanged();
    } catch (e) {
      setError(e.message);
    }
  };

  if (!manga) return (<><button className="btn" onClick={onBack}>← Voltar</button><p className="hint">Carregando…</p></>);
  const desc = pickText(manga.attributes.description, '');

  return (
    <>
      <button className="btn" onClick={onBack}>← Voltar</button>
      {error && <p className="error">{error}</p>}
      <div style={{ height: 12 }} />
      <div className="detail">
        <div>{coverUrl(manga, 512) && <img className="cover" src={coverUrl(manga, 512)} alt="" />}</div>
        <div>
          <h2>{pickText(manga.attributes.title)}</h2>
          <div className="by">{authors(manga)} {manga.attributes.year ? `· ${manga.attributes.year}` : ''} · {manga.attributes.status || ''}</div>
          <div className="savebox">
            <select value={pick} onChange={(e) => setPick(e.target.value)}>
              {STATUSES.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
            </select>
            <button className="btn primary" onClick={save}>{saved ? 'Atualizar estante' : 'Salvar na estante'}</button>
            {saved && <button className="btn danger" onClick={remove}>Remover</button>}
            {saved && <span className="badge saved">{statusLabel(saved.status)}</span>}
            <a className="btn" href={`https://mangadex.org/title/${id}`} target="_blank" rel="noreferrer">MangaDex</a>
          </div>
          <p className="desc">{desc.slice(0, 900)}{desc.length > 900 ? '…' : ''}</p>
        </div>
      </div>

      <div className="section-title">Capítulos ({chTotal})</div>
      <div className="tabs">
        {['en', 'pt-br'].map((l) => (
          <button key={l} className={`btn${lang === l ? ' active' : ''}`} onClick={() => { setLang(l); setChapters([]); }}>{l === 'en' ? 'Inglês' : 'Português'}</button>
        ))}
      </div>
      {chapters.length === 0 && <p className="hint">Nenhum capítulo neste idioma.</p>}
      <div className="chapters">
        {chapters.map((c) => (
          <div className="chapter" key={c.id}>
            <span>Cap. {c.attributes.chapter || '?'} — {c.attributes.title || 'sem título'}</span>
            <span className="meta">{c.attributes.translatedLanguage} · {(c.attributes.publishAt || '').slice(0, 10)}</span>
          </div>
        ))}
      </div>
      {chOffset + 20 < chTotal && (
        <div className="pager">
          <button className="btn" onClick={() => loadChapters(chOffset + 20, lang)}>Carregar mais</button>
        </div>
      )}
    </>
  );
}

function LibraryView({ refreshKey, onOpen }) {
  const [filter, setFilter] = useState('');
  const [entries, setEntries] = useState([]);
  const [error, setError] = useState('');

  const load = useCallback(async (f) => {
    try {
      setEntries(await api.library(f));
    } catch (e) {
      setError(e.message);
    }
  }, []);

  useEffect(() => { load(filter); }, [filter, refreshKey, load]);

  const move = async (e, st) => {
    try {
      await api.update(e.mangaId, { status: st });
      load(filter);
    } catch (err) { setError(err.message); }
  };

  const del = async (e) => {
    if (!confirm(`Remover "${e.title}"?`)) return;
    try {
      await api.remove(e.mangaId);
      load(filter);
    } catch (err) { setError(err.message); }
  };

  return (
    <>
      <div className="tabs">
        <button className={`btn${filter === '' ? ' active' : ''}`} onClick={() => setFilter('')}>Tudo</button>
        {STATUSES.map((s) => (
          <button key={s.value} className={`btn${filter === s.value ? ' active' : ''}`} onClick={() => setFilter(s.value)}>{s.label}</button>
        ))}
      </div>
      {error && <p className="error">{error}</p>}
      {entries.length === 0
        ? <p className="hint">Nada aqui. Busque um título e salve na estante.</p>
        : <div className="grid">
          {entries.map((e) => (
            <article className="manga-card" key={e.mangaId}>
              {e.coverUrl && <img src={e.coverUrl} alt="" loading="lazy" />}
              <div className="body">
                <h3>{e.title}</h3>
                <div><span className="badge saved">{statusLabel(e.status)}</span></div>
                <div className="row">
                  <button className="btn" onClick={() => onOpen(e.mangaId)}>Detalhe</button>
                </div>
                <div className="row">
                  <select value={e.status} onChange={(ev) => move(e, ev.target.value)} style={{ background: '#1e2522', border: '1px solid #2a332f', borderRadius: 8, color: '#eae7de', padding: '6px 8px', fontSize: 13 }}>
                    {STATUSES.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
                  </select>
                  <button className="btn danger" onClick={() => del(e)}>X</button>
                </div>
              </div>
            </article>
          ))}
        </div>}
    </>
  );
}

export default function App() {
  const [view, setView] = useState('search');
  const [openId, setOpenId] = useState(null);
  const [savedIds, setSavedIds] = useState(new Set());
  const [refreshKey, setRefreshKey] = useState(0);

  const refreshSaved = useCallback(async () => {
    try {
      const lib = await api.library();
      setSavedIds(new Set(lib.map((e) => e.mangaId)));
      setRefreshKey((k) => k + 1);
    } catch { /* biblioteca offline: segue sem badges */ }
  }, []);

  useEffect(() => { refreshSaved(); }, [refreshSaved]);

  const open = (id) => { setOpenId(id); setView('detail'); };

  return (
    <div className="wrap">
      <header className="topbar">
        <div className="brand">mangateca<span>.</span></div>
        <nav className="nav">
          <button className={`btn${view !== 'library' ? ' active' : ''}`} onClick={() => setView('search')}>Buscar</button>
          <button className={`btn${view === 'library' ? ' active' : ''}`} onClick={() => setView('library')}>Estante</button>
        </nav>
      </header>

      {view === 'search' && !openId && <SearchView onOpen={open} savedIds={savedIds} />}
      {(view === 'detail' || (view === 'search' && openId)) && openId && (
        <DetailView id={openId} onBack={() => { setOpenId(null); setView('search'); }} onChanged={refreshSaved} />
      )}
      {view === 'library' && <LibraryView refreshKey={refreshKey} onOpen={open} />}

      <footer>
        <span>Mangateca — MangaDex + estante (Lendo · Quero ler · Concluído · Abandonado)</span>
        <span>Go + React + Docker</span>
      </footer>
    </div>
  );
}
