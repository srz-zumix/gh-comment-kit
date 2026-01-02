package reviewer

import (
	"testing"
)

func TestCreateMetaData(t *testing.T) {
	tests := []struct {
		name   string
		source any
		index  int
		group  string
		url    string
	}{
		{
			name:   "basic string source",
			source: "test source",
			index:  1,
			group:  "test-group",
			url:    "https://example.com",
		},
		{
			name:   "numeric source",
			source: 12345,
			index:  2,
			group:  "numeric-group",
			url:    "https://example.com/2",
		},
		{
			name:   "empty url",
			source: "test",
			index:  3,
			group:  "group",
			url:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := CreateMetaData(tt.source, tt.index, tt.group, tt.url)

			if meta.Hash == "" {
				t.Error("Hash should not be empty")
			}
			if len(meta.Hash) != 64 { // SHA256 hash is 64 hex characters
				t.Errorf("Hash length = %d, want 64", len(meta.Hash))
			}
			if meta.Index != tt.index {
				t.Errorf("Index = %d, want %d", meta.Index, tt.index)
			}
			if meta.Group != tt.group {
				t.Errorf("Group = %s, want %s", meta.Group, tt.group)
			}
			if meta.URL != tt.url {
				t.Errorf("URL = %s, want %s", meta.URL, tt.url)
			}
		})
	}
}

func TestCreateMetaData_HashConsistency(t *testing.T) {
	source := "consistent source"
	meta1 := CreateMetaData(source, 1, "group", "url")
	meta2 := CreateMetaData(source, 1, "group", "url")

	if meta1.Hash != meta2.Hash {
		t.Errorf("Hash should be consistent for same source, got %s and %s", meta1.Hash, meta2.Hash)
	}
}

func TestCreateMetaData_DifferentSources(t *testing.T) {
	meta1 := CreateMetaData("source1", 1, "group", "url")
	meta2 := CreateMetaData("source2", 1, "group", "url")

	if meta1.Hash == meta2.Hash {
		t.Error("Hash should be different for different sources")
	}
}

func TestMetaData_ToHTML(t *testing.T) {
	tests := []struct {
		name string
		meta MetaData
		want string
	}{
		{
			name: "complete metadata",
			meta: MetaData{
				Hash:  "abc123",
				URL:   "https://example.com",
				Index: 1,
				Group: "test-group",
			},
			want: `<!--gh-comment-kit {"hash":"abc123","url":"https://example.com","index":1,"group":"test-group"} -->`,
		},
		{
			name: "metadata without url",
			meta: MetaData{
				Hash:  "def456",
				Index: 2,
				Group: "group2",
			},
			want: `<!--gh-comment-kit {"hash":"def456","index":2,"group":"group2"} -->`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.meta.ToHTML()
			if got != tt.want {
				t.Errorf("ToHTML() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestParseMetaData(t *testing.T) {
	tests := []struct {
		name        string
		commentBody string
		wantMeta    *MetaData
		wantBody    string
		wantErr     bool
	}{
		{
			name:        "valid metadata at end",
			commentBody: "This is a comment\n<!--gh-comment-kit {\"hash\":\"abc123\",\"url\":\"https://example.com\",\"index\":1,\"group\":\"test-group\"} -->",
			wantMeta: &MetaData{
				Hash:  "abc123",
				URL:   "https://example.com",
				Index: 1,
				Group: "test-group",
			},
			wantBody: "This is a comment\n",
			wantErr:  false,
		},
		{
			name:        "valid metadata at beginning",
			commentBody: "<!--gh-comment-kit {\"hash\":\"def456\",\"index\":2,\"group\":\"group2\"} -->This is content",
			wantMeta: &MetaData{
				Hash:  "def456",
				Index: 2,
				Group: "group2",
			},
			wantBody: "This is content",
			wantErr:  false,
		},
		{
			name:        "valid metadata in middle",
			commentBody: "Before\n<!--gh-comment-kit {\"hash\":\"ghi789\",\"index\":3,\"group\":\"group3\"} -->\nAfter",
			wantMeta: &MetaData{
				Hash:  "ghi789",
				Index: 3,
				Group: "group3",
			},
			wantBody: "Before\n\nAfter",
			wantErr:  false,
		},
		{
			name:        "no metadata",
			commentBody: "Just a regular comment",
			wantMeta:    nil,
			wantBody:    "Just a regular comment",
			wantErr:     true,
		},
		{
			name:        "incomplete metadata marker",
			commentBody: "<!--gh-comment-kit {\"hash\":\"abc123\",\"index\":1,\"group\":\"test\"}",
			wantMeta:    nil,
			wantBody:    "<!--gh-comment-kit {\"hash\":\"abc123\",\"index\":1,\"group\":\"test\"}",
			wantErr:     true,
		},
		{
			name:        "invalid json",
			commentBody: "<!--gh-comment-kit invalid json -->",
			wantMeta:    nil,
			wantBody:    "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMeta, gotBody, err := ParseMetaData(tt.commentBody)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMetaData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotBody != tt.wantBody {
				t.Errorf("ParseMetaData() gotBody = %q, want %q", gotBody, tt.wantBody)
			}

			if tt.wantErr {
				return
			}

			if gotMeta == nil && tt.wantMeta != nil {
				t.Error("ParseMetaData() gotMeta = nil, want non-nil")
				return
			}

			if gotMeta != nil && tt.wantMeta == nil {
				t.Error("ParseMetaData() gotMeta = non-nil, want nil")
				return
			}

			if gotMeta != nil && tt.wantMeta != nil {
				if gotMeta.Hash != tt.wantMeta.Hash {
					t.Errorf("ParseMetaData() Hash = %s, want %s", gotMeta.Hash, tt.wantMeta.Hash)
				}
				if gotMeta.URL != tt.wantMeta.URL {
					t.Errorf("ParseMetaData() URL = %s, want %s", gotMeta.URL, tt.wantMeta.URL)
				}
				if gotMeta.Index != tt.wantMeta.Index {
					t.Errorf("ParseMetaData() Index = %d, want %d", gotMeta.Index, tt.wantMeta.Index)
				}
				if gotMeta.Group != tt.wantMeta.Group {
					t.Errorf("ParseMetaData() Group = %s, want %s", gotMeta.Group, tt.wantMeta.Group)
				}
			}
		})
	}
}

func TestParseMetaData_RoundTrip(t *testing.T) {
	// Test that ToHTML and ParseMetaData are inverse operations
	originalMeta := MetaData{
		Hash:  "test-hash-123",
		URL:   "https://example.com/test",
		Index: 42,
		Group: "round-trip-group",
	}

	commentBody := "Test comment content\n" + originalMeta.ToHTML()

	parsedMeta, body, err := ParseMetaData(commentBody)
	if err != nil {
		t.Fatalf("ParseMetaData() error = %v", err)
	}

	if body != "Test comment content\n" {
		t.Errorf("Body = %q, want %q", body, "Test comment content\n")
	}

	if parsedMeta.Hash != originalMeta.Hash {
		t.Errorf("Hash = %s, want %s", parsedMeta.Hash, originalMeta.Hash)
	}
	if parsedMeta.URL != originalMeta.URL {
		t.Errorf("URL = %s, want %s", parsedMeta.URL, originalMeta.URL)
	}
	if parsedMeta.Index != originalMeta.Index {
		t.Errorf("Index = %d, want %d", parsedMeta.Index, originalMeta.Index)
	}
	if parsedMeta.Group != originalMeta.Group {
		t.Errorf("Group = %s, want %s", parsedMeta.Group, originalMeta.Group)
	}
}
