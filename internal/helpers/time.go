package helpers

import (
	"time"

	"github.com/gogo/protobuf/types"

	"nn-grpc/build/pb"
)

func GetEndTime(timestamp *types.Timestamp, player *pb.Player) (t time.Time, err error) {
	var location *time.Location
	if location, err = time.LoadLocation(player.GetTimezone().GetName()); err != nil {
		return time.Time{}, err
	}

	// Converto timestamp ricevuto
	if t, err = types.TimestampFromProto(timestamp); err != nil {
		return time.Time{}, err
	}

	return t.In(location), nil
}
