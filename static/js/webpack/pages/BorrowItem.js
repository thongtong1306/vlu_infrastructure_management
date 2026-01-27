// pages/BorrowItem.js
import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Link } from 'react-router-dom';

const QR_BORROW_DIV_ID = 'imx-qr-borrow';
const QR_RETURN_DIV_ID = 'imx-qr-return';

export default function BorrowItem() {
    useEffect(() => { document.title = 'Mượn / Trả | Cơ sở vật chất/thiết bị'; }, []);

    // ---- auth ----
    const session = useMemo(() => {
        try { return JSON.parse(localStorage.getItem('imx_session') || 'null'); } catch { return null; }
        }, []);
    const userId =
        session?.user?.id ??
        session?.user_id ??
        null;

    // Auto-fill from login info (read-only)
    const borrowerName = useMemo(() => {
        const u = session?.user || {};
        return (
            u.full_name ||
            u.name ||
            u.displayName ||
            u.username ||
            u.email ||
            session?.email ||
            ''
        );
    }, [session]);

    // read token from localStorage OR cookie (fallback)
    const token =
        session?.token ||
        (document.cookie.split('; ').find(c => c.startsWith('imx_token='))?.split('=')[1] || null);

    // default headers; add Authorization if we have a token
    const headers = {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
    };

    // unified API helper: always send cookies + our headers
    const api = (url, opts = {}) => {
        const mergedHeaders = { ...headers, ...(opts.headers || {}) };
        return fetch(url, { credentials: 'include', ...opts, headers: mergedHeaders });
    };

    const needAuth = () => {
        if (!token) { setError('Please sign in first.'); return true; }
        return false;
    };

    // ---- tabs ----
    const [tab, setTab] = useState('borrow'); // 'borrow' | 'return'

    // ---- alerts ----
    const [error, setError] = useState('');
    const [ok, setOk] = useState('');
    const resetAlerts = () => { setError(''); setOk(''); };

    // ---- equipment list for suggestions / auto-fill ----
    const [equip, setEquip] = useState([]);
    useEffect(() => {
        let abort = false;
        (async () => {
            try {
                const r = await api('/api/dashboard-stat');
                const j = await r.json();
                if (!abort && j && Array.isArray(j.log_lab_equipment_master)) setEquip(j.log_lab_equipment_master);
            } catch {}
        })();
        return () => { abort = true; };
    }, []);

    // ---- Borrow form ----
    const [sku, setSku] = useState('');
    const [itemId, setItemId] = useState('');
    const [qty, setQty] = useState(1);
    const [due, setDue] = useState('');

    // suggestions
    const [showSug, setShowSug] = useState(false);
    const [sugIdx, setSugIdx] = useState(0);
    const sugWrapRef = useRef(null);
    const suggestions = React.useMemo(() => {
        const q = sku.trim().toLowerCase();
        if (!q) return [];
        return equip
            .map(e => ({ id: e.id, sku: String(e.sku||''), name: String(e.name||''), available: e.available_quantity ?? 0 }))
            .filter(e => e.sku.toLowerCase().includes(q) || e.name.toLowerCase().includes(q))
            .slice(0, 8);
    }, [sku, equip]);
    useEffect(() => {
        function onDoc(e) { if (sugWrapRef.current && !sugWrapRef.current.contains(e.target)) setShowSug(false); }
        document.addEventListener('click', onDoc);
        return () => document.removeEventListener('click', onDoc);
    }, []);
    const pickSuggestion = (s) => { setSku(s.sku); setItemId(String(s.id)); setShowSug(false); };
    const onSkuKeyDown = (e) => {
        if (!showSug || suggestions.length === 0) return;
        if (e.key === 'ArrowDown') { e.preventDefault(); setSugIdx(i => Math.min(i+1, suggestions.length-1)); }
        else if (e.key === 'ArrowUp') { e.preventDefault(); setSugIdx(i => Math.max(i-1, 0)); }
        else if (e.key === 'Enter') { e.preventDefault(); pickSuggestion(suggestions[sugIdx]); }
        else if (e.key === 'Escape') { setShowSug(false); }
    };

    // ---- Return form ----
    let today = new Date()
    const [borrowId, setBorrowId]   = useState('');
    const [retSku, setRetSku]       = useState('');
    const [retItemId, setRetItemId] = useState('');
    const [condition, setCondition] = useState('');
    const [returnedAt, setReturnedAt] = useState(today.toISOString().split('T')[0]);

    // ---- Scanner state (shared) ----
    const [borrowScanOpen, setBorrowScanOpen] = useState(false);
    const [returnScanOpen, setReturnScanOpen] = useState(false);
    const [scanErr, setScanErr] = useState('');
    const [scanning, setScanning] = useState(false);
    const [activeRegion, setActiveRegion] = useState(null); // 'borrow' | 'return' | null
    const scannerRef  = useRef(null);
    const pendingRegionRef = useRef(null); // region we want to start once DOM node exists

    // Start the camera AFTER the panel is rendered
    useEffect(() => {
        (async () => {
            const region = pendingRegionRef.current;
            if (!region) return;
            const divId = region === 'borrow' ? QR_BORROW_DIV_ID : QR_RETURN_DIV_ID;
            const host = document.getElementById(divId);
            if (!host) return; // wait next render
            pendingRegionRef.current = null;

            // Clean any previous instance
            if (scannerRef.current) { try { await scannerRef.current.stop(); } catch {} try { await scannerRef.current.clear(); } catch {} scannerRef.current = null; }

            try {
                const { Html5Qrcode } = await import('html5-qrcode');
                const inst = new Html5Qrcode(divId);
                scannerRef.current = inst;
                await inst.start(
                    { facingMode: 'environment' },
                    { fps: 10, qrbox: { width: 240, height: 240 } },
                    (text) => onScan(region, String(text||'').trim()),
                    () => {}
                );
                setActiveRegion(region);
                setScanning(true);
            } catch (e) {
                setScanErr(String(e?.message || e));
                setScanning(false);
                setActiveRegion(null);
            }
        })();
    }, [borrowScanOpen, returnScanOpen, tab]);

    async function stopScanner() {
        const inst = scannerRef.current;
        scannerRef.current = null;
        pendingRegionRef.current = null;
        try { if (inst?.stop)  await inst.stop(); } catch {}
        try { if (inst?.clear) await inst.clear(); } catch {}
        setScanning(false);
        setActiveRegion(null);
        setBorrowScanOpen(false);
        setReturnScanOpen(false);
    }

    // Open functions (set panel first, effect starts camera)
    function openBorrowScanner() {
        resetAlerts();
        setTab('borrow');
        setReturnScanOpen(false);
        setBorrowScanOpen(true);
        pendingRegionRef.current = 'borrow';
    }
    function openReturnScanner() {
        resetAlerts();
        setTab('return');
        setBorrowScanOpen(false);
        setReturnScanOpen(true);
        pendingRegionRef.current = 'return';
    }

    // Push decoded text into the corresponding input
    function onScan(region, text) {
        if (!text) return;
        const found = equip.find(e => String(e.sku) === text);
        if (region === 'borrow') {
            setSku(text);
            if (found) setItemId(String(found.id)); else setItemId('');
            // focus the field you wanted filled
            document.getElementById('borrow-sku')?.focus();
        } else {
            setRetSku(text);
            if (found) setRetItemId(String(found.id)); else setRetItemId('');
            document.getElementById('return-sku')?.focus();
        }
        stopScanner(); // stop after first decode
    }

    // Stop scanner when leaving page/tab
    useEffect(() => () => { stopScanner(); }, []);
    useEffect(() => { stopScanner(); }, [tab]);

    // ---- actions ----
    async function doBorrow(e) {
        e.preventDefault(); resetAlerts();
        if (needAuth()) return;
        const payload = {
            sku: sku.trim() || undefined,
            item_id: itemId.trim() ? Number(itemId) : undefined,
            quantity: Number(qty) || 1,
            return_date: due.trim() || undefined,
            user_id: userId || undefined,
            borrower_name: (borrowerName || '').trim() || undefined,
        };
        if (!payload.sku && !payload.item_id) { setError('Provide SKU or Item ID'); return; }
        try {
            const res = await api('/api/items/borrow', { method: 'POST', body: JSON.stringify(payload) });
            const j = await res.json().catch(()=> ({}));
            if (!res.ok) throw new Error(j.error || `HTTP ${res.status}`);
            setOk(`Borrowed record #${j.id} (item ${j.item_id})`);
        } catch (err) { setError(String(err.message || err)); }
    }
    async function doReturn(e) {
        e.preventDefault(); resetAlerts();
        if (needAuth()) return;
        const payload = {
            borrow_id: borrowId.trim() ? Number(borrowId) : undefined,
            sku: retSku.trim() || undefined,
            item_id: retItemId.trim() ? Number(retItemId) : undefined,
            condition_on_return: condition.trim() || undefined,
            returned_at: returnedAt.trim() || undefined,
            user_id: userId || undefined,   // <— add this
        };
        if (!payload.borrow_id && !payload.sku && !payload.item_id) { setError('Provide Borrow ID or SKU / Item ID'); return; }
        try {
            const res = await api('/api/items/return', { method: 'POST', body: JSON.stringify(payload) });
            const j = await res.json().catch(()=> ({}));
            if (!res.ok) throw new Error(j.error || `HTTP ${res.status}`);
            setOk('Return completed.');
        } catch (err) { setError(String(err.message || err)); }
    }

    // ---- render ----
    return (
        <div className="imx-container">
            <header className="imx-header">
                <div>
                    <h1 className="imx-title">Mượn / Trả</h1>
                    <p className="imx-subtitle">Gõ hoặc quét mã QR để tự động điền vào ô trống cho mỗi SKU.</p>
                </div>
                <nav className="imx-actions">
                    <Link className="imx-btn" to="/dashboard">Dashboard</Link>
                    <Link className="imx-btn" to="/">Trang chủ</Link>
                </nav>
            </header>

            {!token && (
                <div className="imx-alert imx-alert--error" style={{marginBottom: 12}}>
                    Chưa đăng nhập. <Link className="imx-link" to="/login">Đăng nhập</Link>
                </div>
            )}

            <div className="imx-card">
                <div className="imx-row" style={{gap:8, marginBottom:12, flexWrap:'wrap'}}>
                    <button type="button" className={`imx-btn ${tab==='borrow'?'imx-btn--primary':''}`} onClick={()=>{setTab('borrow'); resetAlerts();}}>Mượn</button>
                    <button type="button" className={`imx-btn ${tab==='return'?'imx-btn--primary':''}`} onClick={()=>{setTab('return'); resetAlerts();}}>Trả</button>
                </div>

                {error && <div className="imx-alert imx-alert--error">{error}</div>}
                {ok && <div className="imx-alert" style={{borderColor:'#2e63ff'}}>{ok}</div>}
                {scanErr && <div className="imx-alert imx-alert--error">{scanErr}</div>}

                {/* ===== Borrow ===== */}
                {tab === 'borrow' && (
                    <form className="imx-form" onSubmit={doBorrow}>
                        <div className="imx-row" style={{gap:12, alignItems:'flex-start'}} ref={sugWrapRef}>
                            <div style={{flex:1, position:'relative'}}>
                                <label className="imx-label">SKU hoặc Tên vật tư</label>
                                <div className="imx-row" style={{gap:8}}>
                                    <input
                                        id="borrow-sku"
                                        className="imx-input"
                                        value={sku}
                                        onChange={(e)=>{ setSku(e.target.value); setShowSug(true); setSugIdx(0); }}
                                        onFocus={()=> setShowSug(true)}
                                        onKeyDown={onSkuKeyDown}
                                        placeholder="Type SKU or item name…"
                                        autoComplete="off"
                                        style={{flex:1}}
                                    />
                                    <button
                                        type="button"
                                        className={`imx-btn ${borrowScanOpen ? 'imx-btn--primary':''}`}
                                        onClick={() => borrowScanOpen ? stopScanner() : openBorrowScanner()}
                                    >
                                        {borrowScanOpen && activeRegion==='borrow' && scanning ? 'Ngừng' : 'Quét QR'}
                                    </button>
                                </div>

                                {showSug && suggestions.length > 0 && (
                                    <div className="imx-card imx-sugs">
                                        {suggestions.map((s, i) => (
                                            <div
                                                key={`${s.id}-${s.sku}`}
                                                onMouseDown={(e)=> e.preventDefault()}
                                                onClick={()=> pickSuggestion(s)}
                                                className="imx-row imx-sugrow"
                                                style={{background: i===sugIdx ? 'rgba(79,124,255,.08)' : 'transparent'}}
                                            >
                                                <div style={{minWidth:110, fontWeight:700}}>{s.sku}</div>
                                                <div style={{flex:1}}>{s.name}</div>
                                                <div style={{opacity:.7, minWidth:90, textAlign:'right'}}>Sẵn sàng: {s.available}</div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                            <div style={{maxWidth:130}}>
                                <label className="imx-label">Người mượn</label>
                                <input style={{maxWidth:130}}
                                    className="imx-input"
                                    value={borrowerName || 'Chưa đăng nhập'}
                                    readOnly
                                    aria-readonly="true"
                                />
                            </div>
                            <div style={{maxWidth:130}}>
                                <label className="imx-label">ID vật tư (tự động)</label>
                                <input style={{maxWidth:130}}
                                       className="imx-input"
                                       value={itemId}
                                       readOnly
                                       aria-readonly="true"
                                       placeholder="optional" />
                            </div>

                            <div style={{maxWidth:130}}>
                                <label className="imx-label">Số lượng</label>
                                <input style={{maxWidth:130}} className="imx-input" type="number" min="1" value={qty} onChange={e=>setQty(e.target.value)} />
                            </div>

                            <div style={{maxWidth:170}}>
                                <label className="imx-label">Hạn trả</label>
                                <input style={{maxWidth:130}} className="imx-input" type="date" value={due} onChange={e=>setDue(e.target.value)} />
                            </div>
                        </div>

                        {borrowScanOpen && (
                            <div className="imx-qrwrap">
                                <div className="imx-qrcard">
                                    <div id={QR_BORROW_DIV_ID} className="imx-qrbox" />
                                    {!(scanning && activeRegion==='borrow') && (
                                        <div className="imx-qr-overlay">Camera chưa sẵn sàng</div>
                                    )}
                                </div>
                                <div className="imx-qrhelp">
                                    <div className="imx-label" style={{marginBottom:8}}>Đã quét SKU</div>
                                    <input className="imx-input" value={sku} onChange={e=>setSku(e.target.value)} placeholder="Waiting for scan…" />
                                </div>
                            </div>
                        )}

                        <div className="imx-row" style={{gap:10, justifyContent:'flex-end'}}>
                            <button className="imx-btn imx-btn--primary" type="submit">Mượn</button>
                        </div>
                    </form>
                )}

                {/* ===== Return ===== */}
                {tab === 'return' && (
                    <form className="imx-form" onSubmit={doReturn}>
                        <div className="imx-row" style={{gap:12, flexWrap:'wrap'}}>
                            <div style={{width:200}}>
                                <label className="imx-label">ID mượn</label>
                                <input className="imx-input"
                                       value={borrowerName || 'Chưa đăng nhập'}
                                       onChange={e=>setReturnedAt(e.target.value)}
                                       placeholder="preferred if multiple" />
                            </div>

                            <div style={{flex:1, minWidth:260}}>
                                <label className="imx-label">SKU</label>
                                <div className="imx-row" style={{gap:8}}>
                                    <input id="return-sku" className="imx-input" value={retSku} onChange={e=>setRetSku(e.target.value)} placeholder="or use Item ID" style={{flex:1}} />
                                    <button
                                        type="button"
                                        className={`imx-btn ${returnScanOpen ? 'imx-btn--primary':''}`}
                                        onClick={() => returnScanOpen ? stopScanner() : openReturnScanner()}
                                    >
                                        {returnScanOpen && activeRegion==='return' && scanning ? 'Ngừng' : 'Quét QR'}
                                    </button>
                                </div>
                            </div>

                            <div style={{maxWidth:180}}>
                                <label className="imx-label">ID Vât tư (tự động)</label>
                                <input style={{maxWidth:180}}
                                       className="imx-input"
                                       value={retItemId}
                                       readOnly
                                       aria-readonly="true"
                                       placeholder="optional" />
                            </div>

                            <div style={{maxWidth:300}}>
                                <label className="imx-label">Trả vào</label>
                                <input style={{maxWidth:180}} className="imx-input"
                                       type="date"
                                       value={returnedAt}
                                       onChange={e=>setReturnedAt(e.target.value)}
                                />
                            </div>
                        </div>

                        {returnScanOpen && (
                            <div className="imx-qrwrap">
                                <div className="imx-qrcard">
                                    <div id={QR_RETURN_DIV_ID} className="imx-qrbox" />
                                    {!(scanning && activeRegion==='return') && (
                                        <div className="imx-qr-overlay">Camera chưa sẵn sàng</div>
                                    )}
                                </div>
                                <div className="imx-qrhelp">
                                    <div className="imx-label" style={{marginBottom:8}}>Đã quét SKU</div>
                                    <input className="imx-input" value={retSku} onChange={e=>setRetSku(e.target.value)} placeholder="Waiting for scan…" />
                                </div>
                            </div>
                        )}

                        <div>
                            <label className="imx-label">Điều kiện hoàn trả</label>
                            <input className="imx-input" value={condition} onChange={e=>setCondition(e.target.value)} placeholder="e.g. Good / minor scratch" />
                        </div>

                        <div className="imx-row" style={{gap:10, justifyContent:'flex-end'}}>
                            <button className="imx-btn imx-btn--primary" type="submit">Trả</button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    );
}
