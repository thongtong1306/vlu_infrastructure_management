import React, { Component } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { connect } from 'react-redux';
import { mapStateToProps, mapDispatchToProps } from './redux-mapping';

class LoginPage extends Component {
    state = {
        email: '',
        password: '',
        loading: false,
        error: null,
    };

    componentDidMount() {
        document.title = 'Infrastructure Management | Login';
    }

    handleSubmit = async (e) => {
        e.preventDefault();
        const { email, password } = this.state;
        if (!email || !password) {
            this.setState({ error: 'Please enter email and password.' });
            return;
        }

        this.setState({ loading: true, error: null });
        try {
            const res = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ identifier: email, password: password }),
            });
            if (!res.ok) throw new Error((await res.text()) || 'Login failed');
            const data = await res.json(); // { token, exp, user }

            const session = { token: data.token, exp: data.exp, user: data.user };
            await this.props.onReduxUpdate('session', session);
            try { localStorage.setItem('imx_session', JSON.stringify(session)); } catch {}

            this.props.navigate('/dashboard'); // redirect to Home
        } catch (err) {
            this.setState({ error: err.message || 'Login failed' });
        } finally {
            this.setState({ loading: false });
        }
    };

    render() {
        const { email, password, loading, error } = this.state;

        return (
            <div className="imx-hero">
                <header className="imx-topbar">
                    <div className="imx-topbar__brand">VLU Infrastructure Account Log In</div>
                    {/*<nav className="imx-topbar__nav">*/}
                    {/*    <a href="#" className="imx-link" onClick={(e) => e.preventDefault()}>Register</a>*/}
                    {/*    <Link to="/login" className="imx-link">Login</Link>*/}
                    {/*</nav>*/}
                </header>

                <main className="imx-auth">
                    <div className="imx-card imx-auth__card">
                        <h1 className="imx-auth__title">Welcome to VLU Infrastructure</h1>
                        <p className="imx-auth__subtitle">
                            Don’t have an account yet?{' '}
                            <a href="#" className="imx-link" onClick={(e) => e.preventDefault()}>Sign up.</a>
                        </p>

                        {error && <div className="imx-alert imx-alert--error" role="alert">{error}</div>}

                        <form className="imx-form" onSubmit={this.handleSubmit}>
                            <label className="imx-label" htmlFor="email">Email</label>
                            <input
                                id="email"
                                className="imx-input"
                                type="email"
                                placeholder="you@company.com"
                                value={email}
                                onChange={(e) => this.setState({ email: e.target.value })}
                            />

                            <label className="imx-label" htmlFor="password">Password</label>
                            <input
                                id="password"
                                className="imx-input"
                                type="password"
                                placeholder="••••••••"
                                value={password}
                                onChange={(e) => this.setState({ password: e.target.value })}
                            />

                            <button className="imx-btn imx-btn--outline" type="submit" disabled={loading}>
                                {loading ? 'Signing in…' : 'Login'}
                            </button>
                        </form>

                        <p className="imx-auth__subtitle">
                            Don’t have an account yet? <Link to="/register" className="imx-link">Sign up.</Link>
                        </p>

                        <p className="imx-auth__help">
                            Forgot your password? <a href="#" className="imx-link">Reset here.</a>
                        </p>
                    </div>
                </main>
            </div>
        );
    }
}

// wrapper to use navigate() inside a class
const withNavigation = (Comp) => (props) => {
    const navigate = useNavigate();
    return <Comp {...props} navigate={navigate} />;
};

export default withNavigation(connect(mapStateToProps, mapDispatchToProps)(LoginPage));
