import { Link } from 'react-router';

import { buttonVariants } from '@/presentation/components/ui/button';

export function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20">
      <h1 className="text-6xl font-bold text-muted-foreground">404</h1>
      <p className="text-xl text-muted-foreground">This page could not be found.</p>
      <Link to="/" className={buttonVariants({ variant: 'outline' })}>
        Go back home
      </Link>
    </div>
  );
}
