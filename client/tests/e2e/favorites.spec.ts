import { expect, test } from '@playwright/test';

test.describe('Favorites', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.evaluate(() => { localStorage.clear(); });
    await page.reload();
  });

  test('toggles favorite from list page', async ({ page }) => {
    await expect(page.getByRole('link', { name: /bulbasaur/i })).toBeVisible({ timeout: 10000 });

    const firstCard = page.getByRole('link', { name: /bulbasaur/i });
    const favButton = firstCard.getByRole('button', { name: /add to favorites/i });
    await favButton.click();

    await expect(page.getByRole('banner').getByText('1')).toBeVisible();
  });

  test('toggles favorite from detail page', async ({ page }) => {
    await page.goto('/pokemon/25');
    await expect(page.getByRole('button', { name: /add to favorites/i })).toBeVisible({
      timeout: 10000,
    });

    await page.getByRole('button', { name: /add to favorites/i }).click();
    await expect(page.getByRole('banner').getByText('1')).toBeVisible();

    await page.getByRole('button', { name: /remove from favorites/i }).click();
    await expect(page.getByRole('banner').getByText('1')).not.toBeVisible();
  });
});
