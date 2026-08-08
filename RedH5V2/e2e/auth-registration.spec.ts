import { expect, test } from '@playwright/test'

async function openRegistration(page: import('@playwright/test').Page) {
  const desktopRegister = page.locator('.ppmx-auth-button:not(.ppmx-auth-button--login)')

  if (await desktopRegister.isVisible()) {
    await desktopRegister.click()
    return
  }

  await page.locator('#navMenu').click()
  await page.locator('.ppmx-drawer-auth--red').click()
}

test('shows an inline error and stays on step one for an invalid registration phone', async ({ page }) => {
  await page.goto('/#/')
  await openRegistration(page)

  const modal = page.locator('.ppmx-auth-card')
  await expect(modal).toBeVisible()
  await modal.locator('input[name="phone"]').fill('123456789')
  await modal.locator('input[name="firstName"]').fill('Prueba')
  await modal.locator('.ppmx-auth-submit').click()

  await expect(modal.getByRole('alert')).toHaveText(/Ingresa un celular válido de Perú|Enter a valid Peru mobile phone|请输入有效的秘鲁手机号/)
  await expect(modal.locator('input[name="phone"]')).toBeVisible()
  await expect(modal.locator('input[name="password"]')).toHaveCount(0)
})
