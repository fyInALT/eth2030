package geth

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core"
)

// --- EFTestForkSupported tests ---

func TestEFTestForkSupported_KnownForks(t *testing.T) {
	known := []string{
		"Frontier",
		"Homestead",
		"EIP150",
		"EIP158",
		"SpuriousDragon",
		"TangerineWhistle",
		"Byzantium",
		"Constantinople",
		"ConstantinopleFix",
		"Istanbul",
		"Berlin",
		"London",
		"Merge",
		"Paris",
		"Shanghai",
		"Cancun",
		"Prague",
	}
	for _, fork := range known {
		if !EFTestForkSupported(fork) {
			t.Errorf("EFTestForkSupported(%q) = false, want true", fork)
		}
	}
}

func TestEFTestForkSupported_UnknownForks(t *testing.T) {
	unknown := []string{"", "unknown", "Glamsterdam", "Hogota", "IPlus", "mainnet"}
	for _, fork := range unknown {
		if EFTestForkSupported(fork) {
			t.Errorf("EFTestForkSupported(%q) = true, want false", fork)
		}
	}
}

// --- EFTestChainConfig tests ---

func TestEFTestChainConfig_Frontier(t *testing.T) {
	cfg, err := EFTestChainConfig("Frontier")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Frontier): %v", err)
	}
	if cfg == nil {
		t.Fatal("got nil config")
	}
	if cfg.ChainID == nil || cfg.ChainID.Int64() != 1 {
		t.Errorf("ChainID = %v, want 1", cfg.ChainID)
	}
	// Frontier has no block forks set.
	if cfg.HomesteadBlock != nil {
		t.Error("Frontier: HomesteadBlock should be nil")
	}
	if cfg.EIP150Block != nil {
		t.Error("Frontier: EIP150Block should be nil")
	}
}

func TestEFTestChainConfig_Homestead(t *testing.T) {
	cfg, err := EFTestChainConfig("Homestead")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Homestead): %v", err)
	}
	if cfg.HomesteadBlock == nil {
		t.Error("Homestead: HomesteadBlock should be set")
	}
	if cfg.EIP150Block != nil {
		t.Error("Homestead: EIP150Block should be nil")
	}
}

func TestEFTestChainConfig_EIP150(t *testing.T) {
	cfg, err := EFTestChainConfig("EIP150")
	if err != nil {
		t.Fatalf("EFTestChainConfig(EIP150): %v", err)
	}
	if cfg.EIP150Block == nil {
		t.Error("EIP150: EIP150Block should be set")
	}
	if cfg.EIP155Block != nil {
		t.Error("EIP150: EIP155Block should be nil (level 2 < 3)")
	}
}

func TestEFTestChainConfig_TangerineWhistle(t *testing.T) {
	// TangerineWhistle maps to level 2 (same as EIP150).
	cfg, err := EFTestChainConfig("TangerineWhistle")
	if err != nil {
		t.Fatalf("EFTestChainConfig(TangerineWhistle): %v", err)
	}
	if cfg.EIP150Block == nil {
		t.Error("TangerineWhistle: EIP150Block should be set")
	}
	if cfg.EIP155Block != nil {
		t.Error("TangerineWhistle: EIP155Block should be nil (level 2 < 3)")
	}
}

func TestEFTestChainConfig_EIP158(t *testing.T) {
	cfg, err := EFTestChainConfig("EIP158")
	if err != nil {
		t.Fatalf("EFTestChainConfig(EIP158): %v", err)
	}
	if cfg.EIP155Block == nil {
		t.Error("EIP158: EIP155Block should be set")
	}
	if cfg.EIP158Block == nil {
		t.Error("EIP158: EIP158Block should be set")
	}
	if cfg.ByzantiumBlock != nil {
		t.Error("EIP158: ByzantiumBlock should be nil (level 3 < 4)")
	}
}

func TestEFTestChainConfig_SpuriousDragon(t *testing.T) {
	// SpuriousDragon maps to level 3 (same as EIP158).
	cfg, err := EFTestChainConfig("SpuriousDragon")
	if err != nil {
		t.Fatalf("EFTestChainConfig(SpuriousDragon): %v", err)
	}
	if cfg.EIP158Block == nil {
		t.Error("SpuriousDragon: EIP158Block should be set")
	}
}

func TestEFTestChainConfig_Byzantium(t *testing.T) {
	cfg, err := EFTestChainConfig("Byzantium")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Byzantium): %v", err)
	}
	if cfg.ByzantiumBlock == nil {
		t.Error("Byzantium: ByzantiumBlock should be set")
	}
	if cfg.ConstantinopleBlock != nil {
		t.Error("Byzantium: ConstantinopleBlock should be nil (level 4 < 5)")
	}
}

