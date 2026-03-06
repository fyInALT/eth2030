package transport

import (
	"testing"
	"time"
)

// --- BB-1.1: MixnetTransportMode parsing ---

func TestMixnetTransportMode_ParseSimulated(t *testing.T) {
	m, err := ParseMixnetMode("simulated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != ModeSimulated {
		t.Fatalf("expected ModeSimulated, got %v", m)
	}
}

func TestMixnetTransportMode_ParseEmpty(t *testing.T) {
	m, err := ParseMixnetMode("")
	if err != nil {
		t.Fatalf("unexpected error for empty string: %v", err)
	}
	if m != ModeSimulated {
		t.Fatalf("expected ModeSimulated for empty, got %v", m)
	}
}

func TestMixnetTransportMode_ParseTor(t *testing.T) {
	m, err := ParseMixnetMode("tor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != ModeTorSocks5 {
		t.Fatalf("expected ModeTorSocks5, got %v", m)
	}
}

func TestMixnetTransportMode_ParseNym(t *testing.T) {
	m, err := ParseMixnetMode("nym")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != ModeNymSocks5 {
		t.Fatalf("expected ModeNymSocks5, got %v", m)
	}
}

func TestMixnetTransportMode_ParseInvalid(t *testing.T) {
	_, err := ParseMixnetMode("flashnet")
	if err == nil {
		t.Fatal("expected error for unknown mode 'flashnet'")
	}
}

func TestMixnetTransportMode_String(t *testing.T) {
	cases := []struct {
		mode MixnetTransportMode
		want string
	}{
		{ModeSimulated, "simulated"},
		{ModeTorSocks5, "tor"},
		{ModeNymSocks5, "nym"},
	}
	for _, tc := range cases {
		if got := tc.mode.String(); got != tc.want {
			t.Errorf("mode %d: got %q, want %q", tc.mode, got, tc.want)
		}
	}
}

// --- BB-1.1: TransportConfig ---

func TestTransportConfig_Defaults(t *testing.T) {
	cfg := DefaultTransportConfig()
	if cfg.Mode != ModeSimulated {
		t.Fatalf("expected ModeSimulated, got %v", cfg.Mode)
	}
	if cfg.TorProxyAddr != "127.0.0.1:9050" {
		t.Fatalf("expected default tor addr '127.0.0.1:9050', got %q", cfg.TorProxyAddr)
	}
	if cfg.NymProxyAddr != "127.0.0.1:1080" {
		t.Fatalf("expected default nym addr '127.0.0.1:1080', got %q", cfg.NymProxyAddr)
	}
	if cfg.DialTimeout != 500*time.Millisecond {
		t.Fatalf("expected 500ms dial timeout, got %v", cfg.DialTimeout)
	}
	if cfg.KohakuCompatible {
		t.Fatal("expected KohakuCompatible=false by default")
	}
}

// --- BB-1.1: TransportManager config methods ---

func TestTransportManager_NewWithConfig(t *testing.T) {
	cfg := DefaultTransportConfig()
	cfg.Mode = ModeTorSocks5
	tm := NewTransportManagerWithConfig(cfg)
	if tm.Config().Mode != ModeTorSocks5 {
		t.Fatalf("expected ModeTorSocks5, got %v", tm.Config().Mode)
	}
}

func TestTransportManager_SelectedMode_Default(t *testing.T) {
	tm := NewTransportManager()
	if tm.SelectedMode() != ModeSimulated {
		t.Fatalf("expected ModeSimulated, got %v", tm.SelectedMode())
	}
}

func TestTransportManager_SetSelectedMode(t *testing.T) {
	tm := NewTransportManager()
	tm.setSelectedMode(ModeTorSocks5)
	if tm.SelectedMode() != ModeTorSocks5 {
		t.Fatalf("expected ModeTorSocks5, got %v", tm.SelectedMode())
	}
}

// --- BB-1.4: Kohaku compatibility ---

func TestKohakuCompatibility_FormatJSON(t *testing.T) {
	msg := FormatControlMessage("test-msg", false)
	if len(msg) == 0 {
		t.Fatal("expected non-empty message")
	}
	// JSON format starts with '{'.
	if msg[0] != '{' {
		t.Fatalf("expected JSON format (starts with '{'), got 0x%02x", msg[0])
	}
}

func TestKohakuCompatibility_FormatKohaku(t *testing.T) {
	msg := FormatControlMessage("test-msg", true)
	if len(msg) < 4 {
		t.Fatalf("kohaku format requires at least 4 bytes, got %d", len(msg))
	}
	// Kohaku: 4-byte big-endian length prefix.
	payloadLen := int(msg[0])<<24 | int(msg[1])<<16 | int(msg[2])<<8 | int(msg[3])
	if payloadLen != len(msg)-4 {
		t.Fatalf("kohaku length prefix %d does not match payload %d", payloadLen, len(msg)-4)
	}
}

func TestKohakuCompatibility_JSONvsBinary(t *testing.T) {
	jsonMsg := FormatControlMessage("hello", false)
	kohakuMsg := FormatControlMessage("hello", true)
	if string(jsonMsg) == string(kohakuMsg) {
		t.Fatal("JSON and Kohaku formats should differ")
	}
}

func TestKohakuCompatibility_Deterministic(t *testing.T) {
	a := FormatControlMessage("deterministic", true)
	b := FormatControlMessage("deterministic", true)
	if string(a) != string(b) {
		t.Fatal("FormatControlMessage should be deterministic")
	}
}
