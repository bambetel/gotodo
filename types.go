package main

import "time"

type Todo struct {
	Id        int       `json:"id"`
	Label     string    `json:"label"`
	Priority  int       `json:"priority"`
	Modified  time.Time `json:"modified"`
	Created   time.Time `json:"created"`
	Progress  int       `json:"progress"`
	Completed bool      `json:"completed"`
	Tags      string    `json:"tags"`
}
