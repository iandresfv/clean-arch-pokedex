import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { SearchBar } from '@/presentation/components/pokemon/SearchBar';

describe('SearchBar', () => {
  it('renders with placeholder text', () => {
    render(<SearchBar value="" onChange={vi.fn()} />);
    expect(screen.getByPlaceholderText('Search Pokemon...')).toBeInTheDocument();
  });

  it('renders with custom placeholder', () => {
    render(<SearchBar value="" onChange={vi.fn()} placeholder="Find a Pokemon" />);
    expect(screen.getByPlaceholderText('Find a Pokemon')).toBeInTheDocument();
  });

  it('calls onChange when typing', async () => {
    const onChange = vi.fn();
    render(<SearchBar value="" onChange={onChange} />);

    const input = screen.getByPlaceholderText('Search Pokemon...');
    await userEvent.type(input, 'p');

    expect(onChange).toHaveBeenCalledWith('p');
  });

  it('shows clear button when value is present', () => {
    render(<SearchBar value="pikachu" onChange={vi.fn()} />);
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('does not show clear button when value is empty', () => {
    render(<SearchBar value="" onChange={vi.fn()} />);
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('calls onChange with empty string when clear button is clicked', async () => {
    const onChange = vi.fn();
    render(<SearchBar value="pikachu" onChange={onChange} />);

    await userEvent.click(screen.getByRole('button'));
    expect(onChange).toHaveBeenCalledWith('');
  });
});
