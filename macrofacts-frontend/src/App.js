import './App.css';
import { useEffect, useState } from 'react';

function App() {
    const [apiMessage, setApiMessage] = useState('');
    const [me, setMe] = useState(null);

    const [mode, setMode] = useState('login'); // or 'register'
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');

    const token = localStorage.getItem('token');

    useEffect(() => {
        fetch('/api/health')
            .then(r => r.json())
            .then(data => setApiMessage(data.message))
            .catch(console.error);
    }, []);

    useEffect(() => {
        if (!token) {
            setMe(null);
            return;
        }
        fetch('/api/me', {
            headers: { Authorization: `Bearer ${token}` }
        })
            .then(r => {
                if (!r.ok) throw new Error('not logged in');
                return r.json();
            })
            .then(setMe)
            .catch(() => setMe(null));
    }, [token]);

    async function submit(e) {
        e.preventDefault();
        const url = mode === 'register' ? '/api/auth/register' : '/api/auth/login';

        const res = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        const data = await res.json().catch(() => ({}));

        if (!res.ok) {
            const msg = (data && data.error && data.error.message) || data.error || 'Request failed';
            alert(msg);
            return;
        }

        if (mode === 'register') {
            alert('Registered. Now login.');
            setMode('login');
            return;
        }

        localStorage.setItem('token', data.token);
        // trigger me refresh
        window.location.reload();
    }

    function logout() {
        localStorage.removeItem('token');
        setMe(null);
        window.location.reload();
    }

    return (
        <div className="App" style={{ padding: 24, maxWidth: 480, margin: '0 auto' }}>
            <h1>Macrofacts</h1>

            <p><b>API:</b> {apiMessage || '...'}</p>

            {me ? (
                <div>
                    <p>Logged in as <b>{me.username}</b></p>
                    <button onClick={logout}>Logout</button>
                </div>
            ) : (
                <form onSubmit={submit}>
                    <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
                        <button type="button" onClick={() => setMode('login')} disabled={mode === 'login'}>Login</button>
                        <button type="button" onClick={() => setMode('register')} disabled={mode === 'register'}>Register</button>
                    </div>

                    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                        <input
                            placeholder="username"
                            value={username}
                            onChange={e => setUsername(e.target.value)}
                            autoComplete="username"
                        />
                        <input
                            placeholder="password"
                            type="password"
                            value={password}
                            onChange={e => setPassword(e.target.value)}
                            autoComplete={mode === 'register' ? 'new-password' : 'current-password'}
                        />
                        <button type="submit">{mode === 'register' ? 'Create account' : 'Login'}</button>
                    </div>
                </form>
            )}
        </div>
    );
}

export default App;
