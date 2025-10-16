export function Footer() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="mt-auto border-t border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-900">
      <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="space-y-4 text-center">
          <p className="text-gray-600 dark:text-gray-400">Built with ❤️ using React & Go</p>
          <div className="flex flex-col items-center justify-center gap-4 text-sm text-gray-500 sm:flex-row dark:text-gray-400">
            <p>© {currentYear} React + Go Starter Kit</p>
            <div className="hidden text-gray-300 sm:block dark:text-gray-600">•</div>
            <p>
              Open source on{" "}
              <a
                href="https://github.com/DailyDisco/react-golang-starter-kit"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 transition-colors hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
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
