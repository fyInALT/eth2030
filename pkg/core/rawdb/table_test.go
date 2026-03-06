package rawdb

import (
	"bytes"
	"sort"
	"testing"
)

// --- Table basic CRUD ---

func TestTable_HasPutGet(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "pfx/")

	key := []byte("block1")
	val := []byte("header-data")

	ok, err := tbl.Has(key)
	if err != nil || ok {
		t.Fatal("expected key absent before Put")
	}
	if err := tbl.Put(key, val); err != nil {
		t.Fatal(err)
	}
	ok, err = tbl.Has(key)
	if err != nil || !ok {
		t.Fatal("expected key present after Put")
	}
	got, err := tbl.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, val) {
		t.Fatalf("Get: got %x, want %x", got, val)
	}
}

func TestTable_Delete(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "t/")
	key := []byte("k")

	tbl.Put(key, []byte("v"))
	if err := tbl.Delete(key); err != nil {
		t.Fatal(err)
	}
	ok, _ := tbl.Has(key)
	if ok {
		t.Fatal("key should be absent after Delete")
	}
}

func TestTable_Close(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "t/")
	if err := tbl.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func TestTable_Prefix(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "chaindata/")
	if tbl.Prefix() != "chaindata/" {
		t.Fatalf("wrong prefix: %q", tbl.Prefix())
	}
}

func TestTable_IsolatesKeys(t *testing.T) {
	db := NewMemoryDB()
	t1 := NewTable(db, "a/")
	t2 := NewTable(db, "b/")
	key := []byte("x")

	t1.Put(key, []byte("from-t1"))
	t2.Put(key, []byte("from-t2"))

	v1, _ := t1.Get(key)
	v2, _ := t2.Get(key)
	if bytes.Equal(v1, v2) {
		t.Fatal("tables should be isolated by prefix")
	}
}

// --- tableBatch ---

func TestTableBatch_PutWriteRead(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "x/")
	b := tbl.NewBatch()

	b.Put([]byte("k1"), []byte("v1"))
	b.Put([]byte("k2"), []byte("v2"))

	if b.ValueSize() == 0 {
		t.Fatal("expected non-zero ValueSize after Put")
	}

	if err := b.Write(); err != nil {
		t.Fatal(err)
	}
	v, err := tbl.Get([]byte("k1"))
	if err != nil || !bytes.Equal(v, []byte("v1")) {
		t.Fatalf("expected v1, got %s %v", v, err)
	}
}

func TestTableBatch_Delete(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "x/")
	tbl.Put([]byte("k"), []byte("v"))

	b := tbl.NewBatch()
	b.Delete([]byte("k"))
	b.Write()

	ok, _ := tbl.Has([]byte("k"))
	if ok {
		t.Fatal("key should be deleted after batch delete")
	}
}

func TestTableBatch_Reset(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "x/")
	b := tbl.NewBatch()
	b.Put([]byte("k"), []byte("v"))

	b.Reset()
	if b.ValueSize() != 0 {
		t.Fatal("expected zero ValueSize after Reset")
	}
	// Write should be a no-op; key should not exist.
	b.Write()
	ok, _ := tbl.Has([]byte("k"))
	if ok {
		t.Fatal("key should not exist after Reset+Write")
	}
}

// --- Iterator ---

func TestTable_NewIterator(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "ns/")

	pairs := map[string]string{
		"alpha": "1",
		"beta":  "2",
		"gamma": "3",
	}
	for k, v := range pairs {
		tbl.Put([]byte(k), []byte(v))
	}

	it := tbl.NewIterator(nil)
	defer it.Release()

	found := map[string]string{}
	for it.Next() {
		found[string(it.Key())] = string(it.Value())
	}
	for k, v := range pairs {
		if found[k] != v {
			t.Fatalf("key %q: want %q, got %q", k, v, found[k])
		}
	}
}

func TestTable_NewIterator_WithPrefix(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "ns/")
	tbl.Put([]byte("foo/1"), []byte("a"))
	tbl.Put([]byte("foo/2"), []byte("b"))
	tbl.Put([]byte("bar/1"), []byte("c"))

	it := tbl.NewIterator([]byte("foo/"))
	defer it.Release()

	var keys []string
	for it.Next() {
		keys = append(keys, string(it.Key()))
	}
	sort.Strings(keys)
	if len(keys) != 2 || keys[0] != "foo/1" || keys[1] != "foo/2" {
		t.Fatalf("expected foo/ keys, got %v", keys)
	}
}

