import { Outlet } from 'react-router';

import { Header } from './Header';

export function AppLayout() {
  return (
    <div className="relative flex min-h-screen flex-col">
      <Header />
      <main className="container mx-auto flex-1 px-4 py-6 max-w-screen-xl">
        <Outlet />
      </main>
    </div>
  );
}
