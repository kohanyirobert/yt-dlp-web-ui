import { test, expect } from '@playwright/test'

const TEST_VIDEO_URL = 'https://www.youtube.com/watch?v=2PuFyjAs7JA'

test('downloads a YouTube video via the UI', async ({ page, baseURL }) => {
  await page.goto('/')
  await page.waitForLoadState('networkidle')

  // Open the speed dial menu, then click "New download"
  await page.getByLabel('Home speed dial').click()
  await page.getByRole('menuitem', { name: /new download/i }).click()

  // Wait for the download dialog to appear
  await expect(page.getByText('Download').first()).toBeVisible()

  // Fill in the video URL
  await page.getByLabel(/video url/i).fill(TEST_VIDEO_URL)

  // Click the Start button
  await page.getByRole('button', { name: /start/i }).click()

  // Wait for a download card to appear with "Downloading" or "Pending" status
  await expect(
    page.getByText(/downloading|pending/i).first()
  ).toBeVisible({ timeout: 30_000 })

  // Wait for the download to complete (status chip shows "Completed")
  await expect(
    page.getByText('Completed').first()
  ).toBeVisible({ timeout: 120_000 })

  // Verify the downloaded file exists and is a video via the filebrowser API
  const res = await page.request.post(`${baseURL}/filebrowser/downloaded`, {
    data: { subdir: '', order_by: 'modtime' },
  })
  expect(res.ok()).toBeTruthy()

  const files: { name: string; size: number; path: string }[] = await res.json()
  const videoFile = files.find(f => /\.(mp4|webm|mkv)$/.test(f.name))

  expect(videoFile).toBeDefined()
  expect(videoFile!.size).toBeGreaterThan(10_000)

  // Clean up: delete the downloaded file via the filebrowser API
  const delRes = await page.request.post(`${baseURL}/filebrowser/delete`, {
    data: { path: videoFile!.path },
  })
  expect(delRes.ok()).toBeTruthy()
})
