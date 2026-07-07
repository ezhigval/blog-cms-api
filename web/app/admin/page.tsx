'use client';

import { useEffect, useState } from 'react';
import { adminPosts, login, type Post } from '@/lib/api';

export default function AdminPage() {
  const [token, setToken] = useState<string | null>(null);
  const [email, setEmail] = useState('admin@example.com');
  const [password, setPassword] = useState('');
  const [posts, setPosts] = useState<Post[]>([]);
  const [error, setError] = useState('');

  useEffect(() => {
    const t = localStorage.getItem('cms_token');
    if (t) setToken(t);
  }, []);

  useEffect(() => {
    if (!token) return;
    adminPosts(token)
      .then((d) => setPosts(d.items))
      .catch(() => setError('Failed to load posts'));
  }, [token]);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    try {
      const { access_token } = await login(email, password);
      localStorage.setItem('cms_token', access_token);
      setToken(access_token);
    } catch {
      setError('Login failed');
    }
  }

  if (!token) {
    return (
      <form onSubmit={handleLogin} style={{ maxWidth: 360 }}>
        <h1>Admin login</h1>
        <p style={{ color: '#8b949e' }}>First registered user becomes admin.</p>
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="email"
          style={inputStyle}
        />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="password"
          style={inputStyle}
        />
        <button type="submit" style={btnStyle}>Sign in</button>
        {error && <p style={{ color: '#f85149' }}>{error}</p>}
      </form>
    );
  }

  return (
    <div>
      <h1>Posts</h1>
      <p style={{ color: '#8b949e' }}>Create posts via API — curl example in README.</p>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr style={{ textAlign: 'left', color: '#8b949e' }}>
            <th>Title</th>
            <th>Status</th>
            <th>Slug</th>
          </tr>
        </thead>
        <tbody>
          {posts.map((p) => (
            <tr key={p.id} style={{ borderTop: '1px solid #21262d' }}>
              <td style={{ padding: '0.75rem 0' }}>{p.title}</td>
              <td>{p.status}</td>
              <td><code>{p.slug}</code></td>
            </tr>
          ))}
        </tbody>
      </table>
      {posts.length === 0 && <p>No posts yet.</p>}
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  display: 'block',
  width: '100%',
  marginBottom: '0.75rem',
  padding: '0.5rem',
  background: '#161b22',
  border: '1px solid #30363d',
  borderRadius: 6,
  color: '#e6edf3',
};

const btnStyle: React.CSSProperties = {
  padding: '0.5rem 1rem',
  background: '#238636',
  border: 'none',
  borderRadius: 6,
  color: '#fff',
  cursor: 'pointer',
};
