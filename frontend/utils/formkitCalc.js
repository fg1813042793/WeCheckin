// Lightweight formkit expression runtime for mobile rendering.
// It intentionally avoids eval/Function; backend submit remains the final authority.

const TOKEN = {
  eof: 'eof',
  number: 'number',
  string: 'string',
  ident: 'ident',
  bool: 'bool',
  null: 'null',
  op: 'op',
  lparen: 'lparen',
  rparen: 'rparen',
  comma: 'comma'
}

function isSpace(ch) {
  return /\s/.test(ch)
}

function isDigit(ch) {
  return ch >= '0' && ch <= '9'
}

function isIdentStart(ch) {
  return ch === '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

function isIdentPart(ch) {
  return isIdentStart(ch) || isDigit(ch)
}

function tokenize(expr) {
  const tokens = []
  let i = 0
  while (i < expr.length) {
    const ch = expr[i]
    if (isSpace(ch)) {
      i += 1
      continue
    }

    if (isDigit(ch) || (ch === '.' && isDigit(expr[i + 1] || ''))) {
      const start = i
      let hasDot = false
      while (i < expr.length) {
        const c = expr[i]
        if (isDigit(c)) {
          i += 1
        } else if (c === '.' && !hasDot) {
          hasDot = true
          i += 1
        } else {
          break
        }
      }
      tokens.push({ type: TOKEN.number, value: expr.slice(start, i), pos: start })
      continue
    }

    if (ch === '"' || ch === "'") {
      const quote = ch
      const start = i
      i += 1
      let value = ''
      while (i < expr.length && expr[i] !== quote) {
        if (expr[i] === '\\' && i + 1 < expr.length) {
          value += expr[i + 1]
          i += 2
        } else {
          value += expr[i]
          i += 1
        }
      }
      if (i >= expr.length) throw new Error(`unterminated string at ${start}`)
      i += 1
      tokens.push({ type: TOKEN.string, value, pos: start })
      continue
    }

    if (isIdentStart(ch)) {
      const start = i
      i += 1
      while (i < expr.length && isIdentPart(expr[i])) i += 1
      const word = expr.slice(start, i)
      const lower = word.toLowerCase()
      if (lower === 'true' || lower === 'false') {
        tokens.push({ type: TOKEN.bool, value: lower === 'true', pos: start })
      } else if (lower === 'null' || lower === 'nil') {
        tokens.push({ type: TOKEN.null, value: null, pos: start })
      } else {
        tokens.push({ type: TOKEN.ident, value: word, pos: start })
      }
      continue
    }

    const two = expr.slice(i, i + 2)
    if (['==', '!=', '<=', '>=', '&&', '||'].includes(two)) {
      tokens.push({ type: TOKEN.op, value: two, pos: i })
      i += 2
      continue
    }
    if ('+-*/%=!<>?:'.includes(ch)) {
      tokens.push({ type: TOKEN.op, value: ch, pos: i })
      i += 1
      continue
    }
    if (ch === '(') {
      tokens.push({ type: TOKEN.lparen, value: ch, pos: i })
      i += 1
      continue
    }
    if (ch === ')') {
      tokens.push({ type: TOKEN.rparen, value: ch, pos: i })
      i += 1
      continue
    }
    if (ch === ',') {
      tokens.push({ type: TOKEN.comma, value: ch, pos: i })
      i += 1
      continue
    }
    throw new Error(`unexpected char ${ch} at ${i}`)
  }
  tokens.push({ type: TOKEN.eof, value: '', pos: expr.length })
  return tokens
}

class Parser {
  constructor(tokens) {
    this.tokens = tokens
    this.pos = 0
  }

  peek() {
    return this.tokens[this.pos]
  }

  matchOp(op) {
    if (this.peek().type === TOKEN.op && this.peek().value === op) {
      this.pos += 1
      return true
    }
    return false
  }

  expectOp(op) {
    if (!this.matchOp(op)) {
      const token = this.peek()
      throw new Error(`expected ${op} at ${token.pos}`)
    }
  }

  parseExpression() {
    return this.parseTernary()
  }

  parseTernary() {
    const condition = this.parseOr()
    if (!this.matchOp('?')) return condition
    const yes = this.parseExpression()
    this.expectOp(':')
    const no = this.parseExpression()
    return { type: 'ternary', condition, yes, no }
  }

  parseOr() {
    let left = this.parseAnd()
    while (this.matchOp('||')) left = { type: 'binary', op: '||', left, right: this.parseAnd() }
    return left
  }

  parseAnd() {
    let left = this.parseEq()
    while (this.matchOp('&&')) left = { type: 'binary', op: '&&', left, right: this.parseEq() }
    return left
  }

  parseEq() {
    let left = this.parseCmp()
    while (true) {
      if (this.matchOp('==')) left = { type: 'binary', op: '==', left, right: this.parseCmp() }
      else if (this.matchOp('!=')) left = { type: 'binary', op: '!=', left, right: this.parseCmp() }
      else return left
    }
  }

  parseCmp() {
    let left = this.parseAdd()
    while (true) {
      const token = this.peek()
      if (token.type === TOKEN.op && ['<', '>', '<=', '>='].includes(token.value)) {
        this.pos += 1
        left = { type: 'binary', op: token.value, left, right: this.parseAdd() }
      } else {
        return left
      }
    }
  }

  parseAdd() {
    let left = this.parseMul()
    while (true) {
      if (this.matchOp('+')) left = { type: 'binary', op: '+', left, right: this.parseMul() }
      else if (this.matchOp('-')) left = { type: 'binary', op: '-', left, right: this.parseMul() }
      else return left
    }
  }

  parseMul() {
    let left = this.parseUnary()
    while (true) {
      if (this.matchOp('*')) left = { type: 'binary', op: '*', left, right: this.parseUnary() }
      else if (this.matchOp('/')) left = { type: 'binary', op: '/', left, right: this.parseUnary() }
      else if (this.matchOp('%')) left = { type: 'binary', op: '%', left, right: this.parseUnary() }
      else return left
    }
  }

  parseUnary() {
    if (this.matchOp('-')) return { type: 'unary', op: '-', expr: this.parseUnary() }
    if (this.matchOp('!')) return { type: 'unary', op: '!', expr: this.parseUnary() }
    return this.parsePrimary()
  }

  parsePrimary() {
    const token = this.peek()
    if (token.type === TOKEN.number) {
      this.pos += 1
      return { type: 'number', value: Number(token.value) }
    }
    if (token.type === TOKEN.string || token.type === TOKEN.bool || token.type === TOKEN.null) {
      this.pos += 1
      return { type: token.type, value: token.value }
    }
    if (token.type === TOKEN.ident) {
      this.pos += 1
      if (this.peek().type === TOKEN.lparen) {
        this.pos += 1
        const args = this.parseArgs()
        return { type: 'call', name: token.value, args }
      }
      return { type: 'ident', name: token.value }
    }
    if (token.type === TOKEN.lparen) {
      this.pos += 1
      const expr = this.parseExpression()
      if (this.peek().type !== TOKEN.rparen) throw new Error(`expected ) at ${this.peek().pos}`)
      this.pos += 1
      return expr
    }
    throw new Error(`unexpected token ${token.value} at ${token.pos}`)
  }

  parseArgs() {
    if (this.peek().type === TOKEN.rparen) {
      this.pos += 1
      return []
    }
    const args = []
    while (true) {
      args.push(this.parseExpression())
      if (this.peek().type === TOKEN.comma) {
        this.pos += 1
        continue
      }
      break
    }
    if (this.peek().type !== TOKEN.rparen) throw new Error(`expected ) or , at ${this.peek().pos}`)
    this.pos += 1
    return args
  }
}

function toNumber(value) {
  if (value === null || value === undefined) return { ok: true, value: 0 }
  if (typeof value === 'number') return Number.isFinite(value) ? { ok: true, value } : { ok: false, value: 0 }
  if (typeof value === 'boolean') return { ok: true, value: value ? 1 : 0 }
  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (!trimmed) return { ok: false, value: 0 }
    const n = Number(trimmed)
    return Number.isFinite(n) ? { ok: true, value: n } : { ok: false, value: 0 }
  }
  return { ok: false, value: 0 }
}

