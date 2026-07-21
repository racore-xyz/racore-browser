package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/racore/god/internal/identity"
	"github.com/racore/god/pkg/api"
)

const ProtocolVersion = "rmp/0.1"

func NewEnvelope(msgType string, data map[string]any, id *identity.NodeIdentity, nodeName string, roles []string) (*api.Envelope, error) {
	env := &api.Envelope{
		Protocol:  ProtocolVersion,
		Type:      msgType,
		NodeID:    id.NodeID(),
		Name:      nodeName,
		PublicKey: id.PublicKeyBase64(),
		Roles:     roles,
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}
	canonical, err := CanonicalJSON(env)
	if err != nil {
		return nil, fmt.Errorf("canonical json: %w", err)
	}
	sig, err := id.Sign(canonical)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	env.Signature = sig
	return env, nil
}

func VerifyEnvelope(raw []byte) (*api.Envelope, error) {
	var env api.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if env.Protocol != ProtocolVersion {
		return nil, fmt.Errorf("unsupported protocol: %s", env.Protocol)
	}

	sig := env.Signature
	env.Signature = ""

	canonical, err := CanonicalJSON(&env)
	if err != nil {
		return nil, fmt.Errorf("canonical json: %w", err)
	}
	pub, err := identity.Verify(env.PublicKey, canonical, sig)
	if err != nil {
		return nil, fmt.Errorf("verify: %w", err)
	}
	if identity.NodeIDFromPublicKey(pub) != env.NodeID {
		return nil, fmt.Errorf("nodeId mismatch")
	}
	env.Signature = sig
	return &env, nil
}

func CanonicalJSON(v any) ([]byte, error) {
	m, err := structToSortedMap(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	return b, nil
}

func structToSortedMap(v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Struct:
		rt := rv.Type()
		out := make(map[string]any, rt.NumField())
		for i := range rt.NumField() {
			field := rt.Field(i)
			if !field.IsExported() {
				continue
			}
			tag := field.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			name, _, _ := cut(tag, ",")
			if name == "" {
				name = field.Name
			}
			v := rv.Field(i).Interface()
			converted, err := structToSortedMap(v)
			if err != nil {
				return nil, err
			}
			out[name] = converted
		}
		return out, nil
	case reflect.Map:
		out := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			k := fmt.Sprintf("%v", iter.Key().Interface())
			converted, err := structToSortedMap(iter.Value().Interface())
			if err != nil {
				return nil, err
			}
			out[k] = converted
		}
		keys := make([]string, 0, len(out))
		for k := range out {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sorted := make(map[string]any, len(out))
		for _, k := range keys {
			sorted[k] = out[k]
		}
		return sorted, nil
	case reflect.Slice, reflect.Array:
		n := rv.Len()
		out := make([]any, n)
		for i := range n {
			converted, err := structToSortedMap(rv.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			out[i] = converted
		}
		return out, nil
	default:
		return v, nil
	}
}

func cut(s, sep string) (string, string, bool) {
	i := 0
	for i < len(s) {
		if s[i] == sep[0] {
			return s[:i], s[i+1:], true
		}
		i++
	}
	return s, "", false
}

func EnvelopeToMessage(env *api.Envelope) []byte {
	data, _ := json.Marshal(env)
	return data
}

func ParseMessage(raw []byte) (*api.Envelope, error) {
	return VerifyEnvelope(raw)
}