func TestEFTestChainConfig_Constantinople(t *testing.T) {
	cfg, err := EFTestChainConfig("Constantinople")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Constantinople): %v", err)
	}
	if cfg.ConstantinopleBlock == nil {
		t.Error("Constantinople: ConstantinopleBlock should be set")
	}
	if cfg.PetersburgBlock == nil {
		t.Error("Constantinople: PetersburgBlock should be set")
	}
	if cfg.IstanbulBlock != nil {
		t.Error("Constantinople: IstanbulBlock should be nil (level 5 < 6)")
	}
}

func TestEFTestChainConfig_ConstantinopleFix(t *testing.T) {
	// ConstantinopleFix maps to level 5 (same as Constantinople).
	cfg, err := EFTestChainConfig("ConstantinopleFix")
	if err != nil {
		t.Fatalf("EFTestChainConfig(ConstantinopleFix): %v", err)
	}
	if cfg.ConstantinopleBlock == nil {
		t.Error("ConstantinopleFix: ConstantinopleBlock should be set")
	}
}

func TestEFTestChainConfig_Istanbul(t *testing.T) {
	cfg, err := EFTestChainConfig("Istanbul")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Istanbul): %v", err)
	}
	if cfg.IstanbulBlock == nil {
		t.Error("Istanbul: IstanbulBlock should be set")
	}
	if cfg.BerlinBlock != nil {
		t.Error("Istanbul: BerlinBlock should be nil (level 6 < 7)")
	}
}

func TestEFTestChainConfig_Berlin(t *testing.T) {
	cfg, err := EFTestChainConfig("Berlin")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Berlin): %v", err)
	}
	if cfg.BerlinBlock == nil {
		t.Error("Berlin: BerlinBlock should be set")
	}
	if cfg.LondonBlock != nil {
		t.Error("Berlin: LondonBlock should be nil (level 7 < 8)")
	}
}

func TestEFTestChainConfig_London(t *testing.T) {
	cfg, err := EFTestChainConfig("London")
	if err != nil {
		t.Fatalf("EFTestChainConfig(London): %v", err)
	}
	if cfg.LondonBlock == nil {
		t.Error("London: LondonBlock should be set")
	}
	if cfg.TerminalTotalDifficulty != nil {
		t.Error("London: TerminalTotalDifficulty should be nil (level 8 < 9)")
	}
}

func TestEFTestChainConfig_Merge(t *testing.T) {
	cfg, err := EFTestChainConfig("Merge")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Merge): %v", err)
	}
	if cfg.TerminalTotalDifficulty == nil {
		t.Error("Merge: TerminalTotalDifficulty should be set")
	}
	if cfg.ShanghaiTime != nil {
		t.Error("Merge: ShanghaiTime should be nil (level 9 < 10)")
	}
}

func TestEFTestChainConfig_Paris(t *testing.T) {
	// Paris maps to level 9 (same as Merge).
	cfg, err := EFTestChainConfig("Paris")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Paris): %v", err)
	}
	if cfg.TerminalTotalDifficulty == nil {
		t.Error("Paris: TerminalTotalDifficulty should be set")
	}
}

func TestEFTestChainConfig_Shanghai(t *testing.T) {
	cfg, err := EFTestChainConfig("Shanghai")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Shanghai): %v", err)
	}
	if cfg.ShanghaiTime == nil {
		t.Error("Shanghai: ShanghaiTime should be set")
	}
	if *cfg.ShanghaiTime != 0 {
		t.Errorf("Shanghai: ShanghaiTime = %d, want 0", *cfg.ShanghaiTime)
	}
	if cfg.CancunTime != nil {
		t.Error("Shanghai: CancunTime should be nil (level 10 < 11)")
	}
}

func TestEFTestChainConfig_Cancun(t *testing.T) {
	cfg, err := EFTestChainConfig("Cancun")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Cancun): %v", err)
	}
	if cfg.CancunTime == nil {
		t.Error("Cancun: CancunTime should be set")
	}
	if *cfg.CancunTime != 0 {
		t.Errorf("Cancun: CancunTime = %d, want 0", *cfg.CancunTime)
	}
	if cfg.PragueTime != nil {
		t.Error("Cancun: PragueTime should be nil (level 11 < 12)")
	}
}

func TestEFTestChainConfig_Prague(t *testing.T) {
	cfg, err := EFTestChainConfig("Prague")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Prague): %v", err)
	}
	if cfg.PragueTime == nil {
		t.Error("Prague: PragueTime should be set")
	}
	if *cfg.PragueTime != 0 {
		t.Errorf("Prague: PragueTime = %d, want 0", *cfg.PragueTime)
	}
}

func TestEFTestChainConfig_Unsupported(t *testing.T) {
	_, err := EFTestChainConfig("Nonexistent")
	if err == nil {
		t.Error("expected error for unsupported fork, got nil")
	}
}

