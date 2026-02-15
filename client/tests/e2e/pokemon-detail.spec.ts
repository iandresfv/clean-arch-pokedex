import { expect, test } from '@playwright/test';

test.describe('Pokemon Detail Page', () => {
  test('navigates to detail page from list', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('link', { name: /bulbasaur/i })).toBeVisible({ timeout: 10000 });
    await page.getByRole('link', { name: /bulbasaur/i }).click();
    await expect(page).toHaveURL(/\/pokemon\/1$/);
  });

  test('displays pokemon details', async ({ page }) => {
    await page.goto('/pokemon/25');
    await expect(page.getByRole('heading', { name: /pikachu/i })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('electric', { exact: true })).toBeVisible();
  });

  test('displays stats chart', async ({ page }) => {
    await page.goto('/pokemon/25');
    await expect(page.getByText(/base stats/i)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('HP')).toBeVisible();
    await expect(page.getByText('Attack')).toBeVisible();
  });

  test('has back navigation link', async ({ page }) => {
    await page.goto('/pokemon/25');
    await expect(page.getByRole('link', { name: /back to list/i })).toBeVisible({ timeout: 10000 });
    await page.getByRole('link', { name: /back to list/i }).click();
    await expect(page).toHaveURL('/');
  });

  test('displays favorite button', async ({ page }) => {
    await page.goto('/pokemon/25');
    await expect(page.getByRole('button', { name: /add to favorites/i })).toBeVisible({
      timeout: 10000,
    });
  });
});
