import type { StatsData } from '@/domain/pokemon';
import { cn } from '@/lib/utils';

interface StatsChartProps {
  stats: StatsData;
}

const STAT_CONFIG: { key: keyof StatsData; label: string; color: string }[] = [
  { key: 'hp', label: 'HP', color: 'bg-red-500' },
  { key: 'attack', label: 'Attack', color: 'bg-orange-500' },
  { key: 'defense', label: 'Defense', color: 'bg-yellow-500' },
  { key: 'specialAttack', label: 'Sp. Atk', color: 'bg-blue-500' },
  { key: 'specialDefense', label: 'Sp. Def', color: 'bg-green-500' },
  { key: 'speed', label: 'Speed', color: 'bg-pink-500' },
];

const MAX_STAT = 255;

export function StatsChart({ stats }: StatsChartProps) {
  const total =
    stats.hp +
    stats.attack +
    stats.defense +
    stats.specialAttack +
    stats.specialDefense +
    stats.speed;

  return (
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Base Stats</h3>
      <div className="space-y-2">
        {STAT_CONFIG.map(({ key, label, color }) => {
          const value = stats[key];
          const percentage = Math.min((value / MAX_STAT) * 100, 100);

          return (
            <div key={key} className="flex items-center gap-3">
              <span className="w-16 text-xs font-medium text-muted-foreground">{label}</span>
              <span className="w-8 text-right text-sm font-semibold">{value}</span>
              <div className="flex-1 h-2 rounded-full bg-muted">
                <div
                  className={cn('h-full rounded-full transition-all', color)}
                  style={{ width: `${percentage.toString()}%` }}
                />
              </div>
            </div>
          );
        })}
      </div>
      <div className="flex items-center gap-3 pt-1 border-t">
        <span className="w-16 text-xs font-medium text-muted-foreground">Total</span>
        <span className="w-8 text-right text-sm font-bold">{total}</span>
      </div>
    </div>
  );
}