function toText(value) {
  if (value === null || value === undefined) return ''
  if (typeof value === 'string') return value
  if (typeof value === 'number') return Number.isInteger(value) ? String(value) : String(value)
  if (typeof value === 'boolean') return value ? 'true' : 'false'
  return String(value)
}

function toBool(value) {
  if (value === null || value === undefined) return false
  if (typeof value === 'boolean') return value
  if (typeof value === 'number') return value !== 0
  if (typeof value === 'string') {
    const s = value.trim().toLowerCase()
    return s !== '' && s !== 'false' && s !== '0'
  }
  return true
}

function isEmpty(value) {
  if (value === null || value === undefined) return true
  if (typeof value === 'string') return value.trim() === ''
  if (Array.isArray(value)) return value.length === 0
  if (typeof value === 'object') return Object.keys(value).length === 0
  return false
}

function lengthOf(value) {
  if (typeof value === 'string' || Array.isArray(value)) return value.length
  if (value && typeof value === 'object') return Object.keys(value).length
  return 0
}

function compareValues(left, right) {
  const lnum = toNumber(left)
  const rnum = toNumber(right)
  if (lnum.ok && rnum.ok) {
    if (lnum.value < rnum.value) return -1
    if (lnum.value > rnum.value) return 1
    return 0
  }
  const ltext = toText(left)
  const rtext = toText(right)
  if (ltext < rtext) return -1
  if (ltext > rtext) return 1
  return 0
}

