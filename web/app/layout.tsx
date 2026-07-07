export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body style={{ fontFamily: 'system-ui, sans-serif', margin: 0, background: '#0f1117', color: '#e6edf3' }}>
        <header style={{ padding: '1rem 2rem', borderBottom: '1px solid #30363d' }}>
          <nav style={{ display: 'flex', gap: '1.5rem', maxWidth: 900, margin: '0 auto' }}>
            <a href="/" style={{ color: '#58a6ff', textDecoration: 'none', fontWeight: 600 }}>Blog</a>
            <a href="/admin" style={{ color: '#8b949e', textDecoration: 'none' }}>Admin</a>
          </nav>
        </header>
        <main style={{ maxWidth: 900, margin: '0 auto', padding: '2rem' }}>{children}</main>
      </body>
    </html>
  );
}
