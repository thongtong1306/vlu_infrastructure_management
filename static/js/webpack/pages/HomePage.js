// pages/HomePage.js
import React from 'react';
import { Link } from 'react-router-dom';

export default function HomePage() {
    return (
        <div className="imx-container">
            <header className="imx-header">
                <div>
                    <h1 className="imx-title">Chào mừng</h1>
                    <p className="imx-subtitle">Trang quản lý vật tư phòng thí nghiệm — Bộ môn Logsitics và Quản trị Chuỗi cung ứng Trường Đại học Văn Lang.</p>
                </div>
            </header>

            <div className="imx-grid imx-grid--three">
                <div className="imx-card imx-card--hover">
                    <div className="imx-card__body">
                        <h2 className="imx-card__title">Dashboard</h2>
                        <p className="imx-subtitle">Xem trạng thái, danh sách thiết bị, và các bảng biểu thống kê.</p>
                        <Link className="imx-btn imx-btn--primary" to="/dashboard" style={{marginTop: 12}}>Xem Dashboard →</Link>
                    </div>
                </div>

                <div className="imx-card imx-card--hover">
                    <div className="imx-card__body">
                        <h2 className="imx-card__title">Tất cả các Phòng thí nghiệm</h2>
                        <p className="imx-subtitle">Xem các nội dung, quy định phòng thí nghiệm, giờ mở cửa và liên hệ.</p>
                        <Link className="imx-btn" to="/labs" style={{marginTop: 12}}>Xem tất cả các Phòng thí nghiệm →</Link>
                    </div>
                </div>

                <div className="imx-card imx-card--hover">
                    <div className="imx-card__body">
                        <h2 className="imx-card__title">Tất cả vật tư</h2>
                        <p className="imx-subtitle">Xem toàn bộ vật tư, hình ảnh chi tiết, hướng dẫn sử dụng và lưu ý.</p>
                        <Link className="imx-btn" to="/equipments" style={{marginTop: 12}}>Xem tất cả vật tư →</Link>
                    </div>
                </div>
            </div>
        </div>
    );
}
