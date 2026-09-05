package workflowcore

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const maxCalculationExpressionLength = 1000

type calculationNodeKind uint8

const (
	calculationNumber calculationNodeKind = iota
	calculationReference
	calculationUnary
	calculationBinary
	calculationAggregate
)

type calculationNode struct {
	kind      calculationNodeKind
	value     float64
	reference string
	operator  byte
	function  string
	left      *calculationNode
	right     *calculationNode
	argument  *calculationNode
	detailKey string
}

type calculationParser struct {
	expression string
	position   int
}

type calculationDetailSchema struct {
	key     string
	columns map[string]FormField
}

type calculationSchema struct {
	fields  map[string]FormField
	details []calculationDetailSchema
}

type calculationEvaluationContext struct {
	data      map[string]interface{}
	detailKey string
	row       map[string]interface{}
}

func ApplyFormCalculations(fields []FormField, data map[string]interface{}) (map[string]interface{}, error) {
	result := MergeFormData(nil, data)
	schema := buildCalculationSchema(fields)
	for _, field := range dataFormFields(fields) {
		if field.Type != FormFieldTypeCalculation {
			continue
		}
		node, err := parseAndValidateCalculation(field, schema)
		if err != nil {
			return nil, fmt.Errorf("%w：%s计算公式无效：%v", ErrFormDataInvalid, field.Label, err)
		}
		value, err := evaluateCalculationNode(node, calculationEvaluationContext{data: result})
		if err != nil {
			return nil, fmt.Errorf("%w：%s计算失败：%v", ErrFormDataInvalid, field.Label, err)
		}
		value = roundCalculationValue(value, calculationPrecision(field.Calculation))
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("%w：%s计算结果超出有效数字范围", ErrFormDataInvalid, field.Label)
		}
		if value == 0 {
			value = 0
		}
		result[field.Key] = value
	}
	return result, nil
}

func roundCalculationValue(value float64, precision int) float64 {
	absolute := math.Abs(value)
	factor := math.Pow10(precision)
	rounded := math.Round(math.Nextafter(absolute, math.Inf(1))*factor) / factor
	if value < 0 {
		return -rounded
	}
	return rounded
}

func validateFormCalculation(field FormField, schema calculationSchema) error {
	_, err := parseAndValidateCalculation(field, schema)
	return err
}

func parseAndValidateCalculation(field FormField, schema calculationSchema) (*calculationNode, error) {
	if field.Calculation == nil {
		return nil, fmt.Errorf("缺少计算配置")
	}
	expression := strings.TrimSpace(field.Calculation.Expression)
	if expression == "" {
		return nil, fmt.Errorf("计算公式不能为空")
	}
	if len(expression) > maxCalculationExpressionLength {
		return nil, fmt.Errorf("计算公式不能超过%d个字符", maxCalculationExpressionLength)
	}
	display := strings.TrimSpace(field.Calculation.Display)
	if display != "" && display != CalculationDisplayLabel && display != CalculationDisplayField {
		return nil, fmt.Errorf("展示方式无效")
	}
	precision := calculationPrecision(field.Calculation)
	if precision < 0 || precision > 6 {
		return nil, fmt.Errorf("小数位数必须在0到6之间")
	}
	parser := calculationParser{expression: expression}
	node, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}
	parser.skipSpaces()
	if parser.position != len(parser.expression) {
		return nil, fmt.Errorf("位置%d存在无法识别的内容", parser.position+1)
	}
	if err := validateCalculationNode(node, schema, false, nil); err != nil {
		return nil, err
	}
	return node, nil
}

func calculationPrecision(config *FormCalculation) int {
	if config == nil || config.Precision == nil {
		return 2
	}
	return *config.Precision
}

func buildCalculationSchema(fields []FormField) calculationSchema {
	schema := calculationSchema{fields: make(map[string]FormField)}
	var appendFields func([]FormField)
	appendFields = func(source []FormField) {
		for _, field := range source {
			if field.Type == FormFieldTypeGroup {
				appendFields(field.Fields)
				continue
			}
			schema.fields[field.Key] = field
			if field.Type != FormFieldTypeDetailList {
				continue
			}
			columns := make(map[string]FormField, len(field.Columns))
			for _, column := range field.Columns {
				columns[column.Key] = column
			}
			schema.details = append(schema.details, calculationDetailSchema{key: field.Key, columns: columns})
		}
	}
	appendFields(fields)
	return schema
}

