import { ArrowLeft } from 'lucide-react';
import { Link } from 'react-router';

import { buttonVariants } from '@/presentation/components/ui/button';

export function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center gap-6 py-20">
      <div className="relative h-24 w-24">
        <div className="absolute inset-0 rounded-full border-4 border-muted-foreground/30" />
        <div className="absolute inset-0 flex items-center">
          <div className="h-1 w-full bg-muted-foreground/30" />
        </div>
        <div className="absolute left-1/2 top-1/2 h-6 w-6 -translate-x-1/2 -translate-y-1/2 rounded-full border-4 border-muted-foreground/30 bg-background" />
      </div>
      <div className="text-center">
        <h1 className="text-6xl font-bold text-muted-foreground">404</h1>
        <p className="mt-2 text-xl text-muted-foreground">This page could not be found.</p>
      </div>
      <Link to="/" className={buttonVariants({ variant: 'outline' })}>
        <ArrowLeft className="mr-2 h-4 w-4" />
        Go back home
      </Link>
    </div>
  );
}
