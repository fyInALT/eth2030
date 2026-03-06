package ethversion

import "testing"

func TestProtocolVersionString(t *testing.T) {
	if got := ETH68Version.String(); got != "ETH/68" {
		t.Errorf("String() = %q, want ETH/68", got)
	}
}

func TestProtocolVersionEqual(t *testing.T) {
	if !ETH68Version.Equal(ETH68Version) {
		t.Error("equal versions should be equal")
	}
	if ETH68Version.Equal(ETH67Version) {
		t.Error("different versions should not be equal")
	}
}

func TestProtocolVersionLess(t *testing.T) {
	if !ETH67Version.Less(ETH68Version) {
		t.Error("ETH67 should be less than ETH68")
	}
	if ETH68Version.Less(ETH67Version) {
		t.Error("ETH68 should not be less than ETH67")
	}
}

func TestNegotiateVersion(t *testing.T) {
	vm := NewVersionManager([]ProtocolVersion{ETH66Version, ETH67Version, ETH68Version})

	// Highest common version.
	got, err := vm.NegotiateVersion([]ProtocolVersion{ETH67Version, ETH68Version})
	if err != nil {
		t.Fatalf("NegotiateVersion: %v", err)
	}
	if !got.Equal(ETH68Version) {
		t.Errorf("got %v, want ETH68", got)
	}

	// No common version.
	_, err = vm.NegotiateVersion([]ProtocolVersion{{Major: 99}})
	if err != ErrNoCommonVersion {
		t.Errorf("expected ErrNoCommonVersion, got %v", err)
	}

	// Empty peer versions.
	_, err = vm.NegotiateVersion(nil)
	if err != ErrNoVersions {
		t.Errorf("expected ErrNoVersions, got %v", err)
	}
}

func TestVersionManager_PeerTracking(t *testing.T) {
	vm := NewVersionManager([]ProtocolVersion{ETH68Version})

	vm.RegisterPeer("peer1", ETH68Version)
	got := vm.GetPeerVersion("peer1")
	if got == nil || !got.Equal(ETH68Version) {
		t.Errorf("GetPeerVersion = %v, want ETH68", got)
	}

	vm.RemovePeer("peer1")
	if vm.GetPeerVersion("peer1") != nil {
		t.Error("peer should be removed")
	}
}
