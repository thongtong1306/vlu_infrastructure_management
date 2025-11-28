import React, { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { marked } from "marked";
import DOMPurify from "dompurify";

// --- tiny auth-aware fetch helpers (same pattern you already use) ---
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
    const text = await r.text();
    const ct = r.headers.get("content-type") || "";
    if (!ct.includes("application/json")) throw new Error(`Expected JSON, got ${ct}: ${text.slice(0,120)}`);
    return JSON.parse(text);
}
// -------------------------------------------------------------------

export default function InstructionView() {
    const { id } = useParams();           // instruction id
    const nav = useNavigate();
    const [instr, setInstr] = useState(null);
    const [item, setItem] = useState(null);
    const [siblings, setSiblings] = useState([]); // other instructions for same item
    const [err, setErr] = useState("");

    useEffect(() => {
        document.title = "Instruction";
        (async () => {
            try {
                // 1) load instruction
                const ins = await getJSON(`/api/instructions/${id}`);
                setInstr(ins);
                // 2) item info (for sidebar + breadcrumb)
                const it = await getJSON(`/api/items/${ins.item_id}`);
                setItem(it);
                // 3) list all instructions for prev/next
                const list = await getJSON(`/api/instructions?item_id=${ins.item_id}`);
                setSiblings(Array.isArray(list) ? list : (list.instructions || []));
            } catch (e) {
                setErr(String(e.message || e));
            }
        })();
    }, [id]);

    const { prevId, nextId } = useMemo(() => {
        if (!instr || siblings.length === 0) return { prevId: null, nextId: null };
        const idx = siblings.findIndex(x => Number(x.id) === Number(instr.id));
        return {
            prevId: idx >= 0 && idx < siblings.length - 1 ? siblings[idx + 1].id : null,
            nextId: idx > 0 ? siblings[idx - 1].id : null,
        };
    }, [instr, siblings]);

    const html = useMemo(() => {
        if (!instr?.body) return "";
        // treat body as Markdown (supports images, lists, bold, etc.)
        const raw = marked.parse(instr.body);
        return DOMPurify.sanitize(raw);
    }, [instr]);

    return (
        <div className="imx-container imx-doc">
            <header className="imx-header">
                <div>
                    <div className="imx-breadcrumb">
                        <Link to="/">Home</Link> <span>›</span>
                        <Link to="/equipments">Equipments</Link> <span>›</span>
                        {item ? <Link to={`/equipments?item_id=${item.id}`}>{item.name || `Item #${item.id}`}</Link> : "…"}
                    </div>
                    <h1 className="imx-title" style={{ marginTop: 8 }}>
                        {instr?.title || "Instruction"}
                    </h1>
                    {item && <p className="imx-subtitle">for {item.name || `Item #${item.id}`} (ID #{item.id})</p>}
                </div>
                <nav className="imx-actions">
                    {prevId && <button className="imx-btn" onClick={() => nav(`/instructions/${prevId}`)}>← Previous</button>}
                    {nextId && <button className="imx-btn" onClick={() => nav(`/instructions/${nextId}`)}>Next →</button>}
                    <button className="imx-btn" onClick={() => window.print()}>Print</button>
                    <Link className="imx-btn" to="/equipments">Back to Equipments</Link>
                </nav>
            </header>

            {err && <div className="imx-alert imx-alert--error">Error: {err}</div>}

            <div className="imx-grid imx-grid--two">
                {/* Main doc */}
                <article className="imx-card">
                    {instr?.image_url && (
                        <div className="imx-image-frame" style={{ margin: 16 }}>
                            <img
                                alt={instr.title || "instruction image"}
                                src={instr.image_url}
                                onError={(e)=>{ e.currentTarget.style.display='none'; }}
                                style={{ width:"100%", borderRadius: 12 }}
                            />
                        </div>
                    )}
                    <div className="imx-card__body" style={{ paddingTop: 0 }}>
                        <div
                            className="imx-prose"
                            style={{ lineHeight: 1.6, fontSize: 16 }}
                            dangerouslySetInnerHTML={{ __html: html }}
                        />
                    </div>
                </article>

                {/* Side info */}
                <aside className="imx-card">
                    <div className="imx-card__header">
                        <h2 className="imx-card__title">Details</h2>
                    </div>
                    <div className="imx-card__body">
                        <table className="imx-meta">
                            <tbody>
                            <tr><th>Instruction ID</th><td>{instr?.id ?? "—"}</td></tr>
                            <tr><th>Item</th><td>#{item?.id ?? "—"}</td></tr>
                            <tr><th>Created</th><td>{instr?.created_at ? new Date(instr.created_at).toLocaleString() : "—"}</td></tr>
                            <tr><th>Updated</th><td>{instr?.updated_at ? new Date(instr.updated_at).toLocaleString() : "—"}</td></tr>
                            </tbody>
                        </table>
                        {item && (
                            <>
                                <hr className="imx-hr" />
                                <div className="imx-subtitle" style={{ marginBottom: 8 }}>Equipment</div>
                                <div style={{ display: "flex", gap: 12 }}>
                                    <img
                                        alt={item.name || "equipment"}
                                        src={item.image_url || `/static/img/equipment/${item.id}.jpg`}
                                        onError={(e)=>{ e.currentTarget.src="/static/img/equipment/no-image.jpg"; }}
                                        style={{ width: 120, height: 90, objectFit: "cover", borderRadius: 8 }}
                                    />
                                    <div>
                                        <div><strong>{item.name || "—"}</strong></div>
                                        <div className="imx-meta-mini">SKU: {item.sku || "—"}</div>
                                        <div className="imx-meta-mini">{item.category || "—"} · {item.location || "—"}</div>
                                    </div>
                                </div>
                            </>
                        )}
                    </div>
                </aside>
            </div>
        </div>
    );
}
