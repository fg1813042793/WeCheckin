#!/usr/bin/env node
import { execFileSync, spawn } from 'node:child_process'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))

function printHelp() {
  console.log(`Usage:
  node build-ios-ipa.mjs --bundle-id com.example.app --identity "Apple Development: Name (TEAMID)" --profile /path/app.mobileprovision

Builds an installable signed iOS IPA from this uni-app project.

Options:
  --bundle-id <id>        iOS bundle identifier. Must match the provisioning profile.
  --identity <name>       Code signing identity from "security find-identity -p codesigning".
  --profile <file>        Provisioning profile (.mobileprovision).
  --output <file>         Output IPA path.
  --display-name <name>   App display name, default from manifest.json.
  --version <version>     CFBundleShortVersionString, default from manifest.json.
  --build <number>        CFBundleVersion, default from manifest.json versionCode.
  --app-id <appid>        uni-app appid, default from manifest.json.
  --hbuilderx <path>      HBuilderX app path, default /Applications/HBuilderX.app.
  --base-ipa <file>       Base iPhone IPA, default from HBuilderX.
  --temp <dir>            Temporary work directory.
  --skip-uni-build        Reuse existing unpackage/resources/<appid>/www.
  --unsigned              Build an unsigned IPA for inspection only; it is not installable.
  --keep-temp             Keep temporary files.
  --dry-run               Print resolved configuration and exit.
  --help                  Show this help.

Environment:
  IOS_BUNDLE_ID, IOS_SIGN_IDENTITY, IOS_MOBILEPROVISION, IOS_OUTPUT_IPA,
  IOS_DISPLAY_NAME, IOS_VERSION, IOS_BUILD_NUMBER, UNI_APP_ID, HBUILDERX_PATH,
  IOS_BASE_IPA, IOS_TEMP_DIR
`)
}

function parseArgs(argv) {
  const options = {
    projectRoot: scriptDir,
    appId: process.env.UNI_APP_ID || '',
    bundleId: process.env.IOS_BUNDLE_ID || '',
    identity: process.env.IOS_SIGN_IDENTITY || '',
    profile: process.env.IOS_MOBILEPROVISION || '',
    outputIpa: process.env.IOS_OUTPUT_IPA || '',
    displayName: process.env.IOS_DISPLAY_NAME || '',
    version: process.env.IOS_VERSION || '',
    buildNumber: process.env.IOS_BUILD_NUMBER || '',
    hbuilderxPath: process.env.HBUILDERX_PATH || '/Applications/HBuilderX.app',
    baseIpa: process.env.IOS_BASE_IPA || '',
    tempDir: process.env.IOS_TEMP_DIR || '',
    skipUniBuild: false,
    unsigned: false,
    keepTemp: false,
    dryRun: false
  }

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i]
    const next = () => {
      const value = argv[i + 1]
      if (!value || value.startsWith('--')) {
        throw new Error(`Missing value for ${arg}`)
      }
      i += 1
      return value
    }

    if (arg === '--help' || arg === '-h') {
      printHelp()
      process.exit(0)
    } else if (arg === '--bundle-id') {
      options.bundleId = next()
    } else if (arg === '--identity') {
      options.identity = next()
    } else if (arg === '--profile') {
      options.profile = next()
    } else if (arg === '--output') {
      options.outputIpa = next()
    } else if (arg === '--display-name') {
      options.displayName = next()
    } else if (arg === '--version') {
      options.version = next()
    } else if (arg === '--build') {
      options.buildNumber = next()
    } else if (arg === '--app-id') {
      options.appId = next()
    } else if (arg === '--hbuilderx') {
      options.hbuilderxPath = next()
    } else if (arg === '--base-ipa') {
      options.baseIpa = next()
    } else if (arg === '--temp') {
      options.tempDir = next()
    } else if (arg === '--skip-uni-build') {
      options.skipUniBuild = true
    } else if (arg === '--unsigned') {
      options.unsigned = true
    } else if (arg === '--keep-temp') {
      options.keepTemp = true
    } else if (arg === '--dry-run') {
      options.dryRun = true
    } else {
      throw new Error(`Unknown option: ${arg}`)
    }
  }

  return options
}

