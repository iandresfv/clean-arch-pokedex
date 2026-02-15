import { expect, test } from '@playwright/test';

test.describe('Pokemon List Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('displays page title and description', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Pokédex' })).toBeVisible();
    await expect(page.getByText('Browse and discover Pokemon from all generations.')).toBeVisible();
  });

  test('displays header with logo', async ({ page }) => {
    await expect(page.getByRole('banner')).toBeVisible();
    await expect(page.getByRole('banner').getByText('Pokédex')).toBeVisible();
  });

  test('loads and displays pokemon cards', async ({ page }) => {
    await expect(page.getByRole('link', { name: /bulbasaur/i })).toBeVisible({ timeout: 10000 });
    const cards = page
      .locator('[data-testid="pokemon-card"]')
      .or(page.getByRole('link').filter({ hasText: /^#\d{3}/ }));
    await expect(cards.first()).toBeVisible();
  });

  test('displays pagination controls', async ({ page }) => {
    await expect(page.getByRole('button', { name: /previous/i })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
    await expect(page.getByText(/page 1 of/i)).toBeVisible();
  });

  test('navigates to next page', async ({ page }) => {
    await expect(page.getByText(/page 1 of/i)).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: /next/i }).click();
    await expect(page.getByText(/page 2 of/i)).toBeVisible();
  });

  test('displays search bar', async ({ page }) => {
    await expect(page.getByPlaceholder(/search/i)).toBeVisible();
  });

  test('displays footer with PokéAPI attribution', async ({ page }) => {
    await expect(page.getByText(/PokéAPI/i)).toBeVisible();
  });
});
