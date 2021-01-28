package items

import "time"

//openapi:schema Item
//openapi:schema Response

type Item struct {
	Kind    string    `json:"kind"`
	Created time.Time `json:"created_at"`
	Items   []Item    `json:"items"`
}

type Response struct {
	Items []Item   `json:"items"`
	Links []string `json:"links"`
}
