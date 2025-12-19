// pages/Dashboard.js
import React, { Component } from 'react';
import { Link } from 'react-router-dom';
import Highcharts from 'highcharts';
import HighchartsReact from 'highcharts-react-official';
import ThemeToggle from "./ThemeToggle";

export default class Dashboard extends Component {
    state = {
        loading: true,
        error: null,
        data: null,              // full JSON from /api/dashboard-stat
        sortKey: 'id',
        sortDir: 'desc',
        query: '',
        statusFilter: 'all',
    };
    
    componentDidMount() {
        document.title = 'Infrastructure Management | Dashboard';
        this.load();
    }

    componentDidUpdate(prevProps, prevState) {
        // when data finishes loading, honor #equipment deep-link
            if (prevState.loading && !this.state.loading) {
                if (window.location.hash === '#equipment') {
                    const el = document.getElementById('equipment');
                    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }
            }
    }

    // ---------- auth helpers ----------
    getSession() {
        try { return JSON.parse(localStorage.getItem('imx_session') || 'null'); }
        catch { return null; }
    }
    isSignedIn() { return !!this.getSession()?.token; }
    tokenHeader() {
        const t = this.getSession()?.token;
        return t ? { Authorization: `Bearer ${t}` } : {};
    }

    // ---------- data load ----------
    async load() {
        try {
            const res = await fetch('/api/dashboard-stat', { headers: { ...this.tokenHeader() } });
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            const data = await res.json();
            this.setState({ data, loading: false });
        } catch (err) {
            this.setState({ error: String(err), loading: false });
        }
    }

    // ---------- utils ----------
    fmtDate(s) {
        if (!s) return '-';
        try { return new Date(s).toLocaleString(); } catch { return String(s); }
    }
    kpis() {
        const d = this.state.data;
        if (!d) return { total: 0, available: 0, borrowed: 0, overdue: 0, maintenanceOpen: 0, utilization: 0 };

        const eq = d.log_lab_equipment_master || [];
        const br = d.log_lab_borrow_records || [];
        const mt = d.log_lab_maintenance_records || [];

        const total = eq.length;
        const available = eq.reduce((acc, e) => acc + (Number(e.available_quantity || 0) > 0 ? 1 : 0), 0);

        const now = Date.now();
        const borrowedActive = br.filter(r => !r.actual_return_date && (r.status || '').toLowerCase() !== 'returned').length;
        const overdue = br.filter(r => {
            if (r.actual_return_date) return false;
            if (!r.return_date) return false;
            return new Date(r.return_date).getTime() < now;
        }).length;

        const maintenanceOpen = mt.filter(m => !m.date_fixed).length;
        const utilization = total ? Math.round((borrowedActive / total) * 100) : 0;

        return { total, available, borrowed: borrowedActive, overdue, maintenanceOpen, utilization };
    }

    // ---------- sort helpers ----------
    setSort = (key) => {
        this.setState(s => ({
            sortKey: key,
            sortDir: s.sortKey === key && s.sortDir === 'desc' ? 'asc' : 'desc'
        }));
    };
    getAriaSort = (key) =>
        this.state.sortKey === key ? (this.state.sortDir === 'asc' ? 'ascending' : 'descending') : 'none';

