'use client';

import type { ThemeProviderProps } from 'next-themes';
import { ThemeProvider as NextThemesProvider } from 'next-themes';
import * as React from 'react';

export function ThemeProvider({ children, ...props }: ThemeProviderProps) {
  return (
    <NextThemesProvider
      {...props}
      enableSystem
      disableTransitionOnChange
      storageKey='theme'
      attribute='class'
    >
      {children}
    </NextThemesProvider>
  );
}
