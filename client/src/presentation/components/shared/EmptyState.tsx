import { SearchX } from 'lucide-react';
import type { ReactNode } from 'react';

interface EmptyStateProps {
  title: string;
  description?: string;
  icon?: ReactNode;
}

export function EmptyState({ title, description, icon }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-16">
      <div className="text-muted-foreground/50">{icon ?? <SearchX className="h-12 w-12" />}</div>
      <p className="text-lg font-medium text-muted-foreground">{title}</p>
      {description && (
        <p className="max-w-sm text-center text-sm text-muted-foreground">{description}</p>
      )}
    </div>
  );
}
