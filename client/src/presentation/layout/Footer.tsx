export function Footer() {
  return (
    <footer className="border-t py-6">
      <div className="container mx-auto max-w-screen-xl px-4 text-center text-sm text-muted-foreground">
        <p>
          Data provided by{' '}
          <a
            href="https://pokeapi.co"
            target="_blank"
            rel="noopener noreferrer"
            className="font-medium underline underline-offset-4 hover:text-foreground"
          >
            PokéAPI
          </a>
          . Built with Clean Architecture.
        </p>
      </div>
    </footer>
  );
}
