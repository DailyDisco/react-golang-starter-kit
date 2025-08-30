import { createFileRoute } from '@tanstack/react-router';
import { useState } from 'react';
import { Button } from '../../components/ui/button';
import { Input } from '../../components/ui/input';
import { Search as SearchIcon } from 'lucide-react';

export const Route = createFileRoute('/(public)/search')({
  component: SearchPage,
  // Validate search parameters
  validateSearch: search => ({
    q: (search.q as string) || '',
    type: (search.type as 'all' | 'users' | 'posts') || 'all',
    page: Number(search.page) || 1,
  }),
});

function SearchPage() {
  const { q, type, page } = Route.useSearch();
  const navigate = Route.useNavigate();
  const [searchQuery, setSearchQuery] = useState(q);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    navigate({
      search: { q: searchQuery, type, page: 1 },
      replace: true,
    });
  };

  const updateSearchType = (newType: 'all' | 'users' | 'posts') => {
    navigate({
      search: { q: searchQuery, type: newType, page },
      replace: true,
    });
  };

  return (
    <div className='mx-auto max-w-4xl px-4 py-8'>
      <h1 className='mb-8 text-3xl font-bold'>Search</h1>

      {/* Search Form */}
      <form onSubmit={handleSearch} className='mb-8'>
        <div className='flex gap-4'>
          <div className='flex-1'>
            <Input
              type='text'
              placeholder='Search for users, posts, or content...'
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              className='w-full'
            />
          </div>
          <Button type='submit'>
            <SearchIcon className='mr-2 h-4 w-4' />
            Search
          </Button>
        </div>
      </form>

      {/* Search Filters */}
      <div className='mb-6'>
        <div className='flex gap-2'>
          <Button
            variant={type === 'all' ? 'default' : 'outline'}
            size='sm'
            onClick={() => updateSearchType('all')}
          >
            All
          </Button>
          <Button
            variant={type === 'users' ? 'default' : 'outline'}
            size='sm'
            onClick={() => updateSearchType('users')}
          >
            Users
          </Button>
          <Button
            variant={type === 'posts' ? 'default' : 'outline'}
            size='sm'
            onClick={() => updateSearchType('posts')}
          >
            Posts
          </Button>
        </div>
      </div>

      {/* Search Results */}
      <div className='bg-card rounded-lg border p-6'>
        <h2 className='mb-4 text-lg font-semibold'>Search Results</h2>

        {q ? (
          <div className='space-y-4'>
            <p className='text-muted-foreground'>
              Searching for "<strong>{q}</strong>" in <strong>{type}</strong>{' '}
              (Page {page})
            </p>

            {/* Mock search results */}
            <div className='space-y-3'>
              <div className='rounded border p-4'>
                <h3 className='font-medium'>Sample Result 1</h3>
                <p className='text-muted-foreground text-sm'>
                  This is a sample search result...
                </p>
              </div>
              <div className='rounded border p-4'>
                <h3 className='font-medium'>Sample Result 2</h3>
                <p className='text-muted-foreground text-sm'>
                  Another sample search result...
                </p>
              </div>
            </div>
          </div>
        ) : (
          <p className='text-muted-foreground'>
            Enter a search query to get started.
          </p>
        )}
      </div>
    </div>
  );
}
