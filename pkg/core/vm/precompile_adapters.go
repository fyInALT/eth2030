// Package vm implements the Ethereum Virtual Machine.
//
// precompile_adapters.go provides exported adapter types for eth2030's custom
// precompiles, enabling the geth adapter package to reference them. Each
// adapter wraps the corresponding unexported precompile struct.
package vm

// BN256AddGlamsterdanAdapter exposes the Glamsterdam-repriced bn256Add.
type BN256AddGlamsterdanAdapter struct{ bn256AddGlamsterdan }

// BN256PairingGlamsterdanAdapter exposes the Glamsterdam-repriced bn256Pairing.
type BN256PairingGlamsterdanAdapter struct{ bn256PairingGlamsterdan }

// Blake2FGlamsterdanAdapter exposes the Glamsterdam-repriced blake2F.
type Blake2FGlamsterdanAdapter struct{ blake2FGlamsterdan }

// KZGPointEvalGlamsterdanAdapter exposes the Glamsterdam-repriced kzgPointEval.
type KZGPointEvalGlamsterdanAdapter struct{ kzgPointEvaluationGlamsterdan }

// NTTPrecompileAdapter exposes the legacy NTT precompile (kept for adapter compatibility).
type NTTPrecompileAdapter struct{ nttPrecompile }

// NTTFWAdapter exposes the forward NTT precompile (0x0f).
type NTTFWAdapter struct{ nttFWPrecompile }

// NTTINVAdapter exposes the inverse NTT precompile (0x10).
type NTTINVAdapter struct{ nttINVPrecompile }

// NTTVecMulModAdapter exposes the vector mul mod precompile (0x11).
type NTTVecMulModAdapter struct{ nttVecMulModPrecompile }

// NTTVecAddModAdapter exposes the vector add mod precompile (0x12).
type NTTVecAddModAdapter struct{ nttVecAddModPrecompile }

// NTTDotProductAdapter exposes the dot product precompile (0x13).
type NTTDotProductAdapter struct{ nttDotProductPrecompile }

// NTTButterflyAdapter exposes the butterfly precompile (0x14).
type NTTButterflyAdapter struct{ nttButterflyPrecompile }

// NiiModExpAdapter exposes the NII modexp precompile (0x0201).
type NiiModExpAdapter struct{ NiiModExpPrecompile }

// NiiFieldMulAdapter exposes the NII field multiplication precompile (0x0202).
type NiiFieldMulAdapter struct{ NiiFieldMulPrecompile }

// NiiFieldInvAdapter exposes the NII field inverse precompile (0x0203).
type NiiFieldInvAdapter struct{ NiiFieldInvPrecompile }

// NiiBatchVerifyAdapter exposes the NII batch verify precompile (0x0204).
type NiiBatchVerifyAdapter struct{ NiiBatchVerifyPrecompile }

// FieldMulExtAdapter exposes the extended field multiplication precompile (0x0205).
type FieldMulExtAdapter struct{ FieldMulExtPrecompile }

// FieldInvExtAdapter exposes the extended field inverse precompile (0x0206).
type FieldInvExtAdapter struct{ FieldInvExtPrecompile }

// FieldExpAdapter exposes the field exponentiation precompile (0x0207).
type FieldExpAdapter struct{ FieldExpPrecompile }

// BatchFieldVerifyAdapter exposes the batch field verify precompile (0x0208).
type BatchFieldVerifyAdapter struct{ BatchFieldVerifyPrecompile }
