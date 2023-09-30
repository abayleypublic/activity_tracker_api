package datetime

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	timeFormat = time.RFC3339
)

type DateTime string
type RawDateTime time.Time

func New() DateTime {
	return DateTime(time.Now().UTC().Format(timeFormat))
}

func (dt *DateTime) MarshalBSON() ([]byte, error) {

	t, err := time.Parse(timeFormat, string(*dt))
	if err != nil {
		return nil, err
	}

	return bson.Marshal(RawDateTime(t))
}

func (dt *DateTime) UnmarshalBSON(b []byte) error {
	t := RawDateTime{}
	err := bson.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	*dt = DateTime(time.Time(t).Format(timeFormat))

	return nil
}
