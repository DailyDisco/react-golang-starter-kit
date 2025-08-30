import { createFileRoute } from '@tanstack/react-router';
import { useState } from 'react';
import { Button } from '../../components/ui/button';
import { Loader2, AlertCircle } from 'lucide-react';

export const Route = createFileRoute('/(public)/blog')({
  component: BlogPage,
  // Add error boundary
  errorComponent: ({ error }: { error: Error }) => (
    <div className='max-w-2xl mx-auto py-12 px-4 text-center'>
      <AlertCircle className='w-12 h-12 text-destructive mx-auto mb-4' />
      <h1 className='text-2xl font-bold mb-2'>Something went wrong</h1>
      <p className='text-muted-foreground mb-4'>{error.message}</p>
      <Button onClick={() => window.location.reload()}>Try Again</Button>
    </div>
  ),
  // Add loading component
  pendingComponent: () => (
    <div className='max-w-4xl mx-auto py-12 px-4 text-center'>
      <Loader2 className='w-8 h-8 animate-spin mx-auto mb-4' />
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
    <div className='max-w-4xl mx-auto py-8 px-4'>
      <div className='mb-8'>
        <h1 className='text-3xl font-bold mb-2'>Blog</h1>
        <p className='text-muted-foreground'>
          Latest articles and tutorials about web development.
        </p>
      </div>

      <div className='grid grid-cols-1 md:grid-cols-2 gap-6'>
        {data.posts.map(post => (
          <article
            key={post.id}
            className='bg-card p-6 rounded-lg border hover:shadow-md transition-shadow cursor-pointer'
            onClick={() =>
              setSelectedPost(selectedPost === post.id ? null : post.id)
            }
          >
            <h2 className='text-xl font-semibold mb-2 hover:text-primary'>
              {post.title}
            </h2>
            <p className='text-muted-foreground mb-4'>{post.excerpt}</p>
            <div className='flex justify-between items-center text-sm text-muted-foreground'>
              <span>By {post.author}</span>
              <span>{new Date(post.date).toLocaleDateString()}</span>
            </div>

            {selectedPost === post.id && (
              <div className='mt-4 pt-4 border-t'>
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