    sortHeader = (label, key) => (
        <th
            key={`hdr-${key}`}
            scope="col"
            data-sortable="true"
            aria-sort={this.getAriaSort(key)}
            tabIndex={0}
            onClick={() => this.setSort(key)}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); this.setSort(key); }}}
        >
            {label}
        </th>
    );

    sortRows(rows) {
        const { sortKey, sortDir } = this.state;
        const dir = sortDir === 'asc' ? 1 : -1;
        return [...(rows || [])].sort((a, b) => {
            const av = a?.[sortKey] ?? '';
            const bv = b?.[sortKey] ?? '';
            const at = Date.parse(av), bt = Date.parse(bv);
            if (!Number.isNaN(at) && !Number.isNaN(bt)) return (at - bt) * dir;
            if (av === bv) return 0;
            return av > bv ? dir : -dir;
        });
    }

    // ---------- chart helpers ----------
    ymd(d) {
        const y = d.getFullYear();
        const m = `${d.getMonth() + 1}`.padStart(2, '0');
        const day = `${d.getDate()}`.padStart(2, '0');
        return `${y}-${m}-${day}`;
    }

    // Utilization series: last N days, % borrowed items
    buildUtilSeries(days = 30) {
        const d = this.state.data;
        if (!d) return { categories: [], series: [] };

        const eq = d.log_lab_equipment_master || [];
        const br = d.log_lab_borrow_records || [];
        const total = eq.length || 1;

        // intervals: [borrow_date, actual_return_date || now)
        const intervals = br.map(r => ({
            start: new Date(r.borrow_date).getTime(),
            end: r.actual_return_date ? new Date(r.actual_return_date).getTime() : Date.now(),
            qty: Number(r.quantity || 1),
        }));

        const cats = [];
        const vals = [];
        const today = new Date(); today.setHours(0,0,0,0);
        for (let i = days - 1; i >= 0; i--) {
            const dt = new Date(today); dt.setDate(today.getDate() - i);
            const ts = dt.getTime();
            const borrowedQty = intervals.reduce((acc, it) => (it.start <= ts && ts < it.end) ? acc + it.qty : acc, 0);
            const util = Math.max(0, Math.min(100, Math.round((borrowedQty / total) * 100)));
            cats.push(this.ymd(dt));
            vals.push(util);
        }
        return { categories: cats, series: [{ name: 'Utilization', data: vals }] };
    }

    // Category bars: Available vs In-use
    buildCategoryBars() {
        const eq = this.state.data?.log_lab_equipment_master || [];
        const byCat = new Map();
        eq.forEach(e => {
            const cat = (e.category || 'Uncategorized').toString();
            const total = Number(e.quantity || 0);
            const avail = Number(e.available_quantity || 0);
            const inUse = Math.max(0, total - avail);
            const cur = byCat.get(cat) || { category: cat, available: 0, in_use: 0, total: 0 };
            cur.available += avail; cur.in_use += inUse; cur.total += total;
            byCat.set(cat, cur);
        });
        const arr = [...byCat.values()].sort((a,b)=> b.total - a.total);
        return {
            categories: arr.map(x => x.category),
            series: [
                { name: 'Available', data: arr.map(x => x.available) },
                { name: 'In use', data: arr.map(x => x.in_use) },
            ]
        };
    }

    // Low stock (Top 10)
    buildLowStockBars() {
        const eq = this.state.data?.log_lab_equipment_master || [];
        const rows = [...eq]
            .map(e => ({
                label: e.name || e.sku || String(e.id),
                available: Number(e.available_quantity || 0),
                total: Number(e.quantity || 0),
            }))
            .sort((a,b)=> a.available - b.available)
            .slice(0, 10);

        return {
            categories: rows.map(r => r.label),
            series: [
                { name: 'Available', data: rows.map(r => r.available) },
                { name: 'Total', data: rows.map(r => r.total) },
            ]
        };
    }

    // Small helper to match dark theme
    themeColors() {
        const root = getComputedStyle(document.documentElement);
        const text = root.getPropertyValue('--text')?.trim() || '#e8edf6';
        const grid = '#2a3250';
        return { text, grid };
    }

    // ---------- view helpers ----------
    renderTableCard(title, rows, columns, domId) {
        return (
            <div className="imx-card" key={title} id={domId}>
                <div className="imx-card__header">
                    <h2 className="imx-card__title">{title}</h2>
                    <span className="imx-subtitle">{rows?.length ?? 0} rows</span>
                </div>
                <div className="imx-table-wrap">
                    {rows && rows.length ? (
                        <table className="imx-table">
                            <thead>
                            <tr>{columns.map(c => this.sortHeader(c.label, c.key))}</tr>
                            </thead>
                            <tbody>
                            {this.sortRows(rows).map((r, i) => (
                                <tr key={r.id ?? i}>
                                    {columns.map(col => {
                                        const val = r[col.key];
                                        const v = col.type === 'date' ? this.fmtDate(val) : (val ?? '-');
                                        return <td key={col.key}>{v}</td>;
                                    })}
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    ) : <div className="imx-subtitle">No data.</div>}
                </div>
            </div>
        );
    }

    render() {
        const { loading, error, data } = this.state;
        const signedIn = this.isSignedIn();

        // quick lookup for joins
        const eq = data?.log_lab_equipment_master || [];
        const equipById = new Map(eq.map(e => [e.id, e]));

        const { total, available, borrowed, overdue, maintenanceOpen, utilization } = this.kpis();

        // borrow view rows with item info & flags
        const borrowRows = (data?.log_lab_borrow_records || []).map(r => {
            const item = equipById.get(r.item_id);
            return {
                ...r,
                item_sku: item?.sku || '-',
                item_name: item?.name || '-',
                overdue_flag: (!r.actual_return_date && r.return_date && Date.parse(r.return_date) < Date.now()) ? 'OVERDUE' : '',
            };
        });

        // maintenance rows with state
        const maintRows = (data?.log_lab_maintenance_records || []).map(m => ({
            ...m,
            item_sku: equipById.get(m.item_id)?.sku || '-',
            item_name: equipById.get(m.item_id)?.name || '-',
            state: m.date_fixed ? 'closed' : 'open',
        }));

        // charts (only after data)
        const { text, grid } = this.themeColors();
        const util = !loading && !error ? this.buildUtilSeries(30) : { categories: [], series: [] };
        const cat  = !loading && !error ? this.buildCategoryBars() : { categories: [], series: [] };
        const low  = !loading && !error ? this.buildLowStockBars() : { categories: [], series: [] };

        const utilOptions = {
            chart: { type: 'line', backgroundColor: 'transparent', height: 260 },
            title: { text: 'Độ tận dụng (30 ngày trước)', style: { color: text, fontSize: '14px' } },
            xAxis: { categories: util.categories, labels: { style: { color: text } }, lineColor: grid, tickColor: grid },
            yAxis: { title: { text: 'Độ tận dụng %', style: { color: text } }, max: 100, labels: { style: { color: text } }, gridLineColor: grid },
            tooltip: { valueSuffix: '%', backgroundColor: '#0b1120', borderColor: grid, style: { color: text } },
            legend: { enabled: false },
            credits: { enabled: false },
            series: util.series
        };

        const catOptions = {
            chart: { type: 'column', backgroundColor: 'transparent', height: 260 },
            title: { text: 'Tồn kho theo loại', style: { color: text, fontSize: '14px' } },
            xAxis: { categories: cat.categories, labels: { style: { color: text } }, lineColor: grid, tickColor: grid },
            yAxis: { min: 0, title: { text: 'Vật tư', style: { color: text } }, labels: { style: { color: text } }, gridLineColor: grid },
            tooltip: { shared: true, backgroundColor: '#0b1120', borderColor: grid, style: { color: text } },
            plotOptions: { column: { stacking: 'normal', borderWidth: 0 } },
            legend: { itemStyle: { color: text } },
            credits: { enabled: false },
            series: cat.series
        };

        const lowOptions = {
            chart: { type: 'bar', backgroundColor: 'transparent', height: 300 },
            title: { text: 'Xếp hạn tồn kho (top 10)', style: { color: text, fontSize: '14px' } },
            xAxis: { categories: low.categories, labels: { style: { color: text } }, lineColor: grid, tickColor: grid },
            yAxis: { min: 0, title: { text: 'Vật tư', style: { color: text } }, labels: { style: { color: text } }, gridLineColor: grid },
            tooltip: { shared: true, backgroundColor: '#0b1120', borderColor: grid, style: { color: text } },
            legend: { itemStyle: { color: text } },
            credits: { enabled: false },
            series: low.series
        };

        return (
            <div className="imx-container">
                <header className="imx-header">
                    <div>
                        <h1 className="imx-title">Dashboard</h1>
                        <p className="imx-subtitle">Xem tất cả vật tư, mượn vật tư, lên lịch điều chỉnh và bảo dưỡng thiết bị.</p>
                    </div>
                    <nav className="imx-actions">
                        {signedIn && <Link className="imx-btn imx-btn--primary" to="/add-item">+ Nhập thêm vật tư</Link>}
                        {signedIn && <Link className="imx-btn" to="/borrow">Mượn / Trả</Link>}
                        <Link className="imx-btn" to="/labs">Tất cả các Phòng thực hành</Link>
                        <Link className="imx-btn" to="/">Trang chủ</Link>
                        {!signedIn && <Link className="imx-btn" to="/login">Đăng nhập</Link>}
                        {signedIn && <button className="imx-btn" onClick={() => { localStorage.removeItem('imx_session'); window.location.reload(); }}>Logout</button>}
                        <ThemeToggle />
                    </nav>
                </header>

                {/* Status / errors */}
                {loading && <div className="imx-card"><div className="imx-subtitle">Đang tải dữ liệu…</div></div>}
                {error && <div className="imx-alert imx-alert--error">Lỗi: {error}</div>}

                {/* KPIs */}
                {!loading && !error && (
                    <section className="imx-grid imx-grid--kpi">
                        <div className="imx-card imx-kpi"><div className="imx-kpi__label">Tổng số vật tư</div><div className="imx-kpi__value">{total}</div></div>
                        <div className="imx-card imx-kpi"><div className="imx-kpi__label">Đang sẵn sàng</div><div className="imx-kpi__value">{available}</div></div>
                        <div className="imx-card imx-kpi"><div className="imx-kpi__label">Đã mượn</div><div className="imx-kpi__value">{borrowed}</div></div>
                        <div className="imx-card imx-kpi"><div className="imx-kpi__label">Quá hạn trả</div><div className="imx-kpi__value imx-danger">{overdue}</div></div>
                        <div className="imx-card imx-kpi" style={{gridColumn: 'span 2'}}>
                            <div className="imx-kpi__label">Độ tận dụng</div>
                            <div className="imx-kpi__value">{utilization}%</div>
                            <div aria-hidden style={{marginTop: 8, height: 8, background: '#1b2340', borderRadius: 6, overflow: 'hidden'}}>
                                <div style={{width: `${utilization}%`, height: '100%', background: 'linear-gradient(90deg,#5d88ff,#2e63ff)'}} />
                            </div>
                        </div>
                        <div className="imx-card imx-kpi"><div className="imx-kpi__label">Cần sửa chửa</div><div className="imx-kpi__value">{maintenanceOpen}</div></div>
                    </section>
                )}

                {/* Charts */}
                {!loading && !error && (
                    <>
                        <section className="imx-grid imx-grid--two" style={{marginTop: 14}}>
                            <div className="imx-card">
                                <HighchartsReact highcharts={Highcharts} options={utilOptions} />
                            </div>
                            <div className="imx-card">
                                <HighchartsReact highcharts={Highcharts} options={catOptions} />
                            </div>
                        </section>

                        <section className="imx-card" style={{marginTop: 14}}>
                            <HighchartsReact highcharts={Highcharts} options={lowOptions} />
                        </section>
                    </>
                )}

                {/* Tables */}
                {!loading && !error && (
                    <>
                        {this.renderTableCard(
                            'Danh sách tổng vật tư',
                            data?.log_lab_equipment_master || [],
                            [
                                { label: 'ID', key: 'id' },
                                { label: 'SKU', key: 'sku' },
                                { label: 'Tên vật tư', key: 'name' },
                                { label: 'Phân loại', key: 'category' },
                                { label: 'Vị trí', key: 'location' },
                                { label: 'Số lượng', key: 'quantity' },
                                { label: 'Sẵn sàng', key: 'available_quantity' },
                                { label: 'Trạng thái', key: 'status' },
                                { label: 'Ngày mua', key: 'date_purchased', type: 'date' },
                            ],
                            'equipment'   // <-- enables /dashboard#equipment jump
                        )}

                        {this.renderTableCard(
                            'Lịch sử mượn',
                            borrowRows,
                            [
                                { label: 'ID', key: 'id' },
                                { label: 'Vật tư', key: 'item_name' },
                                { label: 'SKU', key: 'item_sku' },
                                { label: 'ID người dùng', key: 'user_id' },
                                { label: 'Số lượng', key: 'quantity' },
                                { label: 'Ngày mượn', key: 'borrow_date', type: 'date' },
                                { label: 'Hạn trả', key: 'return_date', type: 'date' },
                                { label: 'Ngày thực trả', key: 'actual_return_date', type: 'date' },
                                { label: 'Trạng thái', key: 'status' },
                                { label: 'Trễ hạn', key: 'overdue_flag' },
                            ]
                        )}

                        {this.renderTableCard(
                            'Lịch sử tinh chỉnh',
                            (data?.log_lab_calibration_logs || []).map(c => ({
                                ...c,
                                item_name: equipById.get(c.item_id)?.name || '-',
                                item_sku: equipById.get(c.item_id)?.sku || '-',
                            })),
                            [
                                { label: 'ID', key: 'id' },
                                { label: 'Vật tư', key: 'item_name' },
                                { label: 'SKU', key: 'item_sku' },
                                { label: 'Cân chỉnh bởi', key: 'calibrated_by' },
                                { label: 'Ngày', key: 'date', type: 'date' },
                                { label: 'Ngày hoàn trả', key: 'next_due_date', type: 'date' },
                                { label: 'File điều chỉnh', key: 'cert_file' },
                            ]
                        )}

                        {this.renderTableCard(
                            'Lịch sử sửa chửa',
                            maintRows,
                            [
                                { label: 'ID', key: 'id' },
                                { label: 'Vật tư', key: 'item_name' },
                                { label: 'SKU', key: 'item_sku' },
                                { label: 'Người báo cáo', key: 'reported_by' },
                                { label: 'Ngày báo cáo', key: 'date_reported', type: 'date' },
                                { label: 'Ngày sửa sử', key: 'date_fixed', type: 'date' },
                                { label: 'Trạng thái', key: 'state' },
                                { label: 'Hoạt động thực hiện', key: 'action_taken' },
                            ]
                        )}

                        {this.renderTableCard(
                            'Lịch sử hoạt động',
                            data?.log_lab_activity_logs || [],
                            [
                                { label: 'ID', key: 'id' },
                                { label: 'ID người dùng', key: 'user_id' },
                                { label: 'Hành động', key: 'action' },
                                { label: 'Vào lúc', key: 'timestamp', type: 'date' },
                            ]
                        )}

                        {this.renderTableCard(
                            'Tồn kho',
                            data?.log_lab_storage || [],
                            [
                                { label: 'ID', key: 'id' },
                                { label: 'Mã tồn', key: 'code_name' },
                                { label: 'Tên vật tư', key: 'name' },
                                { label: 'Thông tin', key: 'info' },
                            ]
                        )}
                    </>
                )}
            </div>
        );
    }
}
