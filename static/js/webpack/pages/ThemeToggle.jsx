import React from 'react';

export default function ThemeToggle() {
    const [theme, setTheme] = React.useState(() =>
        document.documentElement.getAttribute('data-theme') || 'dark'
    );

    function toggle() {
        const next = theme === 'light' ? 'dark' : 'light';
        document.documentElement.setAttribute('data-theme', next);
        try { localStorage.setItem('theme', next); } catch {}
        setTheme(next);
    }

    return (
        <button type="button" className="imx-btn" onClick={toggle} title="Toggle theme">
            {theme === 'light' ? '🌙 Dark' : '☀️ Light'}
        </button>
    );
}