function readJson(file) {
  return JSON.parse(fs.readFileSync(file, 'utf8'))
}

function assertExists(file, label) {
  if (!fs.existsSync(file)) {
    throw new Error(`${label} not found: ${file}`)
  }
}

function isValidBundleId(value) {
  return /^[A-Za-z][A-Za-z0-9-]*(\.[A-Za-z0-9-]+)+$/.test(String(value || ''))
}

function safePackageName(value) {
  return String(value || 'app').replace(/[^0-9A-Za-z_.-]/g, '-')
}

function resolvePath(projectRoot, value) {
  return path.isAbsolute(value) ? value : path.resolve(projectRoot, value)
}

function deriveBundleId(manifest) {
  const ios = manifest?.['app-plus']?.distribute?.ios || {}
  const candidates = [
    ios.bundleIdentifier,
    ios.bundleId,
    ios.id,
    manifest?.['app-harmony']?.distribute?.bundleName,
    manifest?.harmony?.package,
    'com.wecheckin.app'
  ]
  return candidates.find(isValidBundleId) || 'com.wecheckin.app'
}

function resolveConfig(options) {
  const manifest = readJson(path.join(options.projectRoot, 'manifest.json'))
  const packageJsonPath = path.join(options.projectRoot, 'package.json')
  const packageJson = fs.existsSync(packageJsonPath) ? readJson(packageJsonPath) : {}
  const baseIpa = options.baseIpa || path.join(
    options.hbuilderxPath,
    'Contents/HBuilderX/plugins/launcher/base/iPhone_base.ipa'
  )
  const appId = options.appId || manifest.appid || 'HBuilder'
  const displayName = options.displayName || manifest.name || packageJson.name || appId
  const version = String(options.version || manifest.versionName || packageJson.version || '1.0.0')
  const buildNumber = String(options.buildNumber || manifest.versionCode || '1')
  const bundleId = options.bundleId || deriveBundleId(manifest)
  const outputIpa = options.outputIpa
    ? resolvePath(options.projectRoot, options.outputIpa)
    : path.join(
      options.projectRoot,
      'unpackage/dist/build',
      `${safePackageName(packageJson.name || appId)}-ios-${options.unsigned ? 'unsigned' : 'signed'}.ipa`
    )

  return {
    ...options,
    manifest,
    packageJson,
    appId,
    displayName,
    version,
    buildNumber,
    bundleId,
    baseIpa,
    outputIpa: path.isAbsolute(outputIpa) ? outputIpa : path.resolve(options.projectRoot, outputIpa)
  }
}

function printConfig(config) {
  console.log(JSON.stringify({
    projectRoot: config.projectRoot,
    appId: config.appId,
    bundleId: config.bundleId,
    displayName: config.displayName,
    version: config.version,
    buildNumber: config.buildNumber,
    hbuilderxPath: config.hbuilderxPath,
    baseIpa: config.baseIpa,
    outputIpa: config.outputIpa,
    skipUniBuild: config.skipUniBuild,
    unsigned: config.unsigned,
    signingIdentity: config.identity || null,
    provisioningProfile: config.profile || null
  }, null, 2))
}

function run(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: options.cwd,
      env: options.env || process.env,
      stdio: options.stdio || 'inherit'
    })

    child.on('error', reject)
    child.on('close', (code) => {
      if (code === 0) {
        resolve()
      } else {
        reject(new Error(`${command} ${args.join(' ')} exited with code ${code}`))
      }
    })
  })
}

function execText(command, args, options = {}) {
  try {
    return execFileSync(command, args, {
      cwd: options.cwd,
      env: options.env || process.env,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe']
    }).trim()
  } catch (error) {
    const stderr = error.stderr ? error.stderr.toString('utf8').trim() : ''
    throw new Error(`${command} ${args.join(' ')} failed${stderr ? `:\n${stderr}` : ''}`)
  }
}

