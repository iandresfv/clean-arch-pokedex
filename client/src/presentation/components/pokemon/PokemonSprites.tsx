interface PokemonSpritesProps {
  sprites: {
    frontDefault: string | null;
    frontShiny: string | null;
    backDefault: string | null;
    backShiny: string | null;
  };
  name: string;
}

const SPRITE_LABELS = [
  { key: 'frontDefault' as const, label: 'Front' },
  { key: 'backDefault' as const, label: 'Back' },
  { key: 'frontShiny' as const, label: 'Shiny Front' },
  { key: 'backShiny' as const, label: 'Shiny Back' },
];

export function PokemonSprites({ sprites, name }: PokemonSpritesProps) {
  const availableSprites = SPRITE_LABELS.filter(({ key }) => sprites[key] !== null);

  if (availableSprites.length === 0) return null;

  return (
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Sprites</h3>
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
        {availableSprites.map(({ key, label }) => {
          const url = sprites[key];
          if (!url) return null;
          return (
            <div key={key} className="flex flex-col items-center gap-1 rounded-lg bg-muted/50 p-2">
              <img src={url} alt={`${name} ${label}`} className="h-16 w-16 object-contain" />
              <span className="text-xs text-muted-foreground">{label}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
