package service

import (
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type Time time.Time

func NewTime() Time {
	return Time(time.Now().UTC())
}

func (t *Time) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if t == nil {
		return bson.MarshalValue(time.Time(NewTime()))
	}

	return bson.MarshalValue(time.Time(*t))
}

func (t *Time) UnmarshalBSONValue(bt bsontype.Type, b []byte) error {
	if bt != bsontype.DateTime {
		return fmt.Errorf("invalid bson value type '%s'", bt.String())
	}

	s, _, ok := bsoncore.ReadDateTime(b)
	if !ok {
		return fmt.Errorf("invalid bson datetime value")
	}

	*t = Time(time.UnixMilli(s))
	return nil
}

func (t *Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*t))
}

func (t *Time) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, t)
}