function equalValues(left, right) {
  const lnum = toNumber(left)
  const rnum = toNumber(right)
  if (lnum.ok && rnum.ok) return lnum.value === rnum.value
  return toText(left) === toText(right)
}

function evalBinary(op, left, right) {
  if (op === '+') {
    if (typeof left === 'string' || typeof right === 'string') return toText(left) + toText(right)
    const lnum = toNumber(left)
    const rnum = toNumber(right)
    if (!lnum.ok || !rnum.ok) throw new Error('cannot add non-number')
    return lnum.value + rnum.value
  }
  if (['-', '*', '/', '%'].includes(op)) {
    const lnum = toNumber(left)
    const rnum = toNumber(right)
    if (!lnum.ok || !rnum.ok) throw new Error(`cannot ${op} non-number`)
    if (op === '-') return lnum.value - rnum.value
    if (op === '*') return lnum.value * rnum.value
    if (op === '/') {
      if (rnum.value === 0) throw new Error('division by zero')
      return lnum.value / rnum.value
    }
    if (rnum.value === 0) throw new Error('modulo by zero')
    return Math.trunc(lnum.value) % Math.trunc(rnum.value)
  }
  if (['<', '>', '<=', '>='].includes(op)) {
    const cmp = compareValues(left, right)
    if (op === '<') return cmp < 0
    if (op === '>') return cmp > 0
    if (op === '<=') return cmp <= 0
    return cmp >= 0
  }
  if (op === '==') return equalValues(left, right)
  if (op === '!=') return !equalValues(left, right)
  if (op === '&&') return toBool(left) && toBool(right)
  if (op === '||') return toBool(left) || toBool(right)
  throw new Error(`unsupported operator ${op}`)
}

