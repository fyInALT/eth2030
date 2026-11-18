package payload

import "encoding/json"

// BlobAndProofV1 uses canonical Engine DATA fields. Commitment is backend-only.
func (b BlobAndProofV1) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blob  hexBytes `json:"blob"`
		Proof hexBytes `json:"proof"`
	}{hexBytes(b.Blob), hexBytes(b.Proof)})
}
func (b *BlobAndProofV1) UnmarshalJSON(data []byte) error {
	var wire struct {
		Blob  hexBytes `json:"blob"`
		Proof hexBytes `json:"proof"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	b.Blob, b.Proof, b.Commitment = []byte(wire.Blob), []byte(wire.Proof), nil
	return nil
}
