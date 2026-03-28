package golang

import (
	"encoding/base64"
	"reflect"
)

func (ipc *ipcCommon) serialize(arg any) any {
	t := reflect.TypeOf(arg)
	switch t.Kind() {
	case reflect.Slice:
		switch t.Elem().Name() {
		case "uint8":
			return map[string]any{
				"t": "blob",
				"d": base64.StdEncoding.EncodeToString(arg.([]byte)),
			}
		}
	}
	return arg
}

func (ipc *ipcCommon) ConvType(needType reflect.Type, gotType reflect.Type, arg any) any {
	switch needType.Kind() {
	case reflect.Int:
		// JSON decodes any number to float64. If we need int, we should check and convert
		if gotType.Kind() == reflect.Float64 {
			floatArg := arg.(float64)
			if float64(int64(floatArg)) == floatArg && !needType.OverflowInt(int64(floatArg)) {
				arg = int(floatArg)
			}
		}
	}
	return arg
}