func TestTable_NewIterator_Empty(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "ns/")
	it := tbl.NewIterator(nil)
	defer it.Release()
	if it.Next() {
		t.Fatal("expected empty iterator")
	}
}

// --- Compaction hints ---

func TestCompactionHintForTable(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "abc/")
	hint := CompactionHintForTable(tbl)
	if !bytes.HasPrefix(hint.Start, []byte("abc/")) {
		t.Fatalf("Start should have prefix, got %x", hint.Start)
	}
	if bytes.Compare(hint.Start, hint.Limit) >= 0 {
		t.Fatal("Limit should be greater than Start")
	}
}

func TestCompactionHintForPrefix(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTable(db, "ns/")
	hint := CompactionHintForPrefix(tbl, []byte("header"))
	if !bytes.HasPrefix(hint.Start, []byte("ns/header")) {
		t.Fatalf("Start should include table prefix + sub-prefix, got %x", hint.Start)
	}
	if bytes.Compare(hint.Start, hint.Limit) >= 0 {
		t.Fatal("Limit must be greater than Start")
	}
}

func TestIncrementBytes(t *testing.T) {
	b := []byte{0x00, 0xFF}
	incrementBytes(b)
	if b[0] != 0x01 || b[1] != 0x00 {
		t.Fatalf("incrementBytes: got %x, want [0x01 0x00]", b)
	}

	b2 := []byte{0xFF}
	incrementBytes(b2)
	// All-0xFF wraps to 0x00.
	if b2[0] != 0x00 {
		t.Fatalf("incrementBytes wrap: got %x", b2)
	}
}

// --- Predefined table constructors ---

func TestNewChainDataTable(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewChainDataTable(db)
	if tbl.Prefix() != ChainDataNamespace {
		t.Fatalf("wrong prefix: %q", tbl.Prefix())
	}
}

func TestNewStateTrieTable(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewStateTrieTable(db)
	if tbl.Prefix() != StateTrieNamespace {
		t.Fatalf("wrong prefix: %q", tbl.Prefix())
	}
}

func TestNewTxIndexTable(t *testing.T) {
	db := NewMemoryDB()
	tbl := NewTxIndexTable(db)
	if tbl.Prefix() != TxIndexNamespace {
		t.Fatalf("wrong prefix: %q", tbl.Prefix())
	}
}

// --- TableDB ---

func TestTableDB_TableCreatedOnDemand(t *testing.T) {
	db := NewMemoryDB()
	tdb := NewTableDB(db)

	tbl := tdb.Table("chaindata/")
	if tbl == nil {
		t.Fatal("Table should return non-nil")
	}
}

func TestTableDB_SameTableReturnedTwice(t *testing.T) {
	db := NewMemoryDB()
	tdb := NewTableDB(db)

	t1 := tdb.Table("ns/")
	t2 := tdb.Table("ns/")
	if t1 != t2 {
		t.Fatal("expected the same *Table pointer on second call")
	}
}

func TestTableDB_Namespaces(t *testing.T) {
	db := NewMemoryDB()
	tdb := NewTableDB(db)
	tdb.Table("b/")
	tdb.Table("a/")
	tdb.Table("c/")

	ns := tdb.Namespaces()
	if len(ns) != 3 {
		t.Fatalf("expected 3 namespaces, got %d", len(ns))
	}
	if !sort.StringsAreSorted(ns) {
		t.Fatal("namespaces should be sorted")
	}
}

func TestTableDB_Close(t *testing.T) {
	db := NewMemoryDB()
	tdb := NewTableDB(db)
	if err := tdb.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func TestTableDB_DataSharedAcrossNamespaces(t *testing.T) {
	db := NewMemoryDB()
	tdb := NewTableDB(db)
	headers := tdb.Table("hdr/")
	bodies := tdb.Table("body/")

	headers.Put([]byte("1"), []byte("h1"))
	bodies.Put([]byte("1"), []byte("b1"))

	h, _ := headers.Get([]byte("1"))
	b, _ := bodies.Get([]byte("1"))
	if bytes.Equal(h, b) {
		t.Fatal("different namespaces should store different values")
	}
}
