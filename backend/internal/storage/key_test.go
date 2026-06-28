package storage

import "testing"

func TestSafeKeyAllowsMoneyCreekAssetPrefix(t *testing.T) {
	key, err := safeKey("money-creek/assets/asset-id/original/photo.jpg")
	if err != nil {
		t.Fatalf("safeKey returned error: %v", err)
	}
	if filepathKey(key) != "money-creek/assets/asset-id/original/photo.jpg" {
		t.Fatalf("unexpected key: %q", key)
	}
}

func TestSafeKeyRejectsAbsoluteAndTraversal(t *testing.T) {
	cases := []string{"/money-creek/assets/a/original/x.jpg", "money-creek/../x.jpg", "..\\x.jpg"}
	for _, tc := range cases {
		if _, err := safeKey(tc); err == nil {
			t.Fatalf("expected %q to be rejected", tc)
		}
	}
}