async function buildUniAppResources(config) {
  if (config.skipUniBuild) {
    console.log('[1/6] Reusing existing uni app resources')
    return
  }

  console.log('[1/6] Building uni app resources')
  const uniBin = path.join(config.projectRoot, 'node_modules/.bin/uni')
  assertExists(uniBin, 'uni CLI')
  await run(uniBin, ['build', '--platform', 'app'], {
    cwd: config.projectRoot,
    env: { ...process.env, UNI_INPUT_DIR: '.' }
  })
}

function newestFirst(a, b) {
  return fs.statSync(b).mtimeMs - fs.statSync(a).mtimeMs
}

function findUniWww(config) {
  const expected = path.join(config.projectRoot, 'unpackage/resources', config.appId, 'www')
  if (fs.existsSync(path.join(expected, 'manifest.json'))) {
    return expected
  }

  const resourcesRoot = path.join(config.projectRoot, 'unpackage/resources')
  if (!fs.existsSync(resourcesRoot)) {
    throw new Error(`uni app resources not found: ${expected}`)
  }

  const candidates = fs.readdirSync(resourcesRoot)
    .map((name) => path.join(resourcesRoot, name, 'www'))
    .filter((candidate) => fs.existsSync(path.join(candidate, 'manifest.json')))
    .sort(newestFirst)

  if (candidates.length === 0) {
    throw new Error(`No uni app resources found under ${resourcesRoot}`)
  }

  console.warn(`Using ${candidates[0]} because ${expected} does not exist`)
  return candidates[0]
}

