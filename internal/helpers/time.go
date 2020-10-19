package helpers

import (
	"time"

	"github.com/gogo/protobuf/types"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
)

func GetEndTime(timestamp *types.Timestamp, player *pb.Player) (t time.Time, err error) {
	// TODO: Recuperare location da info player
	var location *time.Location
	if location, err = time.LoadLocation("Europe/Rome"); err != nil {
		return time.Time{}, err
	}

	// Converto timestamp ricevuto
	if t, err = types.TimestampFromProto(timestamp); err != nil {
		return time.Time{}, err
	}

	return t.In(location), nil
}
