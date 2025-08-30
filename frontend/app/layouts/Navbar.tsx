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
    <nav className='bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 sticky top-0 z-50 shadow-sm'>
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'>
        <div className='flex justify-between h-16 items-center'>
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
            <div className='hidden md:ml-6 md:flex md:space-x-1'>
              {navigation.map(item =>
                item.external ? (
                  <a
                    key={item.name}
                    href={item.href}
                    target='_blank'
                    rel='noopener noreferrer'
                    className='inline-flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors duration-200 border-transparent text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
                  >
                    {item.name}
                    <span className='ml-1 text-xs opacity-60'>↗</span>
                  </a>
                ) : (
                  <Link
                    key={item.name}
                    to={item.href}
                    search={{}}
                    className={`inline-flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${isActive(item.href)
                      ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 border border-blue-200 dark:border-blue-800'
                      : 'text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-100'
                      }`}
                  >
                    {item.name}
                  </Link>
                )
              )}
            </div>
          </div>
          {/* User Controls */}
          <div className='hidden md:ml-6 md:flex md:items-center md:space-x-4'>
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
          <div className='-mr-2 flex items-center md:hidden'>
            <ThemeToggle />
            <Sheet open={isOpen} onOpenChange={setIsOpen}>
              <SheetTrigger asChild>
                <Button
                  variant='ghost'
                  size='sm'
                  className='ml-2 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors'
                  aria-label={isOpen ? 'Close main menu' : 'Open main menu'}
                  aria-expanded={isOpen}
                  aria-controls='mobile-menu'
                >
                  <Menu className='h-6 w-6' />
                  <span className='sr-only'>{isOpen ? 'Close main menu' : 'Open main menu'}</span>
                </Button>
              </SheetTrigger>
              <SheetContent side='right' className='w-[320px] bg-white dark:bg-gray-900 border-l border-gray-200 dark:border-gray-700' id='mobile-menu'>
                <div className='flex flex-col h-full p-4' role='menu'>
                  {/* Header */}
                  <div className='border-b border-gray-200 dark:border-gray-700 pb-6 mb-6'>
                    <Link
                      to='/'
                      search={{}}
                      className='text-xl font-bold text-gray-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400 transition-colors'
                      onClick={() => setIsOpen(false)}
                    >
                      React + Go
                    </Link>
                  </div>

                  {/* Navigation */}
                  <div className='flex-1'>
                    <div className='space-y-1 mb-8'>
                      <p className='text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-4'>
                        Navigation
                      </p>
                      {navigation.map(item => (
                        <Link
                          key={item.name}
                          to={item.href}
                          target={item.external ? '_blank' : undefined}
                          rel={
                            item.external ? 'noopener noreferrer' : undefined
                          }
                          onClick={() => setIsOpen(false)}
                          role='menuitem'
                          className={`flex items-center px-4 py-3 mx-2 rounded-lg text-sm font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${isActive(item.href)
                            ? 'bg-blue-600 text-white shadow-sm'
                            : 'text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800/50 hover:text-gray-900 dark:hover:text-gray-100 hover:shadow-sm'
                            }`}
                        >
                          {item.name}
                          {item.external && (
                            <span className='ml-auto text-xs opacity-70'>
                              ↗
                            </span>
                          )}
                        </Link>
                      ))}
                    </div>
                  </div>

                  {/* User Section */}
                  <div className='border-t border-gray-200 dark:border-gray-700 pt-6'>
                    {isAuthenticated && user ? (
                      <>
                        <div className='flex items-center p-4 mb-6 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/50 dark:to-indigo-950/50 rounded-lg border border-blue-100 dark:border-blue-900/50 shadow-sm'>
                          <Avatar className='h-10 w-10 mr-3 ring-2 ring-blue-200 dark:ring-blue-800'>
                            <AvatarImage src='' alt={user.name} />
                            <AvatarFallback className='text-sm bg-blue-600 text-white'>
                              {user.name
                                .split(' ')
                                .map(n => n[0])
                                .join('')
                                .toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div className='flex-1 min-w-0'>
                            <p className='font-medium text-sm truncate text-gray-900 dark:text-white'>
                              {user.name}
                            </p>
                            <p className='text-xs text-muted-foreground truncate'>
                              {user.email}
                            </p>
                          </div>
                        </div>

                        <div className='space-y-1'>
                          <p className='text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-4'>
                            Account
                          </p>
                          <Link
                            to='/profile'
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            role='menuitem'
                            className='flex items-center mx-2 px-4 py-3 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800/50 hover:text-gray-900 dark:hover:text-gray-100 hover:shadow-sm transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
                          >
                            <User className='mr-3 h-4 w-4' />
                            Profile
                          </Link>
                          <Link
                            to='/settings'
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            role='menuitem'
                            className='flex items-center mx-2 px-4 py-3 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800/50 hover:text-gray-900 dark:hover:text-gray-100 hover:shadow-sm transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
                          >
                            <Settings className='mr-3 h-4 w-4' />
                            Settings
                          </Link>
                          <button
                            onClick={() => {
                              handleLogout();
                              setIsOpen(false);
                            }}
                            role='menuitem'
                            className='flex items-center w-full mx-2 px-4 py-3 rounded-lg text-sm font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/50 hover:text-red-700 dark:hover:text-red-300 hover:shadow-sm transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2'
                          >
                            <LogOut className='mr-3 h-4 w-4' />
                            Sign Out
                          </button>
                        </div>
                      </>
                    ) : (
                      <div className='space-y-1'>
                        <p className='text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-4'>
                          Authentication
                        </p>
                        <Link
                          to='/login'
                          search={{}}
                          onClick={() => setIsOpen(false)}
                          role='menuitem'
                          className='flex items-center mx-2 px-4 py-3 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800/50 hover:text-gray-900 dark:hover:text-gray-100 hover:shadow-sm transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
                        >
                          Sign In
                        </Link>
                        <Link
                          to='/register'
                          search={{}}
                          onClick={() => setIsOpen(false)}
                          role='menuitem'
                          className='flex items-center mx-2 px-4 py-3 rounded-lg text-sm font-medium bg-gradient-to-r from-blue-600 to-blue-700 text-white hover:from-blue-700 hover:to-blue-800 shadow-sm hover:shadow-md transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
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
