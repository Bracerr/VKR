package ragtext

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// FormatDocumentText собирает читаемое содержимое карточки из заголовка и payload.
func FormatDocumentText(title string, payload json.RawMessage) string {
	var b strings.Builder
	if t := strings.TrimSpace(title); t != "" {
		b.WriteString(t)
		b.WriteString("\n\n")
	}
	var m map[string]any
	if len(payload) > 0 && json.Unmarshal(payload, &m) == nil && len(m) > 0 {
		writeMap(&b, m, 0)
	}
	return strings.TrimSpace(b.String())
}

func writeMap(b *strings.Builder, m map[string]any, depth int) {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k == "lines" || k == "components" || k == "operations" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	indent := strings.Repeat("  ", depth)
	for _, k := range keys {
		label := fieldLabel(k)
		writeValue(b, indent, label, m[k])
	}
	if lines, ok := m["lines"].([]any); ok && len(lines) > 0 {
		b.WriteString(indent)
		b.WriteString("Позиции:\n")
		for i, item := range lines {
			if row, ok := item.(map[string]any); ok {
				b.WriteString(indent)
				b.WriteString("  ")
				b.WriteString(fmt.Sprintf("%d) %s\n", i+1, formatLine(row)))
			}
		}
	}
	if comps, ok := m["components"].([]any); ok && len(comps) > 0 {
		b.WriteString(indent)
		b.WriteString("Компоненты:\n")
		for i, item := range comps {
			if row, ok := item.(map[string]any); ok {
				b.WriteString(indent)
				b.WriteString("  ")
				b.WriteString(fmt.Sprintf("%d) %s\n", i+1, formatLine(row)))
			}
		}
	}
	if ops, ok := m["operations"].([]any); ok && len(ops) > 0 {
		b.WriteString(indent)
		b.WriteString("Операции:\n")
		for i, item := range ops {
			if row, ok := item.(map[string]any); ok {
				b.WriteString(indent)
				b.WriteString("  ")
				b.WriteString(fmt.Sprintf("%d) %s\n", i+1, formatOperation(row)))
			}
		}
	}
}

func writeValue(b *strings.Builder, indent, label string, v any) {
	switch x := v.(type) {
	case nil:
		return
	case string:
		if strings.TrimSpace(x) == "" {
			return
		}
		b.WriteString(indent)
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(x)
		b.WriteString("\n")
	case float64:
		b.WriteString(indent)
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(fmt.Sprintf("%v", x))
		b.WriteString("\n")
	case bool:
		b.WriteString(indent)
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(fmt.Sprintf("%v", x))
		b.WriteString("\n")
	case map[string]any:
		if name, ok := x["name"].(string); ok {
			b.WriteString(indent)
			b.WriteString(label)
			b.WriteString(": ")
			b.WriteString(name)
			if inn, ok := x["inn"].(string); ok {
				b.WriteString(" (ИНН ")
				b.WriteString(inn)
				b.WriteString(")")
			}
			b.WriteString("\n")
			return
		}
		b.WriteString(indent)
		b.WriteString(label)
		b.WriteString(":\n")
		writeMap(b, x, depth(indent)+1)
	default:
		b.WriteString(indent)
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(fmt.Sprintf("%v", v))
		b.WriteString("\n")
	}
}

func depth(indent string) int {
	if indent == "" {
		return 0
	}
	return len(indent)/2 + 1
}

func formatLine(row map[string]any) string {
	parts := []string{}
	if sku, ok := row["sku"].(string); ok {
		parts = append(parts, sku)
	}
	if name, ok := row["name"].(string); ok {
		parts = append(parts, name)
	}
	if qty, ok := row["qty"]; ok {
		unit := ""
		if u, ok := row["unit"].(string); ok {
			unit = " " + u
		}
		parts = append(parts, fmt.Sprintf("— %v%s", qty, unit))
	}
	if up, ok := row["unit_price"].(float64); ok {
		parts = append(parts, fmt.Sprintf("по %v", up))
	}
	if am, ok := row["amount"].(float64); ok {
		parts = append(parts, fmt.Sprintf("сумма %v", am))
	}
	return strings.Join(parts, " ")
}

func formatOperation(row map[string]any) string {
	parts := []string{}
	if no, ok := row["op_no"].(float64); ok {
		parts = append(parts, fmt.Sprintf("№%v", no))
	}
	if wc, ok := row["workcenter"].(string); ok {
		parts = append(parts, wc)
	}
	if name, ok := row["name"].(string); ok {
		parts = append(parts, name)
	}
	if min, ok := row["std_minutes"].(float64); ok {
		parts = append(parts, fmt.Sprintf("%v мин", min))
	}
	return strings.Join(parts, ", ")
}

func fieldLabel(key string) string {
	labels := map[string]string{
		"document_kind":      "Вид документа",
		"request_no":         "Номер заявки",
		"order_no":           "Номер заказа",
		"contract_no":        "Номер договора",
		"shipment_no":        "Номер отгрузки",
		"bom_no":             "Спецификация",
		"routing_no":         "Маршрут",
		"department":         "Подразделение",
		"requested_by":       "Инициатор",
		"needed_by":          "Нужно к дате",
		"priority":           "Приоритет",
		"justification":      "Обоснование",
		"cost_center":        "Центр затрат",
		"currency":           "Валюта",
		"total_amount":       "Сумма",
		"preferred_supplier": "Поставщик",
		"supplier":           "Поставщик",
		"customer":           "Клиент",
		"payment_terms":      "Условия оплаты",
		"delivery_date":      "Дата поставки",
		"subject":            "Предмет",
		"valid_from":         "Действует с",
		"valid_to":           "Действует по",
		"max_amount":         "Лимит суммы",
		"ship_to":            "Адрес доставки",
		"carrier":            "Перевозчик",
		"product_code":       "Изделие",
		"product_name":       "Наименование изделия",
		"revision":           "Ревизия",
	}
	if l, ok := labels[key]; ok {
		return l
	}
	return key
}
