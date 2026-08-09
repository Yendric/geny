// Builds and publishes the npm packages from a finished goreleaser run:
// one binary package per platform, plus a meta package that picks the right 
// one at runtime
//
// Requires dist/ from `goreleaser release` and npm auth
// Set DRY_RUN=1 to build the packages without publishing

import { execFileSync } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'

const SCOPE = '@yendric'
const REPOSITORY = 'github:Yendric/geny'

const PLATFORMS = {
  'linux/amd64': { os: 'linux', cpu: 'x64' },
  'linux/arm64': { os: 'linux', cpu: 'arm64' },
  'darwin/amd64': { os: 'darwin', cpu: 'x64' },
  'darwin/arm64': { os: 'darwin', cpu: 'arm64' },
  'windows/amd64': { os: 'win32', cpu: 'x64' },
  'windows/arm64': { os: 'win32', cpu: 'arm64' },
}

const version = JSON.parse(fs.readFileSync('dist/metadata.json', 'utf8')).version
const artifacts = JSON.parse(fs.readFileSync('dist/artifacts.json', 'utf8'))
const binaries = artifacts.filter((a) => a.type === 'Binary')

const packageDirs = []

function writePackage(dir, packageJson) {
  fs.mkdirSync(path.join(dir, 'bin'), { recursive: true })
  fs.writeFileSync(path.join(dir, 'package.json'), JSON.stringify(packageJson, null, 2) + '\n')
  fs.copyFileSync('LICENSE', path.join(dir, 'LICENSE'))
  packageDirs.push(dir)
}

const optionalDependencies = {}
for (const binary of binaries) {
  const platform = PLATFORMS[`${binary.goos}/${binary.goarch}`]
  if (!platform) continue

  const name = `${SCOPE}/geny-${platform.os}-${platform.cpu}`
  const dir = `dist/npm/geny-${platform.os}-${platform.cpu}`
  writePackage(dir, {
    name,
    version,
    description: `geny binary for ${platform.os} ${platform.cpu}`,
    repository: REPOSITORY,
    license: 'MIT',
    os: [platform.os],
    cpu: [platform.cpu],
  })

  const binaryName = platform.os === 'win32' ? 'geny.exe' : 'geny'
  fs.copyFileSync(binary.path, path.join(dir, 'bin', binaryName))
  fs.chmodSync(path.join(dir, 'bin', binaryName), 0o755)

  optionalDependencies[name] = version
}

if (Object.keys(optionalDependencies).length === 0) {
  throw new Error('no binaries found in dist/artifacts.json')
}

const metaDir = 'dist/npm/geny'
writePackage(metaDir, {
  name: `${SCOPE}/geny`,
  version,
  description: 'A small and efficient static site generator written in Go.',
  repository: REPOSITORY,
  license: 'MIT',
  bin: { geny: 'bin/geny.js' },
  optionalDependencies,
})
fs.copyFileSync('scripts/npm-shim.js', path.join(metaDir, 'bin', 'geny.js'))

for (const dir of packageDirs) {
  if (process.env.DRY_RUN) {
    console.log(`dry run: skipping npm publish of ${dir}`)
    continue
  }
  execFileSync('npm', ['publish', '--access', 'public'], { cwd: dir, stdio: 'inherit' })
}
