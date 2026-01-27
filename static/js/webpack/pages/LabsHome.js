import React from 'react';
import { Link } from 'react-router-dom';

export default function LabsHome() {
    const cards = [
        {
            to: '/labs/main',
            title: 'Phòng thực hành 1 – D.1.01 CS2',
            desc: 'Phòng thực hành Chế tạo và Gia công Cơ khí',
            img: '/static/img/lab/lab1.jpg', // electronics bench
        },
        {
            to: '/labs/lab-2',
            title: 'Phòng thực hành 2 – D.1.04 CS2',
            desc: 'Phòng thực hành Hệ thống Tự động và Robot Logistics',
            img: '/static/img/lab/lab2.jpg', // tools / workshop
        },
        {
            to: '/labs/lab-3',
            title: 'Phòng thực hành 3 – D.1.05 CS2',
            desc: 'Phòng thực hành Lập trình & Tối ưu hóa Hệ thống Logistics',
            img: '/static/img/lab/lab3.jpg', // racks / instruments
        },
    ];

    return (
        <div className="imx-container">
            <header className="imx-header">
                <div>
                    <h1 className="imx-title">Tất cả các Phòng thực hành</h1>
                    <p className="imx-subtitle">
                        Select a lab to read the regulations, hours, and equipment notes.
                    </p>
                </div>
                <nav className="imx-actions">
                    <Link className="imx-btn" to="/dashboard">Dashboard</Link>
                    {/*{!signedIn && <Link className="imx-btn" to="/login">Đăng nhập</Link>}*/}
                    <Link className="imx-btn" to="/">Trang chủ</Link>
                </nav>
            </header>

            <div className="imx-grid imx-grid--three">
                {cards.map((c) => (
                    <Link
                        key={c.to}
                        to={c.to}
                        className="imx-card imx-card--hover"
                        style={{ textDecoration: 'none' }}
                    >
                        <div className="imx-card__media">
                            <img src={c.img} alt={c.title} />
                        </div>
                        <div className="imx-card__header">
                            <h3 className="imx-card__title">{c.title}</h3>
                        </div>
                        <p>{c.desc}</p>
                        <span className="imx-link" style={{ marginTop: 8 }}>Open →</span>
                    </Link>
                ))}
            </div>
        </div>
    );
}
