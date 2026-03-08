# richdata

Block-level rich metadata registry for structured validator annotations.

## Overview

Package `richdata` provides a schema-driven registry where validators can
attach structured metadata (`RichDataEntry`) to attestations and blocks.
Callers first register a named `RichDataSchema` that declares required and
optional fields with their types. Entries submitted against that schema are
validated for field presence, type correctness, and total size before storage.

`RichDataRegistry` stores entries keyed by `(schemaName, slot)` and supports
pruning of old slots to bound memory usage. The package is used by the parent
consensus package to record inclusion delays, attestation scores, and other
per-slot analytics.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `RichDataRegistry` | Thread-safe store: schemas and entries |
| `RichDataSchema` | Named schema: `Fields []FieldDefinition`, `MaxSize int`, `Version` |
| `FieldDefinition` | `Name`, `Type FieldType`, `Required bool` |
| `RichDataEntry` | `SchemaName`, `ValidatorID`, `Slot`, `Data map[string]interface{}`, `Timestamp` |
| `FieldType` | `FieldString`, `FieldInt`, `FieldBool`, `FieldBytes` |

### Functions / methods

| Name | Description |
|------|-------------|
| `NewRichDataRegistry() *RichDataRegistry` | Create empty registry |
| `(*RichDataRegistry).RegisterSchema(schema) error` | Register; error if name already exists |
| `(*RichDataRegistry).GetSchema(name) *RichDataSchema` | Look up a schema |
| `(*RichDataRegistry).SubmitEntry(entry) error` | Validate and store an entry |
| `(*RichDataRegistry).GetEntries(schemaName, slot) []RichDataEntry` | Retrieve entries |
| `(*RichDataRegistry).ValidateEntry(entry) error` | Validate without storing |
| `(*RichDataRegistry).PruneOldEntries(beforeSlot) int` | Remove entries older than a slot |
| `(*RichDataRegistry).SchemaCount() int` | Number of registered schemas |
| `(*RichDataRegistry).EntryCount() int` | Total stored entries |

### Errors

`ErrRichDataSchemaExists`, `ErrRichDataSchemaNotFound`, `ErrRichDataEntryInvalid`, `ErrRichDataTooLarge`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/richdata"

reg := richdata.NewRichDataRegistry()
reg.RegisterSchema(richdata.RichDataSchema{
    Name:    "inclusion_delay",
    Fields:  []richdata.FieldDefinition{{Name: "delay", Type: richdata.FieldInt, Required: true}},
    MaxSize: 64,
})
reg.SubmitEntry(richdata.RichDataEntry{
    SchemaName: "inclusion_delay", ValidatorID: 42, Slot: 100,
    Data: map[string]interface{}{"delay": uint64(1)},
})
```

[← consensus](../README.md)