func validateCalculationNode(node *calculationNode, schema calculationSchema, inAggregate bool, aggregateDetail *string) error {
	if node == nil {
		return fmt.Errorf("计算表达式为空")
	}
	switch node.kind {
	case calculationNumber:
		return nil
	case calculationReference:
		if inAggregate {
			if detailKey, column, ok := resolveCalculationDetailReference(schema, node.reference); ok {
				if !isCalculationNumberField(column) {
					return fmt.Errorf("明细字段[%s]不是数字或金额", node.reference)
				}
				if *aggregateDetail != "" && *aggregateDetail != detailKey {
					return fmt.Errorf("一次明细聚合只能引用同一个明细列表")
				}
				*aggregateDetail = detailKey
				return nil
			}
		}
		field, ok := schema.fields[node.reference]
		if !ok {
			if _, _, detailReference := resolveCalculationDetailReference(schema, node.reference); detailReference {
				return fmt.Errorf("明细字段[%s]必须放在SUM、AVG、MIN、MAX或COUNT中", node.reference)
			}
			return fmt.Errorf("字段[%s]不存在", node.reference)
		}
		if !isCalculationNumberField(field) {
			return fmt.Errorf("字段[%s]不是数字或金额", node.reference)
		}
		return nil
	case calculationUnary:
		return validateCalculationNode(node.argument, schema, inAggregate, aggregateDetail)
	case calculationBinary:
		if err := validateCalculationNode(node.left, schema, inAggregate, aggregateDetail); err != nil {
			return err
		}
		return validateCalculationNode(node.right, schema, inAggregate, aggregateDetail)
	case calculationAggregate:
		if inAggregate {
			return fmt.Errorf("明细聚合函数不能嵌套")
		}
		detailKey := ""
		if err := validateCalculationNode(node.argument, schema, true, &detailKey); err != nil {
			return err
		}
		if detailKey == "" {
			return fmt.Errorf("%s函数必须引用明细列表中的数字列", node.function)
		}
		node.detailKey = detailKey
		return nil
	default:
		return fmt.Errorf("计算表达式节点无效")
	}
}

func resolveCalculationDetailReference(schema calculationSchema, reference string) (string, FormField, bool) {
	for _, detail := range schema.details {
		prefix := detail.key + "."
		if !strings.HasPrefix(reference, prefix) {
			continue
		}
		column, ok := detail.columns[strings.TrimPrefix(reference, prefix)]
		if ok {
			return detail.key, column, true
		}
	}
	return "", FormField{}, false
}

func isCalculationNumberField(field FormField) bool {
	return field.Type == FormFieldTypeNumber || field.Type == FormFieldTypeAmount
}

func evaluateCalculationNode(node *calculationNode, context calculationEvaluationContext) (float64, error) {
	switch node.kind {
	case calculationNumber:
		return node.value, nil
	case calculationReference:
		value := context.data[node.reference]
		if context.detailKey != "" {
			prefix := context.detailKey + "."
			if strings.HasPrefix(node.reference, prefix) {
				value = context.row[strings.TrimPrefix(node.reference, prefix)]
			}
		}
		return calculationNumberValue(value)
	case calculationUnary:
		value, err := evaluateCalculationNode(node.argument, context)
		if err != nil {
			return 0, err
		}
		if node.operator == '-' {
			return -value, nil
		}
		return value, nil
	case calculationBinary:
		left, err := evaluateCalculationNode(node.left, context)
		if err != nil {
			return 0, err
		}
		right, err := evaluateCalculationNode(node.right, context)
		if err != nil {
			return 0, err
		}
		var value float64
		switch node.operator {
		case '+':
			value = left + right
		case '-':
			value = left - right
		case '*':
			value = left * right
		case '/':
			if right == 0 {
				return 0, fmt.Errorf("不能除以0")
			}
			value = left / right
		default:
			return 0, fmt.Errorf("运算符无效")
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, fmt.Errorf("计算结果不是有效数字")
		}
		return value, nil
	case calculationAggregate:
		rows, ok := detailListRows(context.data[node.detailKey])
		if !ok {
			return 0, fmt.Errorf("明细列表[%s]数据无效", node.detailKey)
		}
		if node.function == "COUNT" {
			return float64(len(rows)), nil
		}
		values := make([]float64, 0, len(rows))
		for _, row := range rows {
			value, err := evaluateCalculationNode(node.argument, calculationEvaluationContext{data: context.data, detailKey: node.detailKey, row: row})
			if err != nil {
				return 0, err
			}
			values = append(values, value)
		}
		if len(values) == 0 {
			return 0, nil
		}
		result := values[0]
		switch node.function {
		case "SUM", "AVG":
			result = 0
			for _, value := range values {
				result += value
			}
			if node.function == "AVG" {
				result /= float64(len(values))
			}
		case "MIN":
			for _, value := range values[1:] {
				result = math.Min(result, value)
			}
		case "MAX":
			for _, value := range values[1:] {
				result = math.Max(result, value)
			}
		default:
			return 0, fmt.Errorf("聚合函数无效")
		}
		return result, nil
	default:
		return 0, fmt.Errorf("计算表达式节点无效")
	}
}

