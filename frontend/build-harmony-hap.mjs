#!/usr/bin/env node
import { spawn } from 'node:child_process'
import { createRequire } from 'node:module'
import { fileURLToPath } from 'node:url'
import crypto from 'node:crypto'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'

const require = createRequire(import.meta.url)
const scriptDir = path.dirname(fileURLToPath(import.meta.url))

function printHelp() {
  console.log(`Usage:
  node build-harmony-hap.mjs [options]

Builds a signed Harmony/OpenHarmony HAP from this uni-app project.

Options:
  --bundle <name>              Bundle name, default from manifest harmony.package/app-harmony.
  --output <file>              Output HAP path.
  --deveco <path>              DevEco Studio app path.
  --hbuilderx <path>           HBuilderX app path.
  --compatible-version <num>   HAP compatible API version, default 12.
  --valid-days <num>           Generated profile validity in days, default 365.
  --temp <dir>                 Temporary work directory.
  --skip-uni-build             Reuse unpackage/dist/build/.app-harmony.
  --install <target>           Install generated HAP with hdc, for example 127.0.0.1:5555.
  --run                        Start EntryAbility after install.
  --keep-temp                  Keep temporary Harmony project/signing files.
  --dry-run                    Print resolved configuration and exit.
  --help                       Show this help.

Environment:
  HARMONY_BUNDLE_NAME, HARMONY_OUTPUT_HAP, DEVECO_STUDIO_PATH, HBUILDERX_PATH,
  HARMONY_COMPATIBLE_VERSION, HARMONY_PROFILE_VALID_DAYS, HARMONY_TEMP_DIR,
  HARMONY_INSTALL_TARGET
`)
}

