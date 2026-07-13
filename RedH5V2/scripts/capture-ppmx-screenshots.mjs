import { existsSync, mkdirSync } from 'node:fs'
import { dirname, join, resolve } from 'node:path'
import { spawnSync } from 'node:child_process'
import { fileURLToPath, pathToFileURL } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const rootDir = resolve(scriptDir, '..')
const chromeCandidates = [
  'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
  'C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe',
  'C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe',
  'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe',
]
const chromePath = process.env.CHROME_PATH || chromeCandidates.find(existsSync)
const screenshotDir = join(rootDir, 'docs', 'reference', 'ppmx-home', 'screenshots')
const referencePath = join(rootDir, 'docs', 'reference', 'ppmx-home', 'pp-mx-apple-v2.reference.html')
const referenceUrl = pathToFileURL(referencePath).href
const vueUrl = process.env.PPMX_HOME_URL || 'http://127.0.0.1:4173/'
const viewports = [
  '375x812',
  '390x844',
  '430x932',
  '768x1024',
  '1024x768',
  '1440x900',
]

function fail(message) {
  console.error(`[capture-ppmx] ${message}`)
  process.exit(1)
}

function capture(label, url, viewport) {
  const output = join(screenshotDir, `${label}-${viewport}.png`)
  const result = spawnSync(
    chromePath,
    [
      '--headless=new',
      '--disable-gpu',
      '--hide-scrollbars',
      '--no-first-run',
      '--disable-extensions',
      `--window-size=${viewport.replace('x', ',')}`,
      `--screenshot=${output}`,
      url,
    ],
    {
      encoding: 'utf8',
      stdio: 'pipe',
    },
  )

  if (result.status !== 0) {
    fail(`${label} ${viewport} failed: ${result.stderr || result.stdout}`)
  }

  console.info(`[capture-ppmx] wrote ${output}`)
}

if (!chromePath) fail('Chrome or Edge executable was not found; set CHROME_PATH')
if (!existsSync(referencePath)) fail('missing pp-mx-apple-v2.reference.html')

mkdirSync(screenshotDir, { recursive: true })

for (const viewport of viewports) {
  capture('reference', referenceUrl, viewport)
  capture('vue', vueUrl, viewport)
}

console.info('[capture-ppmx] screenshot capture complete')
