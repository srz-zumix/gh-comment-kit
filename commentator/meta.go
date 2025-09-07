package commentator

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type MetaData struct {
	Hash  string `json:"hash"`
	URL   string `json:"url"`
	Index int    `json:"index"`
	Group string `json:"group,omitempty"`
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

func ParseMetaData(commentBody string) (*MetaData, error) {
	var jsonStr string
	var meta MetaData
	n, err := fmt.Sscanf(commentBody, "<!--gh-commentator %s-->", &jsonStr)
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, fmt.Errorf("no meta data found")
	}
	err = json.Unmarshal([]byte(jsonStr), &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func (m MetaData) ToHTML() string {
	jsonStr, _ := json.Marshal(m)
	return fmt.Sprintf("<!--gh-commentator %s-->", jsonStr)
}
