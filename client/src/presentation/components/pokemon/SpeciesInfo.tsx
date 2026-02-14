import type { SpeciesDTO } from '@/application/dto';
import { Badge } from '@/presentation/components/ui/badge';

interface SpeciesInfoProps {
  species: SpeciesDTO;
}

export function SpeciesInfo({ species }: SpeciesInfoProps) {
  return (
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Species Info</h3>
      <p className="text-sm text-muted-foreground leading-relaxed">{species.flavorText}</p>
      <div className="flex flex-wrap gap-2">
        <Badge variant="secondary">Gen {species.generation}</Badge>
        {species.habitat && <Badge variant="secondary">{species.habitat}</Badge>}
        {species.isLegendary && <Badge variant="default">Legendary</Badge>}
        {species.isMythical && <Badge variant="default">Mythical</Badge>}
      </div>
    </div>
  );
}
