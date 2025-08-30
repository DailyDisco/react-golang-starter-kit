import { Outlet } from '@tanstack/react-router';
import { Navbar } from './Navbar';
import { Footer } from './Footer';
import { Breadcrumbs } from '../components/ui/breadcrumbs';

export default function StandardLayout() {
  return (
    <div className='min-h-screen flex flex-col'>
      <Navbar />
      <div className='border-b bg-muted/30'>
        <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-3'>
          <Breadcrumbs />
        </div>
      </div>
      <main className='flex-1'>
        <Outlet />
      </main>
      <Footer />
    </div>
  );
}
