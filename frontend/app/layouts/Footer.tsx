export function Footer() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className='bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 mt-auto'>
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        <div className='text-center space-y-4'>
          <p className='text-gray-600 dark:text-gray-400'>
            Built with ❤️ using React & Go
          </p>
          <div className='flex flex-col sm:flex-row justify-center items-center gap-4 text-sm text-gray-500 dark:text-gray-400'>
            <p>© {currentYear} React + Go Starter Kit</p>
            <div className='hidden sm:block text-gray-300 dark:text-gray-600'>
              •
            </div>
            <p>
              Open source on{' '}
              <a
                href='https://github.com/DailyDisco/react-golang-starter-kit'
                target='_blank'
                rel='noopener noreferrer'
                className='text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors'
              >
                GitHub
              </a>
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
}