function evalCall(name, argNodes, env) {
  const upperName = name.toUpperCase()
  const args = argNodes.map((node) => evalNode(node, env))
  switch (upperName) {
    case 'CONTAINS':
      return args.length === 2 ? toText(args[0]).includes(toText(args[1])) : false
    case 'EMPTY':
    case 'ISBLANK':
      return args.length === 1 ? isEmpty(args[0]) : false
    case 'LEN':
      return args.length === 1 ? lengthOf(args[0]) : 0
    case 'IF':
      return args.length >= 3 ? (toBool(args[0]) ? args[1] : args[2]) : null
    case 'SUM':
      return args.reduce((sum, value) => {
        const n = toNumber(value)
        if (!n.ok) throw new Error('SUM expects numbers')
        return sum + n.value
      }, 0)
    case 'AVG':
      if (!args.length) throw new Error('AVG expects args')
      return evalCall('SUM', argNodes, env) / args.length
    case 'MIN':
      return Math.min(...args.map((value) => {
        const n = toNumber(value)
        if (!n.ok) throw new Error('MIN expects numbers')
        return n.value
      }))
    case 'MAX':
      return Math.max(...args.map((value) => {
        const n = toNumber(value)
        if (!n.ok) throw new Error('MAX expects numbers')
        return n.value
      }))
    case 'CONCAT':
    case 'CONCATENATE':
      return args.map(toText).join('')
    case 'AND':
      return args.every(toBool)
    case 'OR':
      return args.some(toBool)
    case 'NOT':
      return args.length ? !toBool(args[0]) : true
    case 'IFS':
      for (let i = 0; i + 1 < args.length; i += 2) {
        if (toBool(args[i])) return args[i + 1]
      }
      return args.length % 2 === 1 ? args[args.length - 1] : null
    default:
      throw new Error(`unknown function ${name}`)
  }
}

function evalNode(node, env) {
  switch (node.type) {
    case 'number':
    case 'string':
    case 'bool':
    case 'null':
      return node.value
    case 'ident':
      return env ? env[node.name] : undefined
    case 'unary': {
      const value = evalNode(node.expr, env)
      if (node.op === '!') return !toBool(value)
      const n = toNumber(value)
      if (!n.ok) throw new Error('cannot negate non-number')
      return -n.value
    }
    case 'binary':
      return evalBinary(node.op, evalNode(node.left, env), evalNode(node.right, env))
    case 'ternary':
      return toBool(evalNode(node.condition, env))
        ? evalNode(node.yes, env)
        : evalNode(node.no, env)
    case 'call':
      return evalCall(node.name, node.args, env)
    default:
      throw new Error(`unsupported node ${node.type}`)
  }
}

export function normalizeExpression(expr) {
  return String(expr || '').replace(/#\{([A-Za-z_][A-Za-z0-9_]*)\}/g, '$1').trim()
}

export function evalFormkitExpression(expr, env = {}) {
  const normalized = normalizeExpression(expr)
  if (!normalized) return null
  const parser = new Parser(tokenize(normalized))
  const ast = parser.parseExpression()
  if (parser.peek().type !== TOKEN.eof) throw new Error(`trailing token at ${parser.peek().pos}`)
  return evalNode(ast, env)
}

export function buildCalcEnv(questions = [], answers = {}) {
  const env = {}
  questions.forEach((question, index) => {
    const value = answers ? answers[question.id] : undefined
    env[question.id] = value
    env[`Q${index + 1}`] = value
    env[`q${index + 1}`] = value
  })
  return env
}

function getCalcConfig(question) {
  if (!question) return null
  if (typeof question.calcValue === 'string') return question.calcValue ? { expr: question.calcValue } : null
  if (question.calcValue && question.calcValue.expr) return question.calcValue
  if (question.props && question.props.calculateFormula) return { expr: question.props.calculateFormula }
  return null
}

export function getCalculatedQuestionIds(questions = []) {
  const ids = new Set()
  questions.forEach((question) => {
    const calc = getCalcConfig(question)
    if (!calc || !calc.expr) return
    ids.add(calc.target || question.id)
  })
  return ids
}

export function applyFormkitCalcValues(questions = [], answers = {}) {
  if (!answers || Array.isArray(answers)) return answers
  const next = { ...answers }
  questions.forEach((question) => {
    const calc = getCalcConfig(question)
    if (!calc || !calc.expr) return
    try {
      const value = evalFormkitExpression(calc.expr, buildCalcEnv(questions, next))
      next[calc.target || question.id] = value
    } catch (e) {
      // Keep current answer on invalid formulas; backend submit recalculates as final guard.
    }
  })
  return next
}
