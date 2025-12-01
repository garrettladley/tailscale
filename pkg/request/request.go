package request

import "time"

type Request struct {
	ID        string
	Timestamp time.Time
	Payload   []byte

	TiedTo string
}
