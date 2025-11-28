// pages/AddItem.js
import React, { useState, useMemo, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';

export default function AddItem() {
    const nav = useNavigate();
    useEffect(() => { document.title = 'Add Item | Infrastructure'; }, []);

    // auth helpers
    const session = useMemo(() => {
        try { return JSON.parse(localStorage.getItem('imx_session') || 'null'); } catch { return null; }
    }, []);
    const token = session?.token;

    const [form, setForm] = useState({
        sku: '',
        name: '',
        description: '',
        category: '',
        location: '',
        quantity: 0,
        available_quantity: 0,
        unit_cost: 0,
        supplier: '',
        date_purchased: '', // yyyy-mm-dd
        status: 'active',
    });
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState('');

    const setField = (k, v) => setForm(s => ({ ...s, [k]: v }));

    const validate = () => {
        if (!form.sku.trim()) return 'SKU is required';
        if (!form.name.trim()) return 'Name is required';
        const qty = Number(form.quantity || 0);
        const avail = Number(form.available_quantity || 0);
        const cost = Number(form.unit_cost || 0);
        if (qty < 0) return 'Quantity must be >= 0';
        if (avail < 0 || avail > qty) return 'Available must be between 0 and Quantity';
        if (cost < 0) return 'Unit cost must be >= 0';
        if (form.date_purchased && !/^\d{4}-\d{2}-\d{2}$/.test(form.date_purchased)) {
            return 'Date must be YYYY-MM-DD';
        }
        return '';
    };

    const onSubmit = async (e) => {
        e.preventDefault();
        setError('');
        const v = validate();
        if (v) { setError(v); return; }
        if (!token) { setError('Please sign in first.'); return; }

        setSubmitting(true);
        try {
            const res = await fetch('/api/items', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...(token ? { Authorization: `Bearer ${token}` } : {}),
                },
                body: JSON.stringify({
                    sku: form.sku.trim(),
                    name: form.name.trim(),
                    description: form.description.trim(),
                    category: form.category.trim(),
                    location: form.location.trim(),
                    quantity: Number(form.quantity || 0),
                    available_quantity: Number(form.available_quantity || 0),
                    unit_cost: Number(form.unit_cost || 0),
                    supplier: form.supplier.trim(),
                    date_purchased: form.date_purchased.trim(), // optional
                    status: form.status.trim() || 'active',
                }),
            });
            if (!res.ok) {
                const j = await res.json().catch(() => ({}));
                throw new Error(j.error || `HTTP ${res.status}`);
            }
            await res.json(); // created row (unused)
            nav('/dashboard');
        } catch (err) {
            setError(String(err.message || err));
        } finally {
            setSubmitting(false);
        }
    };

    if (!token) {
        return (
            <div className="imx-container">
                <header className="imx-header">
                    <div>
                        <h1 className="imx-title">Nhập vật tư</h1>
                        <p className="imx-subtitle">Hãy đăng nhập để sử dụng chức năng này.</p>
                    </div>
                    <nav className="imx-actions">
                        <Link className="imx-btn" to="/">Trang chủ</Link>
                        <Link className="imx-btn imx-btn--primary" to="/login">Đăng nhập</Link>
                    </nav>
                </header>
                <div className="imx-alert imx-alert--error">Not signed in.</div>
            </div>
        );
    }

    return (
        <div className="imx-container">
            <header className="imx-header">
                <div>
                    <h1 className="imx-title">Nhập vật tư</h1>
                    <p className="imx-subtitle">Tạo dữ liệu cho vật tư mới.</p>
                </div>
                <nav className="imx-actions">
                    <Link className="imx-btn" to="/dashboard">Dashboard</Link>
                    <Link className="imx-btn" to="/">Trang chủ</Link>
                </nav>
            </header>

            <form className="imx-card imx-form" onSubmit={onSubmit}>
                {error && <div className="imx-alert imx-alert--error">{error}</div>}

                <div className="imx-row" style={{gap:12}}>
                    <div style={{flex:1}}>
                        <label className="imx-label">SKU *</label>
                        <input className="imx-input" value={form.sku} onChange={e=>setField('sku', e.target.value)} placeholder="e.g. 23000140" />
                    </div>
                    <div style={{flex:2}}>
                        <label className="imx-label">Tên vật tư *</label>
                        <input className="imx-input" value={form.name} onChange={e=>setField('name', e.target.value)} placeholder="Oscilloscope" />
                    </div>
                </div>

                <div>
                    <label className="imx-label">Chú thích</label>
                    <textarea className="imx-input" rows={3} value={form.description} onChange={e=>setField('description', e.target.value)} placeholder="Model / details…" />
                </div>

                <div className="imx-row" style={{gap:12}}>
                    <div style={{flex:1}}>
                        <label className="imx-label">Phân loại</label>
                        <input className="imx-input" value={form.category} onChange={e=>setField('category', e.target.value)} placeholder="equipment / tool / material" />
                    </div>
                    <div style={{flex:1}}>
                        <label className="imx-label">Vị trí</label>
                        <input className="imx-input" value={form.location} onChange={e=>setField('location', e.target.value)} placeholder="shelf_A1" />
                    </div>
                </div>

                <div className="imx-row" style={{gap:12}}>
                    <div style={{flex:1}}>
                        <label className="imx-label">Số lượng</label>
                        <input className="imx-input" type="number" min="0" value={form.quantity} onChange={e=>setField('quantity', e.target.value)} />
                    </div>
                    <div style={{flex:1}}>
                        <label className="imx-label">Sẵn sàng</label>
                        <input className="imx-input" type="number" min="0" value={form.available_quantity} onChange={e=>setField('available_quantity', e.target.value)} />
                    </div>
                    <div style={{flex:1}}>
                        <label className="imx-label">Đơn giá</label>
                        <input className="imx-input" type="number" min="0" step="0.01" value={form.unit_cost} onChange={e=>setField('unit_cost', e.target.value)} />
                    </div>
                </div>

                <div className="imx-row" style={{gap:12}}>
                    <div style={{flex:1}}>
                        <label className="imx-label">Nhà cung cấp</label>
                        <input className="imx-input" value={form.supplier} onChange={e=>setField('supplier', e.target.value)} placeholder="Supplier name" />
                    </div>
                    <div style={{flex:1}}>
                        <label className="imx-label">Ngày mua</label>
                        <input className="imx-input" type="date" value={form.date_purchased} onChange={e=>setField('date_purchased', e.target.value)} />
                    </div>
                    <div style={{flex:1}}>
                        <label className="imx-label">Trạng thái</label>
                        <select className="imx-input" value={form.status} onChange={e=>setField('status', e.target.value)}>
                            <option value="active">active</option>
                            <option value="retired">retired</option>
                            <option value="maintenance">maintenance</option>
                        </select>
                    </div>
                </div>

                <div className="imx-row" style={{gap:10, justifyContent:'flex-end'}}>
                    <Link className="imx-btn" to="/dashboard">Huỷ</Link>
                    <button className="imx-btn imx-btn--primary" type="submit" disabled={submitting}>
                        {submitting ? 'Đang lưu…' : 'Lưu vật tư'}
                    </button>
                </div>
            </form>
        </div>
    );
}
