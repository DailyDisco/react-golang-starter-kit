import { Link, useLocation } from '@tanstack/react-router';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Menu, User, LogOut, Settings } from 'lucide-react';
import { useState } from 'react';
import { useAuth } from '../hooks/auth/useAuth';
import { API_BASE_URL } from '../services';

export function Navbar() {
  const location = useLocation();
  const [isOpen, setIsOpen] = useState(false);
  const { user, isAuthenticated, logout } = useAuth();

  const navigation = [
    { name: 'Home', href: '/' },
    { name: 'About', href: '/about' },
    { name: 'Blog', href: '/blog' },
    { name: 'Search', href: '/search' },
    { name: 'Demo', href: '/demo' },
    { name: 'Layout Demo', href: '/layout-demo' },
    {
      name: 'API Docs',
      href: `${API_BASE_URL}/swagger/`,
      external: true,
    },
  ];

  const handleLogout = () => {
    logout();
  };

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
                search={{}}
                className='text-xl font-bold text-gray-900 dark:text-white'
              >
                React + Go
              </Link>
            </div>
            <div className='hidden sm:ml-6 sm:flex sm:space-x-8'>
              {navigation.map(item =>
                item.external ? (
                  <a
                    key={item.name}
                    href={item.href}
                    target='_blank'
                    rel='noopener noreferrer'
                    className={`inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium border-transparent text-gray-500 dark:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 hover:text-gray-700 dark:hover:text-gray-100`}
                  >
                    {item.name}
                  </a>
                ) : (
                  <Link
                    key={item.name}
                    to={item.href}
                    search={{}}
                    className={`inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium ${isActive(item.href)
                      ? 'border-blue-500 text-gray-900 dark:text-white'
                      : 'border-transparent text-gray-500 dark:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 hover:text-gray-700 dark:hover:text-gray-100'
                      }`}
                  >
                    {item.name}
                  </Link>
                )
              )}
            </div>
          </div>
          <div className='hidden sm:ml-6 sm:flex sm:items-center sm:space-x-4'>
            <ThemeToggle />

            {isAuthenticated && user ? (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant='ghost'
                    className='relative h-8 w-8 rounded-full'
                  >
                    <Avatar className='h-8 w-8'>
                      <AvatarImage src='' alt={user.name} />
                      <AvatarFallback>
                        {user.name
                          .split(' ')
                          .map(n => n[0])
                          .join('')
                          .toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className='w-56' align='end' forceMount>
                  <div className='flex items-center justify-start gap-2 p-2'>
                    <div className='flex flex-col space-y-1 leading-none'>
                      <p className='font-medium'>{user.name}</p>
                      <p className='w-[200px] truncate text-sm text-muted-foreground'>
                        {user.email}
                      </p>
                    </div>
                  </div>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem asChild>
                    <Link to='/profile' search={{}} className='cursor-pointer'>
                      <User className='mr-2 h-4 w-4' />
                      <span>Profile</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={handleLogout}
                    className='cursor-pointer'
                  >
                    <LogOut className='mr-2 h-4 w-4' />
                    <span>Log out</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            ) : (
              <div className='flex items-center space-x-2'>
                <Button variant='ghost' asChild>
                  <Link to='/login' search={{}}>
                    Sign in
                  </Link>
                </Button>
                <Button asChild>
                  <Link to='/register' search={{}}>
                    Sign up
                  </Link>
                </Button>
              </div>
            )}
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
              <SheetContent side='right' className='w-[320px]'>
                <div className='flex flex-col h-full'>
                  {/* Header */}
                  <div className='border-b pb-4 mb-4'>
                    <Link
                      to='/'
                      search={{}}
                      className='text-xl font-bold text-gray-900 dark:text-white'
                      onClick={() => setIsOpen(false)}
                    >
                      React + Go
                    </Link>
                  </div>

                  {/* Navigation */}
                  <div className='flex-1'>
                    <div className='space-y-1 mb-6'>
                      <p className='text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2'>
                        Navigation
                      </p>
                      {navigation.map(item => (
                        <Link
                          key={item.name}
                          to={item.href}
                          target={item.external ? '_blank' : undefined}
                          rel={item.external ? 'noopener noreferrer' : undefined}
                          onClick={() => setIsOpen(false)}
                          className={`flex items-center px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${isActive(item.href)
                            ? 'bg-primary text-primary-foreground'
                            : 'text-foreground hover:bg-muted'
                            }`}
                        >
                          {item.name}
                          {item.external && (
                            <span className='ml-auto text-xs text-muted-foreground'>↗</span>
                          )}
                        </Link>
                      ))}
                    </div>
                  </div>

                  {/* User Section */}
                  <div className='border-t pt-4'>
                    {isAuthenticated && user ? (
                      <>
                        <div className='flex items-center px-3 py-3 mb-4 bg-muted/50 rounded-lg'>
                          <Avatar className='h-10 w-10 mr-3'>
                            <AvatarImage src='' alt={user.name} />
                            <AvatarFallback className='text-sm'>
                              {user.name
                                .split(' ')
                                .map(n => n[0])
                                .join('')
                                .toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div className='flex-1 min-w-0'>
                            <p className='font-medium text-sm truncate'>{user.name}</p>
                            <p className='text-xs text-muted-foreground truncate'>
                              {user.email}
                            </p>
                          </div>
                        </div>

                        <div className='space-y-1'>
                          <p className='text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2'>
                            Account
                          </p>
                          <Link
                            to='/profile'
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            className='flex items-center px-3 py-2 rounded-lg text-sm font-medium text-foreground hover:bg-muted'
                          >
                            <User className='mr-3 h-4 w-4' />
                            Profile
                          </Link>
                          <Link
                            to='/profile'
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            className='flex items-center px-3 py-2 rounded-lg text-sm font-medium text-foreground hover:bg-muted'
                          >
                            <Settings className='mr-3 h-4 w-4' />
                            Settings
                          </Link>
                          <button
                            onClick={() => {
                              handleLogout();
                              setIsOpen(false);
                            }}
                            className='flex items-center w-full px-3 py-2 rounded-lg text-sm font-medium text-foreground hover:bg-muted'
                          >
                            <LogOut className='mr-3 h-4 w-4' />
                            Sign Out
                          </button>
                        </div>
                      </>
                    ) : (
                      <div className='space-y-2'>
                        <p className='text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2'>
                          Authentication
                        </p>
                        <Link
                          to='/login'
                          search={{}}
                          onClick={() => setIsOpen(false)}
                          className='flex items-center px-3 py-2 rounded-lg text-sm font-medium text-foreground hover:bg-muted'
                        >
                          Sign In
                        </Link>
                        <Link
                          to='/register'
                          search={{}}
                          onClick={() => setIsOpen(false)}
                          className='flex items-center px-3 py-2 rounded-lg text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90'
                        >
                          Sign Up
                        </Link>
                      </div>
                    )}
                  </div>
                </div>
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </div>
    </nav>
  );
}
