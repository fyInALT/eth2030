# rlp/rlperrors — Sentinel error values for RLP decoding

## Overview

This package exports the canonical sentinel error variables used throughout the
RLP decoder. Centralizing them in a sub-package lets both the decoder
implementation and its callers import error values without creating circular
dependencies.

## Functionality

**Exported errors**

- `ErrExpectedString` — a list was encountered where a string was expected
- `ErrExpectedList` — a string was encountered where a list was expected
- `ErrCanonSize` — non-canonical size encoding in a string prefix
- `ErrEOL` — end of the current list has been reached
- `ErrCanonInt` — integer uses non-canonical encoding (leading zeros)
- `ErrNonCanonicalSize` — size prefix is not in canonical form
- `ErrUint64Range` — decoded integer exceeds the uint64 range
- `ErrValueTooLarge` — value is too large to encode

**Parent package:** [rlp](../)
