package engine

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

type inclusionListAPIMock struct {
	handlerMockBackend
	processErr error
}

func (m *inclusionListAPIMock) ProcessInclusionList(_ *types.InclusionList) error {
	return m.processErr
}

func (m *inclusionListAPIMock) GetInclusionList() *types.InclusionList {
	return &types.InclusionList{Transactions: [][]byte{}}
}

func TestFlexibleUint64RejectsOverflowHex(t *testing.T) {
	var v flexibleUint64
	err := json.Unmarshal([]byte(`"0x10000000000000000"`), &v)
	if err == nil {
		t.Fatal("expected overflow error, got nil")
	}
	if !strings.Contains(err.Error(), "out of uint64 range") {
		t.Fatalf("expected overflow error, got %v", err)
	}
}

func TestFlexibleUint64ParsesQuotedDecimalAndHex(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want flexibleUint64
	}{
		{name: "decimal string", in: `"42"`, want: flexibleUint64(42)},
		{name: "hex string", in: `"0x2a"`, want: flexibleUint64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got flexibleUint64
			if err := json.Unmarshal([]byte(tt.in), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("want %d, got %d", tt.want, got)
			}
		})
	}
}

func TestHexBytesParsesQuotedHex(t *testing.T) {
	var got hexBytes
	if err := json.Unmarshal([]byte(`"0x0102ff"`), &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{0x01, 0x02, 0xff}
	if string(got) != string(want) {
		t.Fatalf("want %x, got %x", want, []byte(got))
	}
}

func TestNewInclusionListV1ReturnsInvalidOnBackendError(t *testing.T) {
	backendErr := errors.New("backend failed")
	api := NewEngineAPI(&inclusionListAPIMock{processErr: backendErr})

	result, err := api.NewInclusionListV1(InclusionListV1{
		Slot:           flexibleUint64(1),
		ValidatorIndex: flexibleUint64(2),
		Transactions:   []hexBytes{{0x01, 0x02}},
	})
	if err != nil {
		t.Fatalf("NewInclusionListV1 returned unexpected error: %v", err)
	}
	if result.Status != ILStatusInvalid {
		t.Fatalf("expected status %q, got %q", ILStatusInvalid, result.Status)
	}
	if result.Error == nil || !strings.Contains(*result.Error, backendErr.Error()) {
		t.Fatalf("expected backend error in response, got %+v", result.Error)
	}
}
