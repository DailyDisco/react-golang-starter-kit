import { Link, useLocation } from 'react-router';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { Menu } from 'lucide-react';
import { useState } from 'react';
import { API_BASE_URL } from '../lib/api';

export function Navbar() {
  const location = useLocation();
  const [isOpen, setIsOpen] = useState(false);

  const navigation = [
    { name: 'Home', href: '/' },
    { name: 'Demo', href: '/demo' },
    { name: 'Layout Demo', href: '/layout-demo' },
    {
      name: 'API Docs',
      href: `${API_BASE_URL}/swagger/`,
      external: true,
    },
  ];

  const isActive = (href: string) => {
    if (href === '/' && location.pathname === '/') return true;
    if (href !== '/' && location.pathname.startsWith(href)) return true;
    return false;
  };

  return (
    <nav className='bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 sticky top-0 z-50'>
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'>
        <div className='flex justify-between h-16'>
          <div className='flex'>
            <div className='flex-shrink-0 flex items-center'>
              <Link
                to='/'
                className='text-xl font-bold text-gray-900 dark:text-white'
              >
                React + Go
              </Link>
            </div>
            <div className='hidden sm:ml-6 sm:flex sm:space-x-8'>
              {navigation.map(item => (
                <Link
                  key={item.name}
                  to={item.href}
                  target={item.external ? '_blank' : undefined}
                  rel={item.external ? 'noopener noreferrer' : undefined}
                  className={`inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium ${
                    isActive(item.href)
                      ? 'border-blue-500 text-gray-900 dark:text-white'
                      : 'border-transparent text-gray-500 dark:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 hover:text-gray-700 dark:hover:text-gray-100'
                  }`}
                >
                  {item.name}
                </Link>
              ))}
            </div>
          </div>
          <div className='hidden sm:ml-6 sm:flex sm:items-center'>
            <ThemeToggle />
          </div>
          <div className='-mr-2 flex items-center sm:hidden'>
            <ThemeToggle />
            <Sheet open={isOpen} onOpenChange={setIsOpen}>
              <SheetTrigger asChild>
                <Button variant='ghost' size='sm' className='ml-2'>
                  <Menu className='h-6 w-6' />
                  <span className='sr-only'>Open main menu</span>
                </Button>
              </SheetTrigger>
              <SheetContent side='right' className='w-[300px] sm:w-[400px]'>
                <div className='flex flex-col space-y-4 mt-4'>
                  {navigation.map(item => (
                    <Link
                      key={item.name}
                      to={item.href}
                      target={item.external ? '_blank' : undefined}
                      rel={item.external ? 'noopener noreferrer' : undefined}
                      onClick={() => setIsOpen(false)}
                      className={`inline-flex items-center px-3 py-2 rounded-md text-base font-medium ${
                        isActive(item.href)
                          ? 'bg-blue-50 dark:bg-blue-900 text-blue-700 dark:text-blue-200'
                          : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800'
                      }`}
                    >
                      {item.name}
                    </Link>
                  ))}
                </div>
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </div>
    </nav>
  );
}
