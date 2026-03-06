package consensus

// rich_data_compat.go re-exports types from consensus/richdata for backward compatibility.

import "github.com/eth2030/eth2030/consensus/richdata"

// RichData type aliases.
type (
	FieldType        = richdata.FieldType
	FieldDefinition  = richdata.FieldDefinition
	RichDataSchema   = richdata.RichDataSchema
	RichDataEntry    = richdata.RichDataEntry
	RichDataRegistry = richdata.RichDataRegistry
)

// RichData field type constants.
const (
	FieldString = richdata.FieldString
	FieldInt    = richdata.FieldInt
	FieldBool   = richdata.FieldBool
	FieldBytes  = richdata.FieldBytes
)

// RichData error variables.
var (
	ErrRichDataSchemaExists   = richdata.ErrRichDataSchemaExists
	ErrRichDataSchemaNotFound = richdata.ErrRichDataSchemaNotFound
	ErrRichDataEntryInvalid   = richdata.ErrRichDataEntryInvalid
	ErrRichDataTooLarge       = richdata.ErrRichDataTooLarge
)

// RichData function wrappers.
func NewRichDataRegistry() *RichDataRegistry { return richdata.NewRichDataRegistry() }
