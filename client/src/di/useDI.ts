import { useContext } from 'react';

import type { DIContainer } from './container';
import { DIContext } from './DIContext';

export function useDI(): DIContainer {
  const context = useContext(DIContext);

  if (!context) {
    throw new Error('useDI must be used within a DIProvider');
  }

  return context;
}
