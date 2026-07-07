import { notFound } from 'next/navigation';
import ReactMarkdown from 'react-markdown';
import { fetchPost } from '@/lib/api';

export const revalidate = 60;

type Props = { params: Promise<{ slug: string }> };

export default async function PostPage({ params }: Props) {
  const { slug } = await params;
  let post;
  try {
    post = await fetchPost(slug);
  } catch {
    notFound();
  }

  return (
    <article>
      <h1>{post.title}</h1>
      {post.published_at && (
        <time style={{ color: '#8b949e', fontSize: '0.9rem' }}>
          {new Date(post.published_at).toLocaleDateString()}
        </time>
      )}
      <div style={{ marginTop: '2rem', lineHeight: 1.7 }}>
        <ReactMarkdown>{post.body}</ReactMarkdown>
      </div>
    </article>
  );
}
