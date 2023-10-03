package service

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Time struct {
	time.Time
}

func NewTime() Time {
	return Time{time.Now().UTC()}
}

func (t *Time) MarshalBSON() ([]byte, error) {
	type RawTime Time

	if t == nil {
		return bson.Marshal((RawTime)(NewTime()))
	}

	return bson.Marshal((RawTime)(*t))
}
