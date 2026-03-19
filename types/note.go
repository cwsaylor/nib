package types

import "time"

type Note struct {
	ID        int
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Preview returns a short preview of the note body.
func (n Note) Preview() string {
	if len(n.Body) > 120 {
		return n.Body[:120]
	}
	return n.Body
}
