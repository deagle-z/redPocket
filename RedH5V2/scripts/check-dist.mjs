import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const distDir = join(scriptDir, '..', 'dist')
const requiredFiles = ['index.html', 'manifest.webmanifest', 'sw.js', 'offline.html']
const FORBIDDEN_RUNTIME_DEPENDENCIES = [
  'fonts.googleapis.com',
  'fonts.gstatic.com',
  'cdnjs.cloudflare.com',
  'images.unsplash.com',
]

function fail(message) {
  console.error(`[check-dist] ${message}`)
  process.exitCode = 1
}

function assertFile(file) {
  const path = join(distDir, file)
  if (!existsSync(path)) fail(`missing ${file}`)
}

function walk(directory) {
  return readdirSync(directory).flatMap(name => {
    const path = join(directory, name)
    const stat = statSync(path)
    return stat.isDirectory() ? walk(path) : [path]
  })
}

requiredFiles.forEach(assertFile)

if (process.exitCode) process.exit()

const indexHtml = readFileSync(join(distDir, 'index.html'), 'utf8')
const manifest = JSON.parse(
  readFileSync(join(distDir, 'manifest.webmanifest'), 'utf8'),
)
const sw = readFileSync(join(distDir, 'sw.js'), 'utf8')
const files = walk(distDir)
const hasJs = files.some(file => file.endsWith('.js'))
const hasCss = files.some(file => file.endsWith('.css'))
const jsBundle = files
  .filter(file => file.endsWith('.js'))
  .map(file => readFileSync(file, 'utf8'))
  .join('\n')
const runtimeText = files
  .filter(file => /\.(html|js|css|webmanifest|json|svg)$/.test(file))
  .map(file => readFileSync(file, 'utf8'))
  .join('\n')

if (!indexHtml.includes('<div id="app"></div>')) {
  fail('index.html does not contain app mount node')
}

if (!manifest.name || !manifest.icons?.length) {
  fail('manifest is missing name or icons')
}

if (!hasJs) fail('missing built JavaScript assets')
if (!hasCss) fail('missing built CSS assets')
if (sw.includes('/api/')) fail('service worker should not precache API paths')
if (!jsBundle.includes('0.0.0')) {
  fail('release version marker was not embedded in JavaScript assets')
}

for (const dependency of FORBIDDEN_RUNTIME_DEPENDENCIES) {
  if (runtimeText.includes(dependency)) {
    fail(`forbidden remote runtime dependency found: ${dependency}`)
  }
}

if (!process.exitCode) {
  console.info('[check-dist] dist artifact checks passed')
}
