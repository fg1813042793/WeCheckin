package workflowcore

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

const (
	bpmnNamespace     = "http://www.omg.org/spec/BPMN/20100524/MODEL"
	xsiNamespace      = "http://www.w3.org/2001/XMLSchema-instance"
	flowableNamespace = "http://flowable.org/bpmn"
)

func CompileBPMN(definition Definition) ([]byte, error) {
	if errors := ValidateDefinition(definition); len(errors) > 0 {
		return nil, ValidationErrors(errors)
	}

	var output bytes.Buffer
	encoder := xml.NewEncoder(&output)
	encoder.Indent("", "  ")
	if err := encoder.EncodeToken(xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0" encoding="UTF-8"`)}); err != nil {
		return nil, err
	}
	definitions := startElement("definitions",
		attr("xmlns", bpmnNamespace),
		attr("xmlns:xsi", xsiNamespace),
		attr("xmlns:flowable", flowableNamespace),
		attr("targetNamespace", "https://wecheckin.local/workflow"),
	)
	if err := encoder.EncodeToken(definitions); err != nil {
		return nil, err
	}
	process := startElement("process",
		attr("id", definition.Key),
		attr("name", definition.Name),
		attr("isExecutable", "true"),
	)
	if err := encoder.EncodeToken(process); err != nil {
		return nil, err
	}

	defaultFlows := defaultFlowBySource(definition.Edges)
	for _, node := range definition.Nodes {
		if err := encodeNode(encoder, node, defaultFlows[node.ID]); err != nil {
			return nil, err
		}
	}
	for _, edge := range definition.Edges {
		if err := encodeEdge(encoder, edge); err != nil {
			return nil, err
		}
	}
	if err := encoder.EncodeToken(process.End()); err != nil {
		return nil, err
	}
	if err := encoder.EncodeToken(definitions.End()); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func encodeNode(encoder *xml.Encoder, node Node, defaultFlow string) error {
	attributes := []xml.Attr{attr("id", node.ID), attr("name", node.Name)}
	attributes = append(attributes, notificationAttributes(node.Notification)...)
	attributes = append(attributes, resultNotificationAttributes(node.ResultNotification)...)
	switch node.Type {
	case NodeTypeStart:
		return encodeEmpty(encoder, startElement("startEvent", attributes...))
	case NodeTypeEnd:
		return encodeEmpty(encoder, startElement("endEvent", attributes...))
	case NodeTypeExclusive:
		if defaultFlow != "" {
			attributes = append(attributes, attr("default", defaultFlow))
		}
		return encodeEmpty(encoder, startElement("exclusiveGateway", attributes...))
	case NodeTypeParallel:
		return encodeEmpty(encoder, startElement("parallelGateway", attributes...))
	case NodeTypeApproval:
		return encodeApproval(encoder, node, attributes)
	case NodeTypeHandle:
		key, value := singleAssigneeAttribute(node.ID, *node.Assignee)
		attributes = append(attributes, attr(key, value))
		return encodeEmpty(encoder, startElement("userTask", attributes...))
	case NodeTypeCC:
		attributes = append(attributes, attr("flowable:type", "external-worker"), attr("flowable:topic", "wecheckin-cc"))
		return encodeEmpty(encoder, startElement("serviceTask", attributes...))
	case NodeTypeNotify:
		attributes = append(attributes, attr("flowable:type", "external-worker"), attr("flowable:topic", "wecheckin-notify"))
		return encodeEmpty(encoder, startElement("serviceTask", attributes...))
	case NodeTypeAutomation:
		attributes = append(attributes, attr("flowable:type", "external-worker"), attr("flowable:topic", "wecheckin-automation"))
		return encodeEmpty(encoder, startElement("serviceTask", attributes...))
	case NodeTypeTimer:
		return encodeTimer(encoder, node, attributes)
	default:
		return fmt.Errorf("unsupported node type %q", node.Type)
	}
}

func notificationAttributes(config *NotificationConfig) []xml.Attr {
	if config == nil {
		return nil
	}
	return []xml.Attr{
		attr("flowable:notificationEnabled", strconv.FormatBool(config.Enabled)),
		attr("flowable:notificationChannels", strings.Join(config.Channels, ",")),
		attr("flowable:notificationTitle", config.Title),
		attr("flowable:notificationContent", config.Content),
	}
}

func resultNotificationAttributes(config *NotificationConfig) []xml.Attr {
	if config == nil {
		return nil
	}
	attributes := []xml.Attr{
		attr("flowable:resultNotificationEnabled", strconv.FormatBool(config.Enabled)),
		attr("flowable:resultNotificationChannels", strings.Join(config.Channels, ",")),
		attr("flowable:resultNotificationTitle", config.Title),
		attr("flowable:resultNotificationContent", config.Content),
	}
	if config.ResultTypes != nil {
		attributes = append(attributes, attr("flowable:resultNotificationResultTypes", strings.Join(config.ResultTypes, ",")))
	}
	return attributes
}

func encodeTimer(encoder *xml.Encoder, node Node, attributes []xml.Attr) error {
	catchEvent := startElement("intermediateCatchEvent", attributes...)
	if err := encoder.EncodeToken(catchEvent); err != nil {
		return err
	}
	timerDefinition := startElement("timerEventDefinition")
	if err := encoder.EncodeToken(timerDefinition); err != nil {
		return err
	}
	duration := startElement("timeDuration")
	if err := encoder.EncodeToken(duration); err != nil {
		return err
	}
	if err := encoder.EncodeToken(xml.CharData("PT" + strconv.FormatInt(node.Timer.DelaySeconds, 10) + "S")); err != nil {
		return err
	}
	if err := encoder.EncodeToken(duration.End()); err != nil {
		return err
	}
	if err := encoder.EncodeToken(timerDefinition.End()); err != nil {
		return err
	}
	return encoder.EncodeToken(catchEvent.End())
}

func encodeApproval(encoder *xml.Encoder, node Node, attributes []xml.Attr) error {
	if config := node.DepartmentApprovalChain; config != nil && config.Enabled {
		attributes = append(attributes,
			attr("flowable:departmentApprovalChain", "true"),
			attr("flowable:departmentApprovalChainStopMode", config.StopMode),
			attr("flowable:departmentApprovalChainStopDepartmentId", strconv.FormatUint(uint64(config.StopDepartmentID), 10)),
			attr("flowable:departmentApprovalChainMissingPolicy", config.MissingAssigneePolicy),
			attr("flowable:departmentApprovalChainSkipStarter", strconv.FormatBool(config.SkipStarter)),
		)
	}
	multiple := node.ApprovalMode != ApprovalModeSingle
	if multiple {
		attributes = append(attributes, attr("flowable:assignee", "${assignee}"))
	} else {
		key, value := singleAssigneeAttribute(node.ID, *node.Assignee)
		attributes = append(attributes, attr(key, value))
	}
	task := startElement("userTask", attributes...)
	if err := encoder.EncodeToken(task); err != nil {
		return err
	}
	if multiple {
		sequential := node.ApprovalMode == ApprovalModeSequential
		loop := startElement("multiInstanceLoopCharacteristics",
			attr("isSequential", strconv.FormatBool(sequential)),
			attr("flowable:collection", assigneeCollection(node.ID, *node.Assignee)),
			attr("flowable:elementVariable", "assignee"),
		)
		if err := encoder.EncodeToken(loop); err != nil {
			return err
		}
		if node.ApprovalMode == ApprovalModeCountersign {
			rate := strconv.FormatFloat(float64(node.CompletionRate)/100, 'f', -1, 64)
			completion := startElement("completionCondition")
			if err := encoder.EncodeToken(completion); err != nil {
				return err
			}
			if err := encoder.EncodeToken(xml.CharData("${nrOfCompletedInstances / nrOfInstances >= " + rate + "}")); err != nil {
				return err
			}
			if err := encoder.EncodeToken(completion.End()); err != nil {
				return err
			}
		}
		if err := encoder.EncodeToken(loop.End()); err != nil {
			return err
		}
	}
	return encoder.EncodeToken(task.End())
}

func encodeEdge(encoder *xml.Encoder, edge Edge) error {
	attributes := []xml.Attr{
		attr("id", edge.ID),
		attr("sourceRef", edge.Source),
		attr("targetRef", edge.Target),
	}
	if strings.TrimSpace(edge.Name) != "" {
		attributes = append(attributes, attr("name", edge.Name))
	}
	flow := startElement("sequenceFlow", attributes...)
	if err := encoder.EncodeToken(flow); err != nil {
		return err
	}
	if edge.Condition != nil && !edge.Default {
		condition := startElement("conditionExpression", attr("xsi:type", "tFormalExpression"))
		if err := encoder.EncodeToken(condition); err != nil {
			return err
		}
		if err := encoder.EncodeToken(xml.CharData(conditionExpression(*edge.Condition))); err != nil {
			return err
		}
		if err := encoder.EncodeToken(condition.End()); err != nil {
			return err
		}
	}
	return encoder.EncodeToken(flow.End())
}

func singleAssigneeAttribute(nodeID string, assignee Assignee) (string, string) {
	if assignee.Type == AssigneeTypeVariable {
		return "flowable:assignee", "${" + assignee.Value + "}"
	}
	return "flowable:assignee", "${workflowApprover_" + variableSegment(nodeID) + "}"
}

func assigneeCollection(nodeID string, assignee Assignee) string {
	if assignee.Type == AssigneeTypeVariable {
		return "${" + assignee.Value + "}"
	}
	return "${workflowApprovers_" + variableSegment(nodeID) + "}"
}

func variableSegment(value string) string {
	var builder strings.Builder
	for index, character := range value {
		valid := character == '_' || character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || index > 0 && character >= '0' && character <= '9'
		if valid {
			builder.WriteRune(character)
		} else {
			builder.WriteByte('_')
		}
	}
	if builder.Len() == 0 {
		return "node"
	}
	return builder.String()
}

func conditionExpression(condition Condition) string {
	operator := map[string]string{
		ConditionEQ:  "==",
		ConditionNE:  "!=",
		ConditionGT:  ">",
		ConditionGTE: ">=",
		ConditionLT:  "<",
		ConditionLTE: "<=",
	}[condition.Operator]
	return "${" + condition.Field + " " + operator + " " + formatConditionValue(condition.Value) + "}"
}

func formatConditionValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return "'" + strings.ReplaceAll(typed, "'", "\\'") + "'"
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case bool:
		return strconv.FormatBool(typed)
	default:
		encoded, _ := json.Marshal(typed)
		return string(encoded)
	}
}

func defaultFlowBySource(edges []Edge) map[string]string {
	result := make(map[string]string)
	for _, edge := range edges {
		if edge.Default {
			result[edge.Source] = edge.ID
		}
	}
	return result
}

func startElement(name string, attributes ...xml.Attr) xml.StartElement {
	return xml.StartElement{Name: xml.Name{Local: name}, Attr: attributes}
}

func attr(name, value string) xml.Attr {
	return xml.Attr{Name: xml.Name{Local: name}, Value: value}
}

func encodeEmpty(encoder *xml.Encoder, element xml.StartElement) error {
	if err := encoder.EncodeToken(element); err != nil {
		return err
	}
	return encoder.EncodeToken(element.End())
}
