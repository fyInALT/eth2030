package core

import (
	"bytes"
	"sync"
	"testing"

	"github.com/eth2030/eth2030/core/execution"
	"github.com/eth2030/eth2030/core/types"
)

// helper to make a schema ID from a short string.
func schemaID(s string) types.Hash {
	var h types.Hash
	copy(h[:], s)
	return h
}

func dataKey(s string) types.Hash {
	var h types.Hash
	copy(h[:], s)
	return h
}

func TestDataTypeString(t *testing.T) {
	tests := []struct {
		dt   execution.DataType
		want string
	}{
		{execution.TypeUint256, "uint256"},
		{execution.TypeAddress, "address"},
		{execution.TypeBytes32, "bytes32"},
		{execution.TypeString, "string"},
		{execution.TypeBool, "bool"},
		{execution.TypeArray, "array"},
		{execution.DataType(255), "unknown"},
	}
	for _, tc := range tests {
		if got := tc.dt.String(); got != tc.want {
			t.Errorf("execution.DataType(%d).String() = %q, want %q", tc.dt, got, tc.want)
		}
	}
}

func TestRegisterSchema(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("test-schema")

	fields := []execution.SchemaField{
		{Name: "name", FieldType: execution.TypeString, Required: true, MaxSize: 100},
		{Name: "balance", FieldType: execution.TypeUint256, Required: false, MaxSize: 32},
	}

	if err := store.RegisterSchema(sid, fields); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Duplicate registration should fail.
	if err := store.RegisterSchema(sid, fields); err != execution.ErrSchemaExists {
		t.Fatalf("expected execution.ErrSchemaExists, got %v", err)
	}
}

func TestRegisterSchemaEmpty(t *testing.T) {
	store := execution.NewRichDataStore()
	err := store.RegisterSchema(schemaID("empty"), nil)
	if err != execution.ErrEmptySchema {
		t.Fatalf("expected execution.ErrEmptySchema, got %v", err)
	}
	err = store.RegisterSchema(schemaID("empty2"), []execution.SchemaField{})
	if err != execution.ErrEmptySchema {
		t.Fatalf("expected execution.ErrEmptySchema for empty slice, got %v", err)
	}
}

func TestRegisterSchemaDuplicateField(t *testing.T) {
	store := execution.NewRichDataStore()
	fields := []execution.SchemaField{
		{Name: "x", FieldType: execution.TypeBool},
		{Name: "x", FieldType: execution.TypeString},
	}
	err := store.RegisterSchema(schemaID("dup"), fields)
	if err != execution.ErrDuplicateFieldName {
		t.Fatalf("expected execution.ErrDuplicateFieldName, got %v", err)
	}
}