func calculationNumberValue(value interface{}) (float64, error) {
	if value == nil {
		return 0, nil
	}
	var number float64
	switch typed := value.(type) {
	case float64:
		number = typed
	case float32:
		number = float64(typed)
	case int:
		number = float64(typed)
	case int8:
		number = float64(typed)
	case int16:
		number = float64(typed)
	case int32:
		number = float64(typed)
	case int64:
		number = float64(typed)
	case uint:
		number = float64(typed)
	case uint8:
		number = float64(typed)
	case uint16:
		number = float64(typed)
	case uint32:
		number = float64(typed)
	case uint64:
		number = float64(typed)
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return 0, fmt.Errorf("%q不是有效数字", typed.String())
		}
		number = parsed
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return 0, nil
		}
		parsed, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return 0, fmt.Errorf("%q不是有效数字", typed)
		}
		number = parsed
	default:
		return 0, fmt.Errorf("字段值不是有效数字")
	}
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, fmt.Errorf("字段值不是有效数字")
	}
	return number, nil
}

func (parser *calculationParser) parseExpression() (*calculationNode, error) {
	left, err := parser.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		parser.skipSpaces()
		operator := parser.peek()
		if operator != '+' && operator != '-' {
			return left, nil
		}
		parser.position++
		right, err := parser.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &calculationNode{kind: calculationBinary, operator: operator, left: left, right: right}
	}
}

func (parser *calculationParser) parseTerm() (*calculationNode, error) {
	left, err := parser.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		parser.skipSpaces()
		operator := parser.peek()
		if operator != '*' && operator != '/' {
			return left, nil
		}
		parser.position++
		right, err := parser.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &calculationNode{kind: calculationBinary, operator: operator, left: left, right: right}
	}
}

func (parser *calculationParser) parseUnary() (*calculationNode, error) {
	parser.skipSpaces()
	operator := parser.peek()
	if operator != '+' && operator != '-' {
		return parser.parsePrimary()
	}
	parser.position++
	argument, err := parser.parseUnary()
	if err != nil {
		return nil, err
	}
	return &calculationNode{kind: calculationUnary, operator: operator, argument: argument}, nil
}

func (parser *calculationParser) parsePrimary() (*calculationNode, error) {
	parser.skipSpaces()
	if parser.position >= len(parser.expression) {
		return nil, fmt.Errorf("计算公式不完整")
	}
	switch current := parser.peek(); {
	case current == '(':
		parser.position++
		node, err := parser.parseExpression()
		if err != nil {
			return nil, err
		}
		parser.skipSpaces()
		if parser.peek() != ')' {
			return nil, fmt.Errorf("位置%d缺少右括号", parser.position+1)
		}
		parser.position++
		return node, nil
	case current == '[':
		return parser.parseReference()
	case current >= '0' && current <= '9' || current == '.':
		return parser.parseNumber()
	case unicode.IsLetter(rune(current)):
		return parser.parseFunction()
	default:
		return nil, fmt.Errorf("位置%d的字符无法识别", parser.position+1)
	}
}

func (parser *calculationParser) parseReference() (*calculationNode, error) {
	start := parser.position + 1
	end := strings.IndexByte(parser.expression[start:], ']')
	if end < 0 {
		return nil, fmt.Errorf("位置%d缺少右方括号", parser.position+1)
	}
	end += start
	reference := strings.TrimSpace(parser.expression[start:end])
	if reference == "" {
		return nil, fmt.Errorf("字段引用不能为空")
	}
	parser.position = end + 1
	return &calculationNode{kind: calculationReference, reference: reference}, nil
}

func (parser *calculationParser) parseNumber() (*calculationNode, error) {
	start := parser.position
	dotSeen := false
	for parser.position < len(parser.expression) {
		current := parser.peek()
		if current == '.' && !dotSeen {
			dotSeen = true
			parser.position++
			continue
		}
		if current < '0' || current > '9' {
			break
		}
		parser.position++
	}
	text := parser.expression[start:parser.position]
	value, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
		return nil, fmt.Errorf("%q不是有效数字", text)
	}
	return &calculationNode{kind: calculationNumber, value: value}, nil
}

func (parser *calculationParser) parseFunction() (*calculationNode, error) {
	start := parser.position
	for parser.position < len(parser.expression) && unicode.IsLetter(rune(parser.peek())) {
		parser.position++
	}
	name := strings.ToUpper(parser.expression[start:parser.position])
	switch name {
	case "SUM", "AVG", "MIN", "MAX", "COUNT":
	default:
		return nil, fmt.Errorf("不支持函数%s", name)
	}
	parser.skipSpaces()
	if parser.peek() != '(' {
		return nil, fmt.Errorf("函数%s后缺少左括号", name)
	}
	parser.position++
	argument, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}
	parser.skipSpaces()
	if parser.peek() != ')' {
		return nil, fmt.Errorf("函数%s缺少右括号", name)
	}
	parser.position++
	return &calculationNode{kind: calculationAggregate, function: name, argument: argument}, nil
}

func (parser *calculationParser) skipSpaces() {
	for parser.position < len(parser.expression) && unicode.IsSpace(rune(parser.expression[parser.position])) {
		parser.position++
	}
}

func (parser *calculationParser) peek() byte {
	if parser.position >= len(parser.expression) {
		return 0
	}
	return parser.expression[parser.position]
}
