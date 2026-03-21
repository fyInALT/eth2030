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
