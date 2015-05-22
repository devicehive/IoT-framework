package rest

import "github.com/devicehive/IoT-framework/devicehive-cloud/param"

//Example timestamp: 2015-05-21T14:18:34.019584

const TimestampKey = "timestamp"

func TimestampParam(stamp string) param.String {
	return param.String{Arg: TimestampKey, Value: stamp}
}
