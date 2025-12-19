// pages/Labs/LabMain.js
import React from 'react';
import { Link } from 'react-router-dom';

const MAIN_IMG =
    '/static/img/lab/lab1.jpg'; // electronics lab
const FALLBACK_IMG =
    'https://picsum.photos/1200/800?random=11';

export default function LabMain() {
    return (
        <div className="imx-container">
            <header className="imx-header">
                <div>
                    <h1 className="imx-title">Phòng thí nghiệm 1 – D.1.01 CS2</h1>
                    <p className="imx-subtitle">
                        Phòng thực hành Chế tạo và Gia công Cơ khí
                    </p>
                </div>
                <nav className="imx-actions">
                    <Link className="imx-btn" to="/labs">Tất cả Phòng thực hành</Link>
                    <Link className="imx-btn" to="/">Trang chủ</Link>
                </nav>
            </header>

            {/* Hero split: image (left) + content (right) */}
            <section className="imx-card imx-card--hero" style={{marginBottom: 14}}>
                <div className="imx-card__media--left">
                    <img
                        src={MAIN_IMG}
                        alt="Main lab workspace"
                        onError={(e) => { e.currentTarget.src = FALLBACK_IMG; }}
                        loading="lazy"
                    />
                </div>

                <div className="imx-card__body">
                    <h2 className="imx-card__title" style={{marginBottom: 8}}>Thông tin chung</h2>
                    <p style={{fontSize: 13, lineHeight: 1.5, textAlign: "justify"}}>
                        <p>Phòng Lab Chế tạo và Gia công Cơ khí (Mechanical Fabrication Lab), năm thành lập 2023, lưu lượng phục vụ tối đa 120 sinh viên/học phần. Phòng được trang bị máy CNC, máy cắt laser, máy in 3D, máy khoan, máy hàn và dụng cụ cơ khí, ... giúp sinh viên phát triển kỹ năng chế tạo, tư duy kỹ thuật và ứng dụng vào các dự án sáng tạo. Đây là không gian thực hành dành cho việc thiết kế, gia công và lắp ráp các chi tiết cơ khí phục vụ các mô hình và thiết bị logistics.</p>
                        <p>Phòng thực hành Chế tạo và Gia công Cơ khí hiện đang được sử dụng phục vụ học phần Các mô hình ứng dụng trong Logistics (71SCMN40293), Quản trị chất lượng (71SCMN40023), Kỹ thuật hệ thống (71SCMN40323), Kỹ thuật Logistics (71SCMN40483), Khóa luận tốt nghiệp (71LSCM40326)</p>
                    </p>
                    <ul style={{ marginTop: 0 }}>
                        <li><strong>Vị trí:</strong> Toà nhà D.1.01 Trường Đại học Văn Lang Cơ sở 2</li>
                        <li><strong>Giờ làm việc:</strong> 08:00-17:00 Từ thứ 2 đến thứ 6; 08:00-11:30 Thứ 7; Nghỉ trưa 11:30-13:00. </li>
                        <li><strong>Sức chứa:</strong> 08 nhóm sinh viên; Mỗi nhóm 2-3 sinh viên</li>
                        <li><strong>An toàn:</strong> Theo quy chuẩn Phòng thực hành chung. Xem quy định bên dưới.</li>
                    </ul>
                </div>
            </section>

            <div className="imx-grid imx-grid--two">
                <div className="imx-card">
                    <div className="imx-card__header"><h2 className="imx-card__title">Quy định Phòng thực hành</h2></div>
                    <ol className="imx-list" style={{textAlign: "justify"}}>
                        <li>Yêu cầu chung về an toàn</li>
                        <ul className="imx-list">
                            <li>Chỉ những cá nhân đã được đào tạo hoặc được giám sát trực tiếp bởi người có chuyên môn mới được phép vận hành thiết bị cơ khí.</li>
                            <li>Phải hiểu rõ nguyên lý hoạt động và rủi ro của thiết bị trước khi sử dụng.</li>
                        </ul>
                        <li>Trang bị bảo hộ cá nhân (PPE)</li>
                        <ul className="imx-list">
                            <li>Trang phục bảo hộ phù hợp là bắt buộc khi sử dụng thiết bị cơ khí: Đồng phục kỹ thuật hoặc trang phục gọn gàng, không vướng víu. Đối với nữ, yêu cầu buộc tóc gọn gàng để đảm bảo an toàn.</li>
                            <li>Giày bảo hộ hoặc giày kín mũi, chống trượt. Kính bảo hộ chống va đập. Bịt tai (nếu làm việc với thiết bị gây tiếng ồn lớn).</li>
                        </ul>
                        <li>An toàn trong vận hành thiết bị</li>
                        <ul className="imx-list">
                            <li>Trước khi vận hành: Kiểm tra toàn bộ thiết bị (điện áp, dây nối, che chắn, trục quay, bộ phận an toàn), đảm bảo khu vực làm việc thông thoáng và không có vật cản/hóa chất dễ cháy.</li>
                            <li>Trong khi vận hành: Tuyệt đối không rời khỏi thiết bị nếu không có chế độ tự động và bảo vệ an toàn phù hợp; dừng máy ngay lập tức và báo cáo khi có bất kỳ hiện tượng bất thường nào (rung lắc, tiếng ồn lạ, mùi khét).</li>
                            <li>Sau khi vận hành: Tắt và ngắt nguồn điện thiết bị; vệ sinh sạch sẽ khu vực làm việc (loại bỏ phoi, bụi, dầu mỡ dư thừa) và ghi chép tình trạng thiết bị nếu được yêu cầu theo dõi.</li>
                        </ul>
                        <li>Bảo trì, kiểm định và sửa chữa</li>
                        <ul className="imx-list">
                            <li>Các thiết bị cơ khí phải được kiểm định theo định kỳ bởi đơn vị có thẩm quyền.</li>
                            <li>Người sử dụng không tự ý sửa chữa khi có hỏng hóc. Báo cáo ngay cho cán bộ kỹ thuật hoặc quản lý phòng.</li>
                            <li>Mọi hoạt động bảo trì phải có nhật ký theo dõi.</li>
                        </ul>
                        <li>Xử lý sự cố và sơ cứu</li>
                        <ul className="imx-list">
                            <li>Khi xảy ra sự cố, cần ưu tiên ngắt nguồn điện và sử dụng bình chữa cháy CO₂ hoặc bột khô nếu có cháy thiết bị điện, đồng thời kêu gọi hỗ trợ và cấp cứu.</li>
                            <li>Về sơ cứu, với vết thương đứt tay/chảy máu, hãy rửa sạch, cầm máu và đến y tế. Đối với bỏng, làm mát bằng nước sạch (nếu an toàn) và không tự ý bôi thuốc. Nếu bị kẹt tay, giữ nguyên hiện trường và chờ cứu hộ chuyên nghiệp.</li>
                        </ul>
                        <li>Quản lý và giám sát thiết bị</li>
                        <ul className="imx-list">
                            <li>Mỗi thiết bị cơ khí cần có: Hồ sơ quản lý riêng (bao gồm thông tin kỹ thuật, hướng dẫn sử dụng, lịch bảo trì).</li>
                            <li>Biển cảnh báo nguy hiểm rõ ràng, có hướng dẫn an toàn ngắn gọn.</li>
                            <li>Người phụ trách thiết bị chịu trách nhiệm theo dõi tình trạng và giám sát việc sử dụng.</li>
                        </ul>
                        <li>Trách nhiệm và xử lý vi phạm</li>
                        <ul className="imx-list">
                            <li>Người sử dụng thiết bị chịu trách nhiệm hoàn toàn nếu không tuân thủ quy định an toàn và gây ra sự cố. Vi phạm quy định có thể dẫn đến:</li>
                            <ul className="imx-list">
                                <li>Đình chỉ quyền sử dụng thiết bị.</li>
                                <li>Bồi thường thiệt hại nếu có hư hỏng thiết bị hoặc ảnh hưởng đến người khác.</li>
                                <li>Xử lý kỷ luật theo quy định của cơ sở đào tạo hoặc tổ chức quản lý.</li>
                            </ul>
                        </ul>
                    </ol>
                </div>

                <div className="imx-card">
                    <div className="imx-card__header"><h2 className="imx-card__title">Một số thiết bị chính</h2></div>
                    <ul className="imx-list">
                        <li>Phục vụ đo đạc: Máy hiện sóng (oscilloscopes), máy cấp nguồn, đồng hồ đo điện, thước kẹp, thước kéo, ...</li>
                        <li>Phục vụ lắp ráp:  Máy khoan bàn, máy khoan cầm tay, trạm hàn, cờ lê, kiềm, búa, máy mài, máy CNC, ...</li>
                        <li>Nguyên vật liệu - vật tư: Cáp tín hiệu, nhôm lá, sắt lỗ, nhôm định hình, ốc vit, băng keo, keo expoxy,...</li>
                    </ul>
                    <p className="imx-subtitle">
                        Cần xem chi tiết danh sách vật tư? Xem <Link to="/dashboard" className="imx-link">Dashboard</Link>.
                    </p>
                </div>

                <div className="imx-card">
                    <div className="imx-card__header"><h2 className="imx-card__title">Đặt phòng sử dụng & Liên hệ</h2></div>
                    <ul className="imx-list">
                        <li>Đăng ký mượn vật tư qua <Link to="/borrow" className="imx-link">Mượn/Trả</Link>.</li>
                        <li>Quản lý phòng: bm.logistics@vlu.edu.vn</li>
                        <li>Số liên hệ khẩn cấp: +84 981392300 (Bảo - Kỹ thuật viên)</li>
                    </ul>
                </div>
            </div>
        </div>
    );
}
