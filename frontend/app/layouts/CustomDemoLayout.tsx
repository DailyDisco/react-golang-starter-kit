import { Outlet } from '@tanstack/react-router';
import { Navbar } from '@/layouts';
import { Footer } from '@/layouts';

export default function Layout() {
  return (
    <div className='bg-background flex min-h-screen flex-col'>
      <Navbar />
      {/* Custom Layout Header */}
      <div className='bg-card text-card-foreground border-b p-4 text-center'>
        <div className='text-lg font-medium'>Custom Demo Layout</div>
        <div className='text-muted-foreground mt-1 text-sm'>
          This page uses a different layout structure, for example you can
          scroll down and see the footer is not present.
        </div>
      </div>

      <main className='flex-1 p-6'>
        <div className='mx-auto max-w-4xl'>
          <div className='bg-card rounded-lg border p-8 shadow-sm'>
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}
