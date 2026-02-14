import { createContext } from 'react';

import type { DIContainer } from './container';

export const DIContext = createContext<DIContainer | null>(null);
