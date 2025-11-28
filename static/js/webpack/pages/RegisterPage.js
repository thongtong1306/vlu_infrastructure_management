import React, { Component } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { connect } from 'react-redux';
import { mapStateToProps, mapDispatchToProps } from './redux-mapping';

class RegisterPage extends Component {
    state = {
        fullName: '',
        username: '',
        email: '',
        password: '',
        confirm: '',
        accepted: true,   // set default as you like
        loading: false,
        error: null,
    };

    componentDidMount() {
        document.title = 'Infrastructure Management | Register';
    }

    validate = () => {
        const { fullName, username, email, password, confirm, accepted } = this.state;
        if (!fullName || !username || !email || !password) return 'Please fill all fields.';
        if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email) && !email.includes('@'))
            return 'Please enter a valid email (admin@local is ok).';
        if (password.length < 8) return 'Password must be at least 8 characters.';
        if (password !== confirm) return 'Passwords do not match.';
        if (!accepted) return 'Please accept the terms.';
        return null;
    };

    handleSubmit = async (e) => {
        e.preventDefault();
        const error = this.validate();
        if (error) return this.setState({ error });

        const { fullName, username, email, password } = this.state;
        this.setState({ loading: true, error: null });

        try {
            const res = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    full_name: fullName,
                    username,
                    email,
                    password,
                }),
            });

            if (!res.ok) throw new Error((await res.text()) || 'Registration failed');

            // If API returns a token/session, store it; otherwise route to login
            let data = {};
            try { data = await res.json(); } catch {}
            if (data && data.token) {
                const session = { token: data.token, exp: data.exp, user: data.user };
                await this.props.onReduxUpdate('session', session);
                try { localStorage.setItem('imx_session', JSON.stringify(session)); } catch {}
                this.props.navigate('/dashboard'); // signed in
            } else {
                this.props.navigate('/login');
            }
        } catch (err) {
            this.setState({ error: err.message || 'Registration failed' });
        } finally {
            this.setState({ loading: false });
        }
    };

    render() {
        const { fullName, username, email, password, confirm, accepted, loading, error } = this.state;

        return (
            <div className="imx-hero">
                <header className="imx-topbar">
                    <div className="imx-topbar__brand">VLU Infrastructure Account Register</div>
                    <nav className="imx-topbar__nav">
                        <Link to="/" className="imx-link">Trang chủ</Link>
                        <Link to="/login" className="imx-link">Đăng nhập</Link>
                    </nav>
                </header>

                <main className="imx-auth">
                    <div className="imx-card imx-auth__card">
                        <h1 className="imx-auth__title">Create your account</h1>
                        <p className="imx-auth__subtitle">It’s quick and free.</p>

                        {error && <div className="imx-alert imx-alert--error" role="alert">{error}</div>}

                        <form className="imx-form" onSubmit={this.handleSubmit}>
                            <label className="imx-label" htmlFor="fullName">Full name</label>
                            <input id="fullName" className="imx-input" value={fullName}
                                   onChange={(e)=>this.setState({fullName:e.target.value})} placeholder="Jane Doe" />

                            <label className="imx-label" htmlFor="username">Username</label>
                            <input id="username" className="imx-input" value={username}
                                   onChange={(e)=>this.setState({username:e.target.value})} placeholder="jane" />

                            <label className="imx-label" htmlFor="email">Email</label>
                            <input id="email" className="imx-input" type="email" value={email}
                                   onChange={(e)=>this.setState({email:e.target.value})} placeholder="you@company.com" />

                            <label className="imx-label" htmlFor="password">Password</label>
                            <input id="password" className="imx-input" type="password" value={password}
                                   onChange={(e)=>this.setState({password:e.target.value})} placeholder="••••••••" />

                            <label className="imx-label" htmlFor="confirm">Confirm password</label>
                            <input id="confirm" className="imx-input" type="password" value={confirm}
                                   onChange={(e)=>this.setState({confirm:e.target.value})} placeholder="••••••••" />

                            <label className="imx-checkbox">
                                <input type="checkbox" checked={accepted}
                                       onChange={(e)=>this.setState({accepted:e.target.checked})} />
                                <span>I agree to the terms</span>
                            </label>

                            <button className="imx-btn imx-btn--outline" type="submit" disabled={loading}>
                                {loading ? 'Creating…' : 'Create account'}
                            </button>
                        </form>

                        <p className="imx-auth__help">
                            Already have an account? <Link to="/login" className="imx-link">Sign in</Link>
                        </p>
                    </div>
                </main>
            </div>
        );
    }
}

// wrapper to use navigate() in a class component
const withNavigation = (Comp) => (props) => {
    const navigate = useNavigate();
    return <Comp {...props} navigate={navigate} />;
};

export default withNavigation(connect(mapStateToProps, mapDispatchToProps)(RegisterPage));
