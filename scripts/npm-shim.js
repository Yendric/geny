#!/usr/bin/env node
const { spawnSync } = require('node:child_process')

const pkg = `@yendric/geny-${process.platform}-${process.arch}`

let binary
try {
  binary = require.resolve(`${pkg}/bin/geny${process.platform === 'win32' ? '.exe' : ''}`)
} catch {
  console.error(`geny: unsupported platform, or the package ${pkg} is missing`)
  console.error('geny: reinstall with optional dependencies enabled')
  process.exit(1)
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: 'inherit' })
process.exit(result.status ?? 1)
