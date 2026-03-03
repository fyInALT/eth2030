package rpc

import "encoding/json"

// isNullResult returns true if v is a JSON null response result,
// either a Go nil or json.RawMessage("null").
func isNullResult(v interface{}) bool {
	if v == nil {
		return true
	}
	raw, ok := v.(json.RawMessage)
	return ok && string(raw) == "null"
}
