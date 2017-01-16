package common

import (
	"testing"
)

func TestUUID_alwaysgood(t *testing.T) {
	uuid, _ := NewUUID()
	if uuid == "" {
		t.Errorf("Failed to return UUID as expected")
	}
}

func TestGetMD5Hash(t *testing.T) {
	hash1 := GetMD5Hash("This is a test")
	hash2 := GetMD5Hash("This is a test")
	if hash1 != hash2 {
		t.Errorf("hashs and hash2 should be the same, but were not")
	}

	hash1 = GetMD5Hash("This is a test")
	hash2 = GetMD5Hash("This is a tfst")
	if hash1 == hash2 {
		t.Errorf("hashs and hash2 are the same, but should not be")
	}
}
