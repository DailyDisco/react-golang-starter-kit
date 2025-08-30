import { Outlet } from '@tanstack/react-router';
import { Navbar } from '@/layouts';
import { Footer } from '@/layouts';

export default function Layout() {
  return (
    <div className='min-h-screen flex flex-col bg-background'>
      <Navbar />
      {/* Custom Layout Header */}
      <div className='bg-card text-card-foreground p-4 text-center border-b'>
        <div className='text-lg font-medium'>Custom Demo Layout</div>
        <div className='text-sm text-muted-foreground mt-1'>
          This page uses a different layout structure, for example you can
          scroll down and see the footer is not present.
        </div>
      </div>

      <main className='flex-1 p-6'>
        <div className='max-w-4xl mx-auto'>
          <div className='bg-card rounded-lg shadow-sm border p-8'>
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}
