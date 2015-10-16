// TODO: seems this file is unused - consider to delete!

package rest

import (
	"time"

	"github.com/devicehive/IoT-framework/devicehive-cloud/param"
)

//Example timestamp: 2015-05-21T14:18:34.019584

const (
	TimestampKey    = "timestamp"
	TimestampLayout = "2006-01-02T15:04:05.999999999"
)

func TimestampParam(stamp string) param.String {
	return param.String{Arg: TimestampKey, Value: stamp}
}

// adding nanosecond
func NearestFutureForTimestamp(stamp string) (future string, err error) {
	t, err := time.Parse(TimestampLayout, stamp)
	if err != nil {
		return
	}

	f := t.Add(time.Nanosecond)
	future = f.Format(TimestampLayout)
	return
}
