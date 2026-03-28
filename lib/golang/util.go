package golang

import (
	"fmt"
	"reflect"
)

func mergeErr(errs ...error) (ret error) {
	for _, err := range errs {
		if err != nil {
			if ret == nil {
				ret = err
			} else {
				ret = fmt.Errorf("%w; %w", ret, err)
			}
		}
	}
	return
}

func mapTypeNames(types []any) map[string]any {
	result := make(map[string]any)
	for _, t := range types {
		if reflect.TypeOf(t).Kind() != reflect.Pointer {
			panic(fmt.Sprintf("LocalAPI argument must be pointer"))
		}
		typeName := reflect.TypeOf(t).Elem().Name()
		result[typeName] = t
	}
	return result
}