function parseArgs(argv) {
  const options = {
    projectRoot: scriptDir,
    hbuilderxPath: process.env.HBUILDERX_PATH || '/Applications/HBuilderX.app',
    devEcoPath: process.env.DEVECO_STUDIO_PATH || '/Applications/DevEco-Studio.app',
    bundleName: process.env.HARMONY_BUNDLE_NAME || '',
    outputHap: process.env.HARMONY_OUTPUT_HAP || '',
    compatibleVersion: process.env.HARMONY_COMPATIBLE_VERSION || '12',
    profileValidDays: Number(process.env.HARMONY_PROFILE_VALID_DAYS || 365),
    tempDir: process.env.HARMONY_TEMP_DIR || '',
    installTarget: process.env.HARMONY_INSTALL_TARGET || '',
    runAfterInstall: false,
    skipUniBuild: false,
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
    } else if (arg === '--bundle') {
      options.bundleName = next()
    } else if (arg === '--output') {
      options.outputHap = next()
    } else if (arg === '--deveco') {
      options.devEcoPath = next()
    } else if (arg === '--hbuilderx') {
      options.hbuilderxPath = next()
    } else if (arg === '--compatible-version') {
      options.compatibleVersion = next()
    } else if (arg === '--valid-days') {
      options.profileValidDays = Number(next())
    } else if (arg === '--temp') {
      options.tempDir = next()
    } else if (arg === '--install') {
      options.installTarget = next()
    } else if (arg === '--run') {
      options.runAfterInstall = true
    } else if (arg === '--skip-uni-build') {
      options.skipUniBuild = true
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

function writeJson(file, value) {
  fs.mkdirSync(path.dirname(file), { recursive: true })
  fs.writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`)
}

function assertExists(file, label) {
  if (!fs.existsSync(file)) {
    throw new Error(`${label} not found: ${file}`)
  }
}

function isValidBundleName(value) {
  if (typeof value !== 'string') return false
  if (value.length < 7 || value.length > 128) return false
  if (!/^[a-zA-Z]([0-9a-zA-Z_]*[0-9a-zA-Z])?(\.[0-9a-zA-Z]([0-9a-zA-Z_]*[0-9a-zA-Z])?){2,}$/.test(value)) {
    return false
  }
  const reserved = new Set(['oh', 'ohos', 'harmony', 'harmonyos', 'openharmony', 'system'])
  return value.split('.').every((part) => !reserved.has(part))
}

function deriveBundleName(manifest) {
  const appHarmonyName = manifest?.['app-harmony']?.distribute?.bundleName
  const harmonyName = manifest?.harmony?.package
  if (isValidBundleName(appHarmonyName)) return appHarmonyName
  if (isValidBundleName(harmonyName)) return harmonyName
  return 'com.wecheckin.app'
}

function resolveOutputPath(options, packageJson) {
  if (options.outputHap) {
    return path.isAbsolute(options.outputHap)
      ? options.outputHap
      : path.resolve(options.projectRoot, options.outputHap)
  }

  const packageName = String(packageJson.name || 'app').replace(/[^0-9A-Za-z_.-]/g, '-')
  return path.join(options.projectRoot, 'unpackage/dist/build', `${packageName}-openharmony-release-signed.hap`)
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

function extractPemBlocks(text) {
  return text.match(/-----BEGIN CERTIFICATE-----[\s\S]*?-----END CERTIFICATE-----/g) || []
}

function createTempManifestProject(options, manifest, tempRoot) {
  const tempProject = path.join(tempRoot, 'project-manifest')
  const tempManifest = structuredClone(manifest)

  tempManifest.harmony = tempManifest.harmony || {}
  tempManifest.harmony.package = options.bundleName
  tempManifest['app-harmony'] = tempManifest['app-harmony'] || {}
  tempManifest['app-harmony'].distribute = tempManifest['app-harmony'].distribute || {}
  tempManifest['app-harmony'].distribute.bundleName = options.bundleName
  tempManifest['app-harmony'].distribute.signingConfigs =
    tempManifest['app-harmony'].distribute.signingConfigs || {}

  fs.mkdirSync(tempProject, { recursive: true })
  writeJson(path.join(tempProject, 'manifest.json'), tempManifest)
  return tempProject
}

async function buildUniHarmonyResources(options) {
  if (options.skipUniBuild) {
    console.log('[1/5] Reusing existing uni app-harmony resources')
    return
  }

  console.log('[1/5] Building uni app-harmony resources')
  const uniBin = path.join(options.projectRoot, 'node_modules/.bin/uni')
  assertExists(uniBin, 'uni CLI')
  await run(uniBin, ['build', '--platform', 'app-harmony'], {
    cwd: options.projectRoot,
    env: { ...process.env, UNI_INPUT_DIR: '.' }
  })
}

async function assembleUnsignedHap(options, tempProject, compilePath) {
  console.log('[2/5] Assembling Harmony project and unsigned HAP')
  const launcherFile = path.join(
    options.hbuilderxPath,
    'Contents/HBuilderX/plugins/launcher/out/export/HarmonyLauncher.js'
  )
  assertExists(launcherFile, 'HBuilderX HarmonyLauncher')

  const { HarmonyLauncher } = require(launcherFile)
  const launcher = new HarmonyLauncher()
  await launcher.init({
    devToolsPath: options.devEcoPath,
    projectPath: tempProject,
    compilePath,
    udid: ''
  })
  await launcher.prepareHarmonyProjectPath()
  await launcher.assembleCompileOutput()
  await launcher.assembleCompiledManifest()
  await launcher.assembleSigningConfigs()
  await launcher.installDependencies()
  await launcher.buildHap()

  assertExists(launcher.hapPackagePath, 'Unsigned HAP')
  return launcher.hapPackagePath
}

async function signHap(options, unsignedHap, tempRoot, outputHap) {
  console.log('[3/5] Signing HAP with OpenHarmony SDK certificate')
  const sdkLib = path.join(options.devEcoPath, 'Contents/sdk/default/openharmony/toolchains/lib')
  const java = path.join(options.devEcoPath, 'Contents/jbr/Contents/Home/bin/java')
  const signTool = path.join(sdkLib, 'hap-sign-tool.jar')
  const openHarmonyP12 = path.join(sdkLib, 'OpenHarmony.p12')
  const releaseProfilePem = path.join(sdkLib, 'OpenHarmonyProfileRelease.pem')
  const releaseProfileTemplate = path.join(sdkLib, 'UnsgnedReleasedProfileTemplate.json')

  assertExists(java, 'DevEco Java')
  assertExists(signTool, 'hap-sign-tool')
  assertExists(openHarmonyP12, 'OpenHarmony.p12')
  assertExists(releaseProfilePem, 'OpenHarmonyProfileRelease.pem')
  assertExists(releaseProfileTemplate, 'UnsgnedReleasedProfileTemplate.json')

  const signingDir = path.join(tempRoot, 'signing')
  fs.mkdirSync(signingDir, { recursive: true })

  const profileTemplate = readJson(releaseProfileTemplate)
  const profile = structuredClone(profileTemplate)
  const now = Math.floor(Date.now() / 1000)
  profile.uuid = crypto.randomUUID()
  profile.validity = {
    'not-before': now - 3600,
    'not-after': now + (options.profileValidDays * 24 * 60 * 60)
  }
  profile['bundle-info']['bundle-name'] = options.bundleName

  const profileJson = path.join(signingDir, 'release-profile.json')
  const profileP7b = path.join(signingDir, 'release-profile.p7b')
  writeJson(profileJson, profile)

  await run(java, [
    '-jar', signTool,
    'sign-profile',
    '-mode', 'localSign',
    '-keyAlias', 'openharmony application profile release',
    '-keyPwd', '123456',
    '-profileCertFile', releaseProfilePem,
    '-inFile', profileJson,
    '-signAlg', 'SHA256withECDSA',
    '-keystoreFile', openHarmonyP12,
    '-keystorePwd', '123456',
    '-outFile', profileP7b
  ])

  const appReleaseLeaf = profileTemplate['bundle-info']['distribution-certificate'].trim()
  const profileCertBlocks = extractPemBlocks(fs.readFileSync(releaseProfilePem, 'utf8'))
  if (!appReleaseLeaf.includes('BEGIN CERTIFICATE') || profileCertBlocks.length < 2) {
    throw new Error('Cannot build OpenHarmony app certificate chain from SDK files')
  }

  const appCertChain = path.join(signingDir, 'openharmony-app-release-chain.cer')
  fs.writeFileSync(appCertChain, `${[appReleaseLeaf, profileCertBlocks[1], profileCertBlocks[0]].join('\n')}\n`)

  fs.mkdirSync(path.dirname(outputHap), { recursive: true })
  fs.rmSync(outputHap, { force: true })

  await run(java, [
    '-jar', signTool,
    'sign-app',
    '-mode', 'localSign',
    '-keyAlias', 'openharmony application release',
    '-keyPwd', '123456',
    '-appCertFile', appCertChain,
    '-profileFile', profileP7b,
    '-inFile', unsignedHap,
    '-signAlg', 'SHA256withECDSA',
    '-keystoreFile', openHarmonyP12,
    '-keystorePwd', '123456',
    '-outFile', outputHap,
    '-compatibleVersion', String(options.compatibleVersion),
    '-signCode', '1'
  ])

  return { java, signTool }
}

async function verifyHap(signingTools, outputHap, tempRoot) {
  console.log('[4/5] Verifying signed HAP')
  const verifyDir = path.join(tempRoot, 'verify')
  fs.mkdirSync(verifyDir, { recursive: true })
  await run(signingTools.java, [
    '-jar', signingTools.signTool,
    'verify-app',
    '-inFile', outputHap,
    '-outCertChain', path.join(verifyDir, 'cert-chain.cer'),
    '-outProfile', path.join(verifyDir, 'profile.p7b')
  ])

  await run('unzip', ['-t', outputHap])
}

async function installHap(options, outputHap) {
  if (!options.installTarget) return

  console.log('[5/5] Installing HAP with hdc')
  const hdc = path.join(options.devEcoPath, 'Contents/sdk/default/openharmony/toolchains/hdc')
  assertExists(hdc, 'hdc')
  await run(hdc, ['-t', options.installTarget, 'install', '-r', outputHap])

  if (options.runAfterInstall) {
    await run(hdc, [
      '-t', options.installTarget,
      'shell', 'aa', 'start',
      '-a', 'EntryAbility',
      '-b', options.bundleName
    ])
  }
}

function printConfig(options, compilePath, outputHap) {
  console.log('Resolved Harmony HAP build config:')
  console.log(`  projectRoot: ${options.projectRoot}`)
  console.log(`  bundleName: ${options.bundleName}`)
  console.log(`  hbuilderxPath: ${options.hbuilderxPath}`)
  console.log(`  devEcoPath: ${options.devEcoPath}`)
  console.log(`  compilePath: ${compilePath}`)
  console.log(`  outputHap: ${outputHap}`)
  console.log(`  compatibleVersion: ${options.compatibleVersion}`)
  console.log(`  profileValidDays: ${options.profileValidDays}`)
  console.log(`  installTarget: ${options.installTarget || '(none)'}`)
}

async function main() {
  const options = parseArgs(process.argv.slice(2))
  const manifestFile = path.join(options.projectRoot, 'manifest.json')
  const packageFile = path.join(options.projectRoot, 'package.json')
  assertExists(manifestFile, 'manifest.json')
  assertExists(packageFile, 'package.json')
  assertExists(options.hbuilderxPath, 'HBuilderX')
  assertExists(options.devEcoPath, 'DevEco Studio')

  const manifest = readJson(manifestFile)
  const packageJson = readJson(packageFile)
  options.bundleName = options.bundleName || deriveBundleName(manifest)
  if (!isValidBundleName(options.bundleName)) {
    throw new Error(`Invalid Harmony bundle name: ${options.bundleName}`)
  }
  if (!Number.isFinite(options.profileValidDays) || options.profileValidDays < 1) {
    throw new Error(`Invalid --valid-days value: ${options.profileValidDays}`)
  }

  const compilePath = path.join(options.projectRoot, 'unpackage/dist/build/.app-harmony')
  const outputHap = resolveOutputPath(options, packageJson)
  printConfig(options, compilePath, outputHap)
  if (options.dryRun) return

  const tempRoot = options.tempDir
    ? path.resolve(options.tempDir)
    : fs.mkdtempSync(path.join(os.tmpdir(), 'wecheckin-harmony-'))

  let success = false
  try {
    await buildUniHarmonyResources(options)
    assertExists(compilePath, 'uni app-harmony resources')
    const tempProject = createTempManifestProject(options, manifest, tempRoot)
    const unsignedHap = await assembleUnsignedHap(options, tempProject, compilePath)
    const signingTools = await signHap(options, unsignedHap, tempRoot, outputHap)
    await verifyHap(signingTools, outputHap, tempRoot)
    await installHap(options, outputHap)
    const stats = fs.statSync(outputHap)
    console.log(`HAP ready: ${outputHap}`)
    console.log(`HAP size: ${(stats.size / 1024 / 1024).toFixed(1)} MB`)
    success = true
  } finally {
    if (success && !options.keepTemp && !options.tempDir) {
      fs.rmSync(tempRoot, { recursive: true, force: true })
    } else {
      console.log(`Temporary files kept at: ${tempRoot}`)
    }
  }
}

main().catch((error) => {
  console.error(error?.stack || error)
  process.exit(1)
})
