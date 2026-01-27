import React, { useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useNavigate } from "react-router-dom";

// simple auth-aware fetch
function authHeaders(extra = {}) {
    let token = "";
    try {
        const s = JSON.parse(localStorage.getItem("imx_session") || "{}");
        token = s.token || localStorage.getItem("token") || "";
    } catch {}
    const h = { "Content-Type": "application/json", ...extra };
    if (token) h.Authorization = `Bearer ${token}`;
    return h;
}
async function getJSON(url) {
    const r = await fetch(url, { headers: authHeaders(), credentials: "include" });
    if (r.status === 401) { window.location.replace("/login"); throw new Error("401"); }
    const ct = r.headers.get("content-type") || "";
    const text = await r.text();
    if (!ct.includes("application/json")) throw new Error(`Expected JSON, got ${ct}: ${text.slice(0,120)}`);
    return JSON.parse(text);
}

export default function Equipments() {
    const [items, setItems] = useState([]);
    const [filter, setFilter] = useState("");
    const [selectedId, setSelectedId] = useState(null);
    const [instructions, setInstructions] = useState([]);
    const [notes, setNotes] = useState([]);
    const [err, setErr] = useState("");
    const [loading, setLoading] = useState(true);
    const [params, setParams] = useSearchParams();

    useEffect(() => {
        document.title = "Infrastructure | Equipments";
        (async () => {
            try {
                setLoading(true);
                // your ItemController.GetAll → it returns array of log_lab_equipment_master rows
                const data = await getJSON("/api/items");
                const list = Array.isArray(data) ? data : (data.items || []);
                setItems(list);
                const qid = Number(params.get("item_id") || 0);
                setSelectedId(qid || list?.[0]?.id || null);
            } catch (e) {
                setErr(e.message || "Failed to load equipment list");
            } finally {
                setLoading(false);
            }
        })();
    }, []);

    // load instructions + notes for selected item
    useEffect(() => {
        if (!selectedId) { setInstructions([]); setNotes([]); return; }
        params.set("item_id", String(selectedId)); setParams(params, { replace: true });

        (async () => {
            try {
                const instr = await getJSON(`/api/instructions?item_id=${selectedId}`);
                setInstructions(Array.isArray(instr) ? instr : (instr.instructions || []));
            } catch (e) { setInstructions([]); }

            try {
                const ns = await getJSON(`/api/equipment-notes?item_id=${selectedId}`);
                setNotes(Array.isArray(ns) ? ns : (ns.notes || []));
            } catch (e) { setNotes([]); }
        })();
    }, [selectedId]);

    const filtered = useMemo(() => {
        const q = filter.trim().toLowerCase();
        if (!q) return items;
        return items.filter(it => {
            const s = `${it.name||""} ${it.sku||""} ${it.category||""} ${it.location||""} ${it.status||""}`.toLowerCase();
            return s.includes(q);
        });
    }, [items, filter]);

    const navigate = useNavigate();

    const current = useMemo(() => items.find(x => Number(x.id) === Number(selectedId)), [items, selectedId]);
    console.log(current)
    return (
        <div className="imx-container">
            <header className="imx-header">
                <div>
                    <h1 className="imx-title">Tất cả vật tư</h1>
                    <p className="imx-subtitle">Tìm tất cả vật tư. Xem thông tin, các lưu ý, và các hướng dẫn sử dụng.</p>
                </div>
                <nav className="imx-actions">
                    <Link className="imx-btn" to="/dashboard">Dashboard</Link>
                    <Link className="imx-btn" to="/labs">Tất cả Phòng thí nghiệm</Link>
                    <Link className="imx-btn" to="/">Trang chủ</Link>
                </nav>
            </header>

            {err && <div className="imx-alert imx-alert--error">Error: {String(err)}</div>}

            <div className="imx-grid imx-grid--two">
                {/* Left column: list + search */}
                <div className="imx-card">
                    <div className="imx-card__header">
                        <h2 className="imx-card__title">Danh sách vật tư</h2>
                        <input
                            className="imx-input"
                            placeholder="Tìm theo tên, SKU, phân loại, vị trí…"
                            value={filter}
                            onChange={e => setFilter(e.target.value)}
                            style={{ maxWidth: 340 }}
                        />
                    </div>

                    <div className="imx-table-wrap" style={{maxHeight:560, overflow:"auto"}}>
                        <table className="imx-table imx-table--clickable">
                            <thead>
                            <tr>
                                <th width="60">ID</th>
                                <th>Tên vật tư</th>
                                <th width="140">SKU</th>
                                <th width="140">Phân loại</th>
                                <th width="140">Vị trí</th>
                                <th width="100">Trạng thái</th>
                            </tr>
                            </thead>
                            <tbody>
                            {filtered.length === 0 && (
                                <tr><td colSpan={6} style={{opacity:.7}}>Không có kết quả.</td></tr>
                            )}
                            {filtered.map(it => (
                                <tr
                                    key={it.id}
                                    onClick={() => setSelectedId(it.id)}
                                    className={Number(it.id) === Number(selectedId) ? "is-active" : ""}
                                    style={{cursor:"pointer"}}
                                >
                                    <td>{it.id}</td>
                                    <td>{it.name || "—"}</td>
                                    <td>{it.sku || "—"}</td>
                                    <td>{it.category || "—"}</td>
                                    <td>{it.location || "—"}</td>
                                    <td>{it.status || "—"}</td>
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    </div>
                </div>

                {/* Right column: details + notes + instructions */}
                <div className="imx-stack">
                    {/* Info card */}
                    <div className="imx-card">
                        <div className="imx-card__header">
                            <h2 className="imx-card__title">Thông tin chi tiết</h2>
                            {current && <span className="imx-subtitle">Vật tư số #{current.id}</span>}
                        </div>
                        <div className="imx-card__body">
                            {current ? (
                                <div className="imx-grid imx-grid--two">
                                    <div>
                                        <div className="imx-image-frame" style={{ marginBottom: 12 }}>
                                            <img
                                                alt={current?.name || "equipment"}
                                                src={
                                                    current?.image_url
                                                    || `/static/img/equipment/${current?.id}.jpg`
                                                }
                                                onError={(e)=>{ e.currentTarget.src = "/static/img/no-image.jpg"; }}
                                                style={{ width:"100%", borderRadius: 12 }}
                                            />

                                        </div>
                                        <div className="imx-subtitle">{current.description || "—"}</div>
                                    </div>
                                    <div>
                                        <table className="imx-meta">
                                            <tbody>
                                            <tr><th>SKU</th><td>{current.sku || "—"}</td></tr>
                                            <tr><th>Phân loại</th><td>{current.category || "—"}</td></tr>
                                            <tr><th>Vị trí</th><td>{current.location || "—"}</td></tr>
                                            <tr><th>Trạng thái</th><td>{current.status || "—"}</td></tr>
                                            <tr><th>Số lượng</th><td>{current.quantity ?? "—"}</td></tr>
                                            <tr><th>Sẵn sàng</th><td>{current.available_quantity ?? "—"}</td></tr>
                                            <tr><th>Đơn giá</th><td>{current.unit_cost ?? "—"}</td></tr>
                                            <tr><th>Nhà cung cấp</th><td>{current.supplier || "—"}</td></tr>
                                            <tr><th>Ngày mua</th><td>{current.date_purchased ? new Date(current.date_purchased).toLocaleDateString() : "—"}</td></tr>
                                            <tr><th>Ngày lưu kho</th><td>{current.create_at ? new Date(current.create_at).toLocaleString() : "—"}</td></tr>
                                            </tbody>
                                        </table>
                                    </div>
                                </div>
                            ) : <div className="imx-subtitle">Chọn một vật tư.</div>}
                        </div>
                    </div>

                    {/* Notes */}
                    <div className="imx-card">
                        <div className="imx-card__header">
                            <h2 className="imx-card__title">Lưu ý</h2>
                            <span className="imx-subtitle">{selectedId ? `Vật tư #${selectedId}` : ""}</span>
                        </div>
                        <div className="imx-table-wrap">
                            <table className="imx-table">
                                <thead><tr><th width="80">ID</th><th>Lưu ý</th><th width="180">Ngày tạo</th></tr></thead>
                                <tbody>
                                {notes.length === 0 && <tr><td colSpan={3} style={{opacity:.7}}>Không có lưu ý nào.</td></tr>}
                                {notes.map(n => (
                                    <tr key={n.id}>
                                        <td>{n.id}</td><td>{n.note_text}</td>
                                        <td>{n.created_at ? new Date(n.created_at).toLocaleString() : ""}</td>
                                    </tr>
                                ))}
                                </tbody>
                            </table>
                        </div>
                    </div>

                    {/* Instructions */}
                    <div className="imx-card">
                        <div className="imx-card__header">
                            <h2 className="imx-card__title">Hướng dẫn sử dụng</h2>
                            <span className="imx-subtitle">{selectedId ? `Vật tư #${selectedId}` : ""}</span>
                        </div>
                        <div className="imx-table-wrap">
                            <table className="imx-table">
                                <thead><tr><th width="80">ID</th><th>Hướng dẫn</th><th width="180">Ngày tạo</th></tr></thead>
                                <tbody>
                                {instructions.length === 0 && <tr><td colSpan={3} style={{opacity:.7}}>Không có hướng dẫn sử dụng nào.</td></tr>}
                                {instructions.map(row => (
                                    <tr key={row.id} onClick={()=>navigate(`/instructions/${row.id}`)} style={{cursor:"pointer"}}>
                                        <td>{row.id}</td><td>{row.title}</td>
                                        <td>{row.created_at ? new Date(row.created_at).toLocaleString() : ""}</td>
                                    </tr>
                                ))}
                                </tbody>
                            </table>
                        </div>
                    </div>

                </div>
            </div>
        </div>
    );
}
