import { Outlet } from 'react-router';

import { Footer } from './Footer';
import { Header } from './Header';

import { ErrorBoundary } from '@/presentation/components/shared/ErrorBoundary';
import { ScrollToTop } from '@/presentation/components/shared/ScrollToTop';

export function AppLayout() {
  return (
    <div className="relative flex min-h-screen flex-col">
      <ScrollToTop />
      <Header />
      <main className="container mx-auto flex-1 px-4 py-6 max-w-screen-xl">
        <ErrorBoundary>
          <Outlet />
        </ErrorBoundary>
      </main>
      <Footer />
    </div>
  );
}