func TestEFTestChainConfig_CumulativeForks(t *testing.T) {
	// Prague includes all forks up through level 12.
	cfg, err := EFTestChainConfig("Prague")
	if err != nil {
		t.Fatalf("EFTestChainConfig(Prague): %v", err)
	}
	if cfg.HomesteadBlock == nil {
		t.Error("Prague: HomesteadBlock should be set (cumulative)")
	}
	if cfg.EIP150Block == nil {
		t.Error("Prague: EIP150Block should be set (cumulative)")
	}
	if cfg.ByzantiumBlock == nil {
		t.Error("Prague: ByzantiumBlock should be set (cumulative)")
	}
	if cfg.LondonBlock == nil {
		t.Error("Prague: LondonBlock should be set (cumulative)")
	}
	if cfg.ShanghaiTime == nil {
		t.Error("Prague: ShanghaiTime should be set (cumulative)")
	}
	if cfg.CancunTime == nil {
		t.Error("Prague: CancunTime should be set (cumulative)")
	}
}

// --- ToGethChainConfig tests ---

func TestToGethChainConfig_Nil(t *testing.T) {
	result := ToGethChainConfig(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestToGethChainConfig_Basic(t *testing.T) {
	ts := uint64(1000)
	c := &core.ChainConfig{
		ChainID:        big.NewInt(5),
		HomesteadBlock: big.NewInt(0),
		LondonBlock:    big.NewInt(10),
		ShanghaiTime:   &ts,
		CancunTime:     &ts,
	}

	gc := ToGethChainConfig(c)
	if gc == nil {
		t.Fatal("got nil config")
	}
	if gc.ChainID.Int64() != 5 {
		t.Errorf("ChainID = %d, want 5", gc.ChainID.Int64())
	}
	if gc.HomesteadBlock == nil || gc.HomesteadBlock.Int64() != 0 {
		t.Errorf("HomesteadBlock = %v, want 0", gc.HomesteadBlock)
	}
	if gc.LondonBlock == nil || gc.LondonBlock.Int64() != 10 {
		t.Errorf("LondonBlock = %v, want 10", gc.LondonBlock)
	}
	if gc.ShanghaiTime == nil || *gc.ShanghaiTime != 1000 {
		t.Errorf("ShanghaiTime = %v, want 1000", gc.ShanghaiTime)
	}
	if gc.CancunTime == nil || *gc.CancunTime != 1000 {
		t.Errorf("CancunTime = %v, want 1000", gc.CancunTime)
	}
}

func TestToGethChainConfig_AllBlockForks(t *testing.T) {
	c := &core.ChainConfig{
		ChainID:                 big.NewInt(1),
		HomesteadBlock:          big.NewInt(1),
		EIP150Block:             big.NewInt(2),
		EIP155Block:             big.NewInt(3),
		EIP158Block:             big.NewInt(3),
		ByzantiumBlock:          big.NewInt(4),
		ConstantinopleBlock:     big.NewInt(5),
		PetersburgBlock:         big.NewInt(5),
		IstanbulBlock:           big.NewInt(6),
		BerlinBlock:             big.NewInt(7),
		LondonBlock:             big.NewInt(8),
		TerminalTotalDifficulty: big.NewInt(0),
	}

	gc := ToGethChainConfig(c)
	if gc.HomesteadBlock.Int64() != 1 {
		t.Errorf("HomesteadBlock = %d, want 1", gc.HomesteadBlock.Int64())
	}
	if gc.EIP150Block.Int64() != 2 {
		t.Errorf("EIP150Block = %d, want 2", gc.EIP150Block.Int64())
	}
	if gc.ByzantiumBlock.Int64() != 4 {
		t.Errorf("ByzantiumBlock = %d, want 4", gc.ByzantiumBlock.Int64())
	}
	if gc.ConstantinopleBlock.Int64() != 5 {
		t.Errorf("ConstantinopleBlock = %d, want 5", gc.ConstantinopleBlock.Int64())
	}
	if gc.IstanbulBlock.Int64() != 6 {
		t.Errorf("IstanbulBlock = %d, want 6", gc.IstanbulBlock.Int64())
	}
	if gc.BerlinBlock.Int64() != 7 {
		t.Errorf("BerlinBlock = %d, want 7", gc.BerlinBlock.Int64())
	}
	if gc.LondonBlock.Int64() != 8 {
		t.Errorf("LondonBlock = %d, want 8", gc.LondonBlock.Int64())
	}
	if gc.TerminalTotalDifficulty == nil {
		t.Error("TerminalTotalDifficulty should be set")
	}
}

func TestToGethChainConfig_PragueTime(t *testing.T) {
	ts := uint64(9999)
	c := &core.ChainConfig{
		ChainID:    big.NewInt(1),
		PragueTime: &ts,
	}
	gc := ToGethChainConfig(c)
	if gc.PragueTime == nil || *gc.PragueTime != 9999 {
		t.Errorf("PragueTime = %v, want 9999", gc.PragueTime)
	}
}
