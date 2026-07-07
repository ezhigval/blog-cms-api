const API = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090';

export type Post = {
  id: number;
  title: string;
  slug: string;
  excerpt?: string;
  body: string;
  status: string;
  published_at?: string;
  category?: { name: string; slug: string };
  tags?: { name: string; slug: string }[];
};

export type PostList = {
  items: Post[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
};

export async function fetchPosts(page = 1, q = ''): Promise<PostList> {
  const params = new URLSearchParams({ page: String(page), per_page: '10' });
  if (q) params.set('q', q);
  const res = await fetch(`${API}/api/v1/posts?${params}`, { next: { revalidate: 60 } });
  if (!res.ok) throw new Error('failed to fetch posts');
  return res.json();
}

export async function fetchPost(slug: string): Promise<Post> {
  const res = await fetch(`${API}/api/v1/posts/slug/${slug}`, { next: { revalidate: 60 } });
  if (!res.ok) throw new Error('post not found');
  return res.json();
}

export async function login(email: string, password: string): Promise<{ access_token: string }> {
  const res = await fetch(`${API}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) throw new Error('login failed');
  return res.json();
}

export async function adminPosts(token: string): Promise<PostList> {
  const res = await fetch(`${API}/api/v1/admin/posts`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  if (!res.ok) throw new Error('admin fetch failed');
  return res.json();
}