function escapeXml(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

function patchControlXml(controlPath, config) {
  const appTag = `<app appid="${escapeXml(config.appId)}" appver="${escapeXml(config.version)}"/>`
  let xml = fs.existsSync(controlPath) ? fs.readFileSync(controlPath, 'utf8') : ''

  if (/<apps>[\s\S]*?<\/apps>/.test(xml)) {
    xml = xml.replace(/<apps>[\s\S]*?<\/apps>/, `<apps>${appTag}</apps>`)
  } else if (/<hbuilder\b[^>]*>/.test(xml)) {
    xml = xml.replace(/(<hbuilder\b[^>]*>)/, `$1<apps>${appTag}</apps>`)
  } else {
    xml = `<?xml version="1.0"?>\n<hbuilder version="1.9.9.81691" ns="true" authority="*" debug="false" syncDebug="false"><apps>${appTag}</apps></hbuilder>`
  }

  xml = xml.replace(/\sdebug="true"/g, ' debug="false"')
  xml = xml.replace(/\ssyncDebug="true"/g, ' syncDebug="false"')
  fs.writeFileSync(controlPath, xml.endsWith('\n') ? xml : `${xml}\n`)
}

function plistReplaceString(plist, keyPath, value) {
  execFileSync('/usr/bin/plutil', ['-replace', keyPath, '-string', String(value), plist], {
    stdio: ['ignore', 'pipe', 'pipe']
  })
}

function plistDelete(plist, keyPath) {
  try {
    execFileSync('/usr/libexec/PlistBuddy', ['-c', `Delete :${keyPath}`, plist], {
      stdio: ['ignore', 'pipe', 'pipe']
    })
  } catch {
    // Optional cleanup.
  }
}

function patchInfoPlist(appDir, config) {
  const infoPlist = path.join(appDir, 'Info.plist')
  assertExists(infoPlist, 'iOS Info.plist')

  plistReplaceString(infoPlist, 'CFBundleIdentifier', config.bundleId)
  plistReplaceString(infoPlist, 'CFBundleDisplayName', config.displayName)
  plistReplaceString(infoPlist, 'CFBundleName', config.displayName)
  plistReplaceString(infoPlist, 'CFBundleShortVersionString', config.version)
  plistReplaceString(infoPlist, 'CFBundleVersion', config.buildNumber)

  try {
    plistReplaceString(infoPlist, 'CFBundleURLTypes.0.CFBundleURLName', config.bundleId)
    plistReplaceString(infoPlist, 'CFBundleURLTypes.0.CFBundleURLSchemes.0', config.bundleId.toLowerCase())
  } catch {
    // HBuilderX base IPA variants may not expose URLTypes in the same shape.
  }

  try {
    plistReplaceString(infoPlist, 'marketChannel', `${config.bundleId}|${config.displayName}||`)
  } catch {
    // Optional DCloud metadata.
  }

  plistDelete(infoPlist, 'UIApplicationShortcutItems')
}

function walk(root, visitor) {
  if (!fs.existsSync(root)) return
  const stat = fs.lstatSync(root)
  visitor(root, stat)
  if (!stat.isDirectory() || stat.isSymbolicLink()) return

  for (const entry of fs.readdirSync(root)) {
    walk(path.join(root, entry), visitor)
  }
}

function removeCodeSignatures(root) {
  const signatureDirs = []
  walk(root, (entry, stat) => {
    if (stat.isDirectory() && path.basename(entry) === '_CodeSignature') {
      signatureDirs.push(entry)
    }
  })

  signatureDirs
    .sort((a, b) => b.length - a.length)
    .forEach((entry) => fs.rmSync(entry, { recursive: true, force: true }))

  fs.rmSync(path.join(root, 'embedded.mobileprovision'), { force: true })
}

function plistPrint(plist, keyPath, required = true) {
  try {
    return execText('/usr/libexec/PlistBuddy', ['-c', `Print :${keyPath}`, plist])
  } catch (error) {
    if (required) throw error
    return ''
  }
}

function bundleHasExecutable(bundleDir) {
  const infoPlist = path.join(bundleDir, 'Info.plist')
  if (!fs.existsSync(infoPlist)) return false
  const executable = plistPrint(infoPlist, 'CFBundleExecutable', false)
  return Boolean(executable && fs.existsSync(path.join(bundleDir, executable)))
}

function findSignableTargets(appDir) {
  const targets = []
  walk(appDir, (entry, stat) => {
    if (entry === appDir) return
    if (stat.isDirectory()) {
      const ext = path.extname(entry)
      if (ext === '.framework' || ext === '.appex' || ext === '.app') {
        targets.push(entry)
      } else if (ext === '.bundle' && bundleHasExecutable(entry)) {
        targets.push(entry)
      }
    } else if (stat.isFile() && path.extname(entry) === '.dylib') {
      targets.push(entry)
    }
  })

  return [...new Set(targets)].sort((a, b) => {
    const depthDelta = b.split(path.sep).length - a.split(path.sep).length
    return depthDelta || b.length - a.length
  })
}

function profileBundleId(applicationIdentifier) {
  const value = String(applicationIdentifier || '')
  const dot = value.indexOf('.')
  return dot === -1 ? value : value.slice(dot + 1)
}

function bundleMatchesProfile(pattern, bundleId) {
  if (pattern === bundleId) return true
  if (pattern === '*') return true
  if (pattern.endsWith('.*')) {
    return bundleId.startsWith(pattern.slice(0, -1))
  }
  return false
}

function prepareSigning(config, tempRoot, appDir) {
  const profilePlist = path.join(tempRoot, 'profile.plist')
  const entitlementsFile = path.join(tempRoot, 'entitlements.plist')

  execFileSync('/usr/bin/security', ['cms', '-D', '-i', config.profile, '-o', profilePlist], {
    stdio: ['ignore', 'pipe', 'pipe']
  })

  const appIdentifier = plistPrint(profilePlist, 'Entitlements:application-identifier')
  const allowedBundleId = profileBundleId(appIdentifier)
  if (!bundleMatchesProfile(allowedBundleId, config.bundleId)) {
    throw new Error(`Provisioning profile allows ${allowedBundleId}, but bundle id is ${config.bundleId}`)
  }

  const entitlements = execFileSync('/usr/libexec/PlistBuddy', ['-x', '-c', 'Print :Entitlements', profilePlist])
  fs.writeFileSync(entitlementsFile, entitlements)
  fs.copyFileSync(config.profile, path.join(appDir, 'embedded.mobileprovision'))
  return entitlementsFile
}

async function signTarget(target, identity, entitlementsFile = '') {
  const args = ['--force', '--sign', identity, '--timestamp=none']
  if (entitlementsFile) {
    args.push('--entitlements', entitlementsFile, '--generate-entitlement-der')
  }
  args.push(target)
  await run('/usr/bin/codesign', args)
}

async function signApp(config, tempRoot, appDir) {
  console.log('[4/6] Signing iOS app')
  const entitlementsFile = prepareSigning(config, tempRoot, appDir)
  const nestedTargets = findSignableTargets(appDir)

  for (const target of nestedTargets) {
    await signTarget(target, config.identity)
  }

  await signTarget(appDir, config.identity, entitlementsFile)
  await run('/usr/bin/codesign', ['--verify', '--deep', '--strict', '--verbose=2', appDir])
}

function replaceAppResources(appDir, wwwDir, config) {
  const appsDir = path.join(appDir, 'Pandora/apps')
  const targetWww = path.join(appsDir, config.appId, 'www')

  fs.rmSync(appsDir, { recursive: true, force: true })
  fs.mkdirSync(path.dirname(targetWww), { recursive: true })
  fs.cpSync(wwwDir, targetWww, { recursive: true })
  patchControlXml(path.join(appDir, 'control.xml'), config)
  patchInfoPlist(appDir, config)
}

async function packageIpa(stageRoot, outputIpa) {
  fs.mkdirSync(path.dirname(outputIpa), { recursive: true })
  fs.rmSync(outputIpa, { force: true })
  await run('/usr/bin/zip', ['-qry', outputIpa, 'Payload'], { cwd: stageRoot })
}

function validateConfig(config) {
  if (!isValidBundleId(config.bundleId)) {
    throw new Error(`Invalid iOS bundle id: ${config.bundleId}`)
  }
  assertExists(config.baseIpa, 'HBuilderX iPhone base IPA')

  if (!config.unsigned) {
    if (!config.identity) {
      throw new Error('Missing signing identity. Pass --identity or set IOS_SIGN_IDENTITY.')
    }
    if (!config.profile) {
      throw new Error('Missing provisioning profile. Pass --profile or set IOS_MOBILEPROVISION.')
    }
    assertExists(config.profile, 'Provisioning profile')
  }
}

async function main() {
  const config = resolveConfig(parseArgs(process.argv.slice(2)))

  if (config.dryRun) {
    printConfig(config)
    return
  }

  validateConfig(config)

  const tempRoot = config.tempDir
    ? resolvePath(config.projectRoot, config.tempDir)
    : fs.mkdtempSync(path.join(os.tmpdir(), 'wecheckin-ios-'))
  const stageRoot = path.join(tempRoot, 'ipa')

  try {
    await buildUniAppResources(config)
    const wwwDir = findUniWww(config)

    console.log('[2/6] Extracting HBuilderX iOS base IPA')
    fs.rmSync(stageRoot, { recursive: true, force: true })
    fs.mkdirSync(stageRoot, { recursive: true })
    await run('/usr/bin/unzip', ['-q', config.baseIpa, '-d', stageRoot])

    const appDir = path.join(stageRoot, 'Payload/HBuilder.app')
    assertExists(appDir, 'HBuilder.app in base IPA')

    console.log('[3/6] Injecting uni app resources')
    replaceAppResources(appDir, wwwDir, config)
    removeCodeSignatures(appDir)

    if (config.unsigned) {
      console.log('[4/6] Skipping code signing; unsigned IPA is not installable')
    } else {
      await signApp(config, tempRoot, appDir)
    }

    console.log('[5/6] Packaging IPA')
    await packageIpa(stageRoot, config.outputIpa)

    console.log('[6/6] Validating IPA archive')
    await run('/usr/bin/unzip', ['-tq', config.outputIpa])
    console.log(`Done: ${config.outputIpa}`)
  } finally {
    if (!config.keepTemp) {
      fs.rmSync(tempRoot, { recursive: true, force: true })
    } else {
      console.log(`Temporary files kept: ${tempRoot}`)
    }
  }
}

main().catch((error) => {
  console.error(`Error: ${error.message}`)
  process.exit(1)
})
