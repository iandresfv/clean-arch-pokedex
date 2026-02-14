import { Link } from 'react-router';

export function Header() {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto flex h-14 max-w-screen-xl items-center px-4">
        <Link to="/" className="flex items-center gap-2">
          <div className="h-8 w-8 rounded-lg bg-red-600 flex items-center justify-center">
            <div className="h-4 w-4 rounded-full border-2 border-white bg-white/30" />
          </div>
          <span className="text-lg font-bold tracking-tight">Pokédex</span>
        </Link>
      </div>
    </header>
  );
}
