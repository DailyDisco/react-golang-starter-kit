import { createFileRoute } from '@tanstack/react-router';
import { AlertCircle, Loader2 } from 'lucide-react';
import { useState } from 'react';

import { Button } from '../../components/ui/button';

export const Route = createFileRoute('/(public)/blog')({
  component: BlogPage,
  // Add error boundary
  errorComponent: ({ error }: { error: Error }) => (
    <div className='mx-auto max-w-2xl px-4 py-12 text-center'>
      <AlertCircle className='text-destructive mx-auto mb-4 h-12 w-12' />
      <h1 className='mb-2 text-2xl font-bold'>Something went wrong</h1>
      <p className='text-muted-foreground mb-4'>{error.message}</p>
      <Button onClick={() => window.location.reload()}>Try Again</Button>
    </div>
  ),
  // Add loading component
  pendingComponent: () => (
    <div className='mx-auto max-w-4xl px-4 py-12 text-center'>
      <Loader2 className='mx-auto mb-4 h-8 w-8 animate-spin' />
      <p className='text-muted-foreground'>Loading blog posts...</p>
    </div>
  ),
  // Add loader with error handling
  loader: async () => {
    // Simulate API call that might fail
    await new Promise(resolve => setTimeout(resolve, 1500));

    // Simulate random error for demo
    if (Math.random() > 0.7) {
      throw new Error('Failed to load blog posts. Please try again.');
    }

    return {
      posts: [
        {
          id: 1,
          title: 'Getting Started with TanStack Router',
          excerpt: 'Learn how to set up routing in your React application...',
          author: 'John Doe',
          date: '2024-01-15',
        },
        {
          id: 2,
          title: 'Advanced Route Patterns',
          excerpt:
            'Explore dynamic routes, search parameters, and nested layouts...',
          author: 'Jane Smith',
          date: '2024-01-10',
        },
        {
          id: 3,
          title: 'Type Safety in React Router',
          excerpt:
            'How TanStack Router provides excellent TypeScript support...',
          author: 'Bob Johnson',
          date: '2024-01-05',
        },
      ],
    };
  },
});

function BlogPage() {
  const data = Route.useLoaderData();
  const [selectedPost, setSelectedPost] = useState<number | null>(null);

  return (
    <div className='mx-auto max-w-4xl px-4 py-8'>
      <div className='mb-8'>
        <h1 className='mb-2 text-3xl font-bold'>Blog</h1>
        <p className='text-muted-foreground'>
          Latest articles and tutorials about web development.
        </p>
      </div>

      <div className='grid grid-cols-1 gap-6 md:grid-cols-2'>
        {data.posts.map(post => (
          <article
            key={post.id}
            className='bg-card cursor-pointer rounded-lg border p-6 transition-shadow hover:shadow-md'
            onClick={() =>
              setSelectedPost(selectedPost === post.id ? null : post.id)
            }
          >
            <h2 className='hover:text-primary mb-2 text-xl font-semibold'>
              {post.title}
            </h2>
            <p className='text-muted-foreground mb-4'>{post.excerpt}</p>
            <div className='text-muted-foreground flex items-center justify-between text-sm'>
              <span>By {post.author}</span>
              <span>{new Date(post.date).toLocaleDateString()}</span>
            </div>

            {selectedPost === post.id && (
              <div className='mt-4 border-t pt-4'>
                <p className='text-sm'>
                  This is the full content of the blog post. In a real
                  application, this would contain the complete article text,
                  images, and formatting.
                </p>
              </div>
            )}
          </article>
        ))}
      </div>

      <div className='mt-8 text-center'>
        <Button variant='outline'>Load More Posts</Button>
      </div>
    </div>
  );
}
