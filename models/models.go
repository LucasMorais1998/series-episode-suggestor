package models

type Episode struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Season   int    `json:"season"`
    Number   int    `json:"number"`
}