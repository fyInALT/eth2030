package debugapi

import (
	"encoding/json"
	"testing"
)

func TestBlockTraceResult_JSON(t *testing.T) {
	result := &BlockTraceResult{
		TxHash: "0x1234",
		Result: &TraceResult{
			Gas:         21000,
			Failed:      false,
			ReturnValue: "",
			StructLogs:  []StructLog{},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded BlockTraceResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.TxHash != "0x1234" {
		t.Fatalf("want txHash 0x1234, got %v", decoded.TxHash)
	}
	if decoded.Result.Gas != 21000 {
		t.Fatalf("want gas 21000, got %d", decoded.Result.Gas)
	}
}
