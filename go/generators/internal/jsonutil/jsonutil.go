// Package jsonutil provides deterministic JSON-object marshaling shared by
// the lexgen and parsegen table encoders. Output bytes are part of the tables
// interchange format: sorted keys, int keys quoted, nil maps as null.
package jsonutil

import (
	"bytes"
	"encoding/json"
	"slices"
	"strconv"
)

// Field is a pre-marshaled JSON object member.
type Field struct {
	Name  string
	Value []byte
}

// MarshalOrderedFields renders fields as a JSON object, in the given order.
func MarshalOrderedFields(fields []Field) []byte {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, field := range fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(strconv.Quote(field.Name))
		buf.WriteByte(':')
		buf.Write(field.Value)
	}
	buf.WriteByte('}')
	return buf.Bytes()
}

// MarshalSortedMap renders m as a JSON object with keys in sorted order and
// values marshaled by marshalValue. A nil map renders as null.
func MarshalSortedMap[K int | string, V any](m map[K]V, marshalValue func(V) ([]byte, error)) ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(quoteKey(key))
		buf.WriteByte(':')
		valueBytes, err := marshalValue(m[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// MarshalSortedMapJSON is MarshalSortedMap with encoding/json for the values.
func MarshalSortedMapJSON[K int | string, V any](m map[K]V) ([]byte, error) {
	return MarshalSortedMap(m, func(v V) ([]byte, error) { return json.Marshal(v) })
}

func quoteKey[K int | string](key K) string {
	switch k := any(key).(type) {
	case int:
		return strconv.Quote(strconv.Itoa(k))
	case string:
		return strconv.Quote(k)
	}
	panic("unreachable")
}
