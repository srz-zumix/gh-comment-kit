package reviewer

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
)

type MetaData struct {
	Hash       string `json:"hash"`
	URL        string `json:"url,omitempty"`
	Index      int    `json:"index"`
	Group      string `json:"group"`
	TotalParts int    `json:"total_parts,omitempty"`
	PartNumber int    `json:"part_number,omitempty"`
}

func CreateMetaData(source any, index int, group string, url string) MetaData {
	sha256sum := sha256.Sum256([]byte(fmt.Sprintf("%v", source)))
	return MetaData{
		Hash:  fmt.Sprintf("%x", sha256sum),
		URL:   url,
		Index: index,
		Group: group,
	}
}

func ParseMetaData(commentBody string) (*MetaData, string, error) {
	const marker = "<!--gh-comment-kit "
	start := strings.Index(commentBody, marker)
	if start == -1 {
		return nil, commentBody, fmt.Errorf("no meta data found")
	}
	const endMarker = "-->"
	end := strings.Index(commentBody[start:], endMarker)
	if end == -1 {
		return nil, commentBody, fmt.Errorf("no meta data end found")
	}
	jsonStr := commentBody[start+len(marker) : start+end]
	body := commentBody[:start]
	body += commentBody[start+end+len(endMarker):]
	var meta MetaData
	if err := json.Unmarshal([]byte(jsonStr), &meta); err != nil {
		return nil, body, err
	}
	return &meta, body, nil
}

func (m MetaData) ToHTML() string {
	jsonStr, _ := json.Marshal(m)
	return fmt.Sprintf("<!--gh-comment-kit %s -->", jsonStr)
}
