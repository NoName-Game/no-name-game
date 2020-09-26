package helpers

import (
	"time"

	"github.com/golang/protobuf/ptypes"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func GetEndTime(timestamp *timestamp.Timestamp, player *pb.Player) (t time.Time, err error) {
	// TODO: Recuperare location da info player
	var location *time.Location
	if location, err = time.LoadLocation("Europe/Rome"); err != nil {
		return time.Time{}, err
	}

	// Converto timestamp ricevuto
	if t, err = ptypes.Timestamp(timestamp); err != nil {
		return time.Time{}, err
	}

	return t.In(location), nil
}
