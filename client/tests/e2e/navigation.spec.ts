import { expect, test } from '@playwright/test';

test.describe('Navigation', () => {
  test('shows 404 page for invalid routes', async ({ page }) => {
    await page.goto('/invalid-route');
    await expect(page.getByText('404')).toBeVisible();
    await expect(page.getByText('This page could not be found.')).toBeVisible();
  });

  test('navigates home from 404 page', async ({ page }) => {
    await page.goto('/invalid-route');
    await page.getByRole('link', { name: /go back home/i }).click();
    await expect(page).toHaveURL('/');
  });

  test('header logo navigates home', async ({ page }) => {
    await page.goto('/pokemon/25');
    await page
      .getByRole('banner')
      .getByRole('link', { name: /pokédex/i })
      .click();
    await expect(page).toHaveURL('/');
  });
});
