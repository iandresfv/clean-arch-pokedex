import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import type { StatsData } from '@/domain/pokemon';
import { StatsChart } from '@/presentation/components/pokemon/StatsChart';

const mockStats: StatsData = {
  hp: 35,
  attack: 55,
  defense: 40,
  specialAttack: 50,
  specialDefense: 65,
  speed: 90,
};

describe('StatsChart', () => {
  it('renders all stat labels', () => {
    render(<StatsChart stats={mockStats} />);

    expect(screen.getByText('HP')).toBeInTheDocument();
    expect(screen.getByText('Attack')).toBeInTheDocument();
    expect(screen.getByText('Defense')).toBeInTheDocument();
    expect(screen.getByText('Sp. Atk')).toBeInTheDocument();
    expect(screen.getByText('Sp. Def')).toBeInTheDocument();
    expect(screen.getByText('Speed')).toBeInTheDocument();
  });

  it('renders stat values', () => {
    render(<StatsChart stats={mockStats} />);

    expect(screen.getByText('35')).toBeInTheDocument();
    expect(screen.getByText('55')).toBeInTheDocument();
    expect(screen.getByText('90')).toBeInTheDocument();
  });

  it('renders total stat sum', () => {
    render(<StatsChart stats={mockStats} />);

    expect(screen.getByText('335')).toBeInTheDocument();
  });

  it('renders section title', () => {
    render(<StatsChart stats={mockStats} />);

    expect(screen.getByText('Base Stats')).toBeInTheDocument();
  });
});
