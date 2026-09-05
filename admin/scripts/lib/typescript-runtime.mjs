import vm from 'node:vm'
import ts from 'typescript'

export function executeTypeScriptModule(source, fileName, modules = {}) {
  const result = ts.transpileModule(source, {
    fileName,
    reportDiagnostics: true,
    compilerOptions: {
      esModuleInterop: true,
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2020,
    },
  })
  const errors = (result.diagnostics || []).filter(diagnostic => diagnostic.category === ts.DiagnosticCategory.Error)
  if (errors.length > 0) {
    throw new Error(ts.formatDiagnostics(errors, {
      getCanonicalFileName: name => name,
      getCurrentDirectory: () => process.cwd(),
      getNewLine: () => '\n',
    }))
  }

  const module = { exports: {} }
  const requireModule = (specifier) => {
    if (Object.prototype.hasOwnProperty.call(modules, specifier)) return modules[specifier]
    throw new Error(`unexpected runtime import in ${fileName}: ${specifier}`)
  }
  const context = vm.createContext({
    console,
    exports: module.exports,
    module,
    require: requireModule,
  })
  new vm.Script(result.outputText, { filename: fileName }).runInContext(context)
  return module.exports
}