func TestGetSchema(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("s1")
	fields := []execution.SchemaField{
		{Name: "owner", FieldType: execution.TypeAddress, Required: true, MaxSize: 20},
	}
	store.RegisterSchema(sid, fields)

	got, err := store.GetSchema(sid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "owner" {
		t.Fatalf("unexpected schema: %+v", got)
	}

	// Mutation of returned slice should not affect the store.
	got[0].Name = "mutated"
	got2, _ := store.GetSchema(sid)
	if got2[0].Name != "owner" {
		t.Fatal("GetSchema must return a defensive copy")
	}
}

func TestGetSchemaNotFound(t *testing.T) {
	store := execution.NewRichDataStore()
	_, err := store.GetSchema(schemaID("nope"))
	if err != execution.ErrSchemaNotFound {
		t.Fatalf("expected execution.ErrSchemaNotFound, got %v", err)
	}
}

func TestListSchemas(t *testing.T) {
	store := execution.NewRichDataStore()

	// Empty store.
	if list := store.ListSchemas(); list != nil {
		t.Fatalf("expected nil for empty store, got %v", list)
	}

	s1 := schemaID("s1")
	s2 := schemaID("s2")
	store.RegisterSchema(s1, []execution.SchemaField{{Name: "a", FieldType: execution.TypeBool}})
	store.RegisterSchema(s2, []execution.SchemaField{{Name: "b", FieldType: execution.TypeBool}})

	list := store.ListSchemas()
	if len(list) != 2 || list[0] != s1 || list[1] != s2 {
		t.Fatalf("unexpected schema list: %v", list)
	}
}

func TestValidateData(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("v")
	fields := []execution.SchemaField{
		{Name: "name", FieldType: execution.TypeString, Required: true, MaxSize: 10},
		{Name: "opt", FieldType: execution.TypeBool, Required: false},
	}
	store.RegisterSchema(sid, fields)

	// Valid data.
	data := map[string][]byte{"name": []byte("hello")}
	if err := store.ValidateData(sid, data); err != nil {
		t.Fatalf("valid data rejected: %v", err)
	}

	// Missing required field.
	if err := store.ValidateData(sid, map[string][]byte{"opt": {1}}); err != execution.ErrMissingRequired {
		t.Fatalf("expected execution.ErrMissingRequired, got %v", err)
	}

	// Unknown field.
	bad := map[string][]byte{"name": []byte("x"), "unknown": {1}}
	if err := store.ValidateData(sid, bad); err != execution.ErrFieldNotInSchema {
		t.Fatalf("expected execution.ErrFieldNotInSchema, got %v", err)
	}

	// Field too large.
	big := map[string][]byte{"name": make([]byte, 11)}
	if err := store.ValidateData(sid, big); err != execution.ErrFieldTooLarge {
		t.Fatalf("expected execution.ErrFieldTooLarge, got %v", err)
	}

	// Non-existent schema.
	if err := store.ValidateData(schemaID("nope"), data); err != execution.ErrSchemaNotFound {
		t.Fatalf("expected execution.ErrSchemaNotFound, got %v", err)
	}
}

func TestStoreAndGetData(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("sg")
	fields := []execution.SchemaField{
		{Name: "owner", FieldType: execution.TypeAddress, Required: true, MaxSize: 20},
		{Name: "label", FieldType: execution.TypeString, Required: false, MaxSize: 64},
	}
	store.RegisterSchema(sid, fields)

	key := dataKey("k1")
	data := map[string][]byte{
		"owner": {0x01, 0x02, 0x03},
		"label": []byte("test"),
	}

	if err := store.StoreData(sid, key, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := store.GetData(sid, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(got["owner"], data["owner"]) {
		t.Fatal("owner field mismatch")
	}
	if !bytes.Equal(got["label"], data["label"]) {
		t.Fatal("label field mismatch")
	}

	// Returned data is a deep copy.
	got["owner"][0] = 0xFF
	got2, _ := store.GetData(sid, key)
	if got2["owner"][0] != 0x01 {
		t.Fatal("GetData must return a deep copy")
	}
}

func TestStoreDataDuplicate(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("dup")
	store.RegisterSchema(sid, []execution.SchemaField{{Name: "x", FieldType: execution.TypeBool, Required: true}})

	key := dataKey("k")
	data := map[string][]byte{"x": {1}}
	store.StoreData(sid, key, data)

	err := store.StoreData(sid, key, data)
	if err != execution.ErrDataExists {
		t.Fatalf("expected execution.ErrDataExists, got %v", err)
	}
}

func TestStoreDataValidation(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("val")
	store.RegisterSchema(sid, []execution.SchemaField{
		{Name: "req", FieldType: execution.TypeBool, Required: true},
	})

	// Missing required.
	err := store.StoreData(sid, dataKey("k"), map[string][]byte{})
	if err != execution.ErrMissingRequired {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestGetDataNotFound(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("gd")
	store.RegisterSchema(sid, []execution.SchemaField{{Name: "a", FieldType: execution.TypeBool}})

	_, err := store.GetData(sid, dataKey("nope"))
	if err != execution.ErrDataNotFound {
		t.Fatalf("expected execution.ErrDataNotFound, got %v", err)
	}

	// Unknown schema.
	_, err = store.GetData(schemaID("unknown"), dataKey("k"))
	if err != execution.ErrSchemaNotFound {
		t.Fatalf("expected execution.ErrSchemaNotFound, got %v", err)
	}
}

func TestQueryByField(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("q")
	store.RegisterSchema(sid, []execution.SchemaField{
		{Name: "status", FieldType: execution.TypeString, Required: true, MaxSize: 32},
		{Name: "owner", FieldType: execution.TypeAddress, Required: false, MaxSize: 20},
	})

	k1 := dataKey("k1")
	k2 := dataKey("k2")
	k3 := dataKey("k3")

	store.StoreData(sid, k1, map[string][]byte{"status": []byte("active")})
	store.StoreData(sid, k2, map[string][]byte{"status": []byte("active")})
	store.StoreData(sid, k3, map[string][]byte{"status": []byte("closed")})

	// Query "active" status.
	results, err := store.QueryByField(sid, "status", []byte("active"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Ensure k1 and k2 are in results.
	found := make(map[types.Hash]bool)
	for _, r := range results {
		found[r] = true
	}
	if !found[k1] || !found[k2] {
		t.Fatal("expected k1 and k2 in results")
	}

	// Query "closed" status.
	results, err = store.QueryByField(sid, "status", []byte("closed"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0] != k3 {
		t.Fatalf("expected [k3], got %v", results)
	}

	// Query non-matching value.
	results, err = store.QueryByField(sid, "status", []byte("unknown"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Fatalf("expected nil for no matches, got %v", results)
	}
}

func TestQueryByFieldErrors(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("qe")
	store.RegisterSchema(sid, []execution.SchemaField{{Name: "a", FieldType: execution.TypeBool}})

	// Unknown schema.
	_, err := store.QueryByField(schemaID("nope"), "a", nil)
	if err != execution.ErrSchemaNotFound {
		t.Fatalf("expected execution.ErrSchemaNotFound, got %v", err)
	}

	// Unknown field.
	_, err = store.QueryByField(sid, "bad_field", nil)
	if err != execution.ErrFieldNotInSchema {
		t.Fatalf("expected execution.ErrFieldNotInSchema, got %v", err)
	}
}

func TestDeleteData(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("del")
	store.RegisterSchema(sid, []execution.SchemaField{
		{Name: "tag", FieldType: execution.TypeString, Required: true, MaxSize: 32},
	})

	key := dataKey("k1")
	store.StoreData(sid, key, map[string][]byte{"tag": []byte("x")})

	// Verify data exists.
	if _, err := store.GetData(sid, key); err != nil {
		t.Fatal("data should exist before delete")
	}

	// Delete.
	if err := store.DeleteData(sid, key); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify gone.
	_, err := store.GetData(sid, key)
	if err != execution.ErrDataNotFound {
		t.Fatalf("expected execution.ErrDataNotFound after delete, got %v", err)
	}

	// Index should also be cleaned up.
	results, _ := store.QueryByField(sid, "tag", []byte("x"))
	if results != nil {
		t.Fatalf("expected no index entries after delete, got %v", results)
	}
}

func TestDeleteDataErrors(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("de")
	store.RegisterSchema(sid, []execution.SchemaField{{Name: "a", FieldType: execution.TypeBool}})

	// Unknown key.
	err := store.DeleteData(sid, dataKey("nope"))
	if err != execution.ErrDataNotFound {
		t.Fatalf("expected execution.ErrDataNotFound, got %v", err)
	}

	// Unknown schema.
	err = store.DeleteData(schemaID("nope"), dataKey("k"))
	if err != execution.ErrSchemaNotFound {
		t.Fatalf("expected execution.ErrSchemaNotFound, got %v", err)
	}
}

func TestStoreDataDeepCopy(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("cp")
	store.RegisterSchema(sid, []execution.SchemaField{
		{Name: "val", FieldType: execution.TypeBytes32, Required: true, MaxSize: 32},
	})

	original := []byte{1, 2, 3}
	data := map[string][]byte{"val": original}
	store.StoreData(sid, dataKey("k"), data)

	// Mutate the original slice after storing.
	original[0] = 0xFF

	got, _ := store.GetData(sid, dataKey("k"))
	if got["val"][0] != 1 {
		t.Fatal("StoreData must deep-copy input data")
	}
}

func TestConcurrentRichData(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("conc")
	store.RegisterSchema(sid, []execution.SchemaField{
		{Name: "idx", FieldType: execution.TypeString, Required: true, MaxSize: 32},
	})

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			key := dataKey("k" + string(rune(n+65)))
			val := []byte{byte(n)}
			_ = store.StoreData(sid, key, map[string][]byte{"idx": val})
			_, _ = store.GetData(sid, key)
			_, _ = store.QueryByField(sid, "idx", val)
			store.ListSchemas()
		}(i)
	}

	wg.Wait()
}

func TestMaxSizeZeroMeansUnlimited(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("unlim")
	store.RegisterSchema(sid, []execution.SchemaField{
		{Name: "data", FieldType: execution.TypeString, Required: true, MaxSize: 0},
	})

	// A large value should be accepted when MaxSize is 0 (unlimited).
	bigVal := make([]byte, 10000)
	err := store.StoreData(sid, dataKey("k"), map[string][]byte{"data": bigVal})
	if err != nil {
		t.Fatalf("MaxSize 0 should mean unlimited, got error: %v", err)
	}
}

func TestQueryAfterDeleteAndRestore(t *testing.T) {
	store := execution.NewRichDataStore()
	sid := schemaID("qdr")
	store.RegisterSchema(sid, []execution.SchemaField{
		{Name: "color", FieldType: execution.TypeString, Required: true, MaxSize: 10},
	})

	key := dataKey("k1")
	store.StoreData(sid, key, map[string][]byte{"color": []byte("red")})

	// Delete.
	store.DeleteData(sid, key)

	// After delete, query should return nothing.
	results, _ := store.QueryByField(sid, "color", []byte("red"))
	if results != nil {
		t.Fatal("expected no results after delete")
	}

	// Re-store with same key, different value.
	store.StoreData(sid, key, map[string][]byte{"color": []byte("blue")})

	// Old value no longer indexed.
	results, _ = store.QueryByField(sid, "color", []byte("red"))
	if results != nil {
		t.Fatal("old value should not be indexed")
	}

	// New value indexed.
	results, _ = store.QueryByField(sid, "color", []byte("blue"))
	if len(results) != 1 || results[0] != key {
		t.Fatal("new value should be indexed")
	}
}
