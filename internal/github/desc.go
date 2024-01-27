package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func queryDescription(user, repo string, log io.Writer) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo)
	fmt.Fprintln(log, "Get:", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

type Description struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	License     map[string]string `json:"license"`
}

func GetDescription(user, repo string, log io.Writer) (*Description, error) {
	bin, err := queryDescription(user, repo, log)
	if err != nil {
		return nil, err
	}
	var desc *Description
	if err = json.Unmarshal(bin, &desc); err != nil {
		return nil, err
	}
	return desc, nil
}
