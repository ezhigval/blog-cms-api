import Link from 'next/link';
import { fetchPosts } from '@/lib/api';

export const revalidate = 60;

export default async function HomePage() {
  let data;
  try {
    data = await fetchPosts();
  } catch {
    return <p>API unavailable — start Go server on :8090</p>;
  }

  return (
    <div>
      <h1 style={{ marginBottom: '0.25rem' }}>Blog</h1>
      <p style={{ color: '#8b949e', marginTop: 0 }}>ISR revalidate 60s · Go API + Next.js</p>
      <ul style={{ listStyle: 'none', padding: 0 }}>
        {data.items.map((post) => (
          <li key={post.id} style={{ marginBottom: '1.5rem', paddingBottom: '1.5rem', borderBottom: '1px solid #21262d' }}>
            <Link href={`/posts/${post.slug}`} style={{ color: '#58a6ff', fontSize: '1.25rem', textDecoration: 'none' }}>
              {post.title}
            </Link>
            {post.excerpt && <p style={{ color: '#8b949e' }}>{post.excerpt}</p>}
            {post.category && <span style={{ fontSize: '0.85rem', color: '#6e7681' }}>{post.category.name}</span>}
          </li>
        ))}
      </ul>
      {data.items.length === 0 && <p>No published posts yet. Use admin to create one.</p>}
    </div>
  );
}
