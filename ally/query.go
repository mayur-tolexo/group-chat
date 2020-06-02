package ally

import (
	"errors"
	"reflect"
	"strings"
	"time"
)

//MergeMap will merge two maps
func MergeMap(a, b map[string]interface{}) {
	for k, v := range b {
		a[k] = v
	}
}

//StructTagValue will return struct model tag field and value from specific tag
func StructTagValue(input interface{}, tag string) (fields map[string]interface{}, err error) {
	if tag == "" {
		tag = "sql"
	}
	refObj := reflect.ValueOf(input)
	fields = make(map[string]interface{})
	if refObj.Kind() == reflect.Ptr {
		refObj = refObj.Elem()
	}
	if refObj.IsValid() {
		for i := 0; i < refObj.NumField(); i++ {
			refField := refObj.Field(i)
			refType := refObj.Type().Field(i)
			if refType.Name[0] > 'Z' {
				continue
			}
			if refType.Anonymous && refField.Kind() == reflect.Struct {
				var embdFields map[string]interface{}
				if embdFields, err = StructTagValue(refField.Interface(), tag); err != nil {
					break
				}
				MergeMap(fields, embdFields)
			} else {
				if col, exists := refType.Tag.Lookup(tag); exists {
					isDef := IsDefaultVal(refField)
					if col == "-" ||
						(strings.Contains(col, ",omitempty") && isDef) {
						continue
					}
					if tag == "gorm" {
						if sqlCol := refType.Tag.Get("sql"); sqlCol == "-" {
							continue
						}
						col = strings.TrimPrefix(col, "column:")
					}
					col = strings.Split(col, ",")[0]
					dVal := refType.Tag.Get("default")
					if dVal == "null" && isDef {
						fields[col] = nil
					} else if fields[col], err = GetFieldVal(refField); err != nil {
						break
					}
				}
			}
		}
	}
	return
}

//IsDefaultVal Check whether zero value or not
func IsDefaultVal(curFieldVal reflect.Value) (isDefault bool) {
	defaultValue := reflect.Zero(curFieldVal.Type())
	fieldKind := curFieldVal.Kind()
	if curTime, isTimeType := curFieldVal.Interface().(time.Time); isTimeType {
		zeroTime, _ := defaultValue.Interface().(time.Time)
		isDefault = (curTime.After(zeroTime) == false)
	} else if fieldKind == reflect.Struct || fieldKind == reflect.Ptr {
		isDefault = reflect.DeepEqual(curFieldVal, defaultValue)
	} else if fieldKind == reflect.Map || fieldKind == reflect.Slice {
		if curFieldVal.Len() == 0 {
			isDefault = true
		}
	} else if curFieldVal.Interface() == defaultValue.Interface() {
		isDefault = true
	}
	return
}

//GetFieldVal Convert reflect value to its corrosponding data type
func GetFieldVal(val reflect.Value) (castValue interface{}, err error) {
	switch val.Kind() {
	case reflect.String:
		castValue = val.String()
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		castValue = val.Int()
	case reflect.Float32, reflect.Float64:
		castValue = val.Float()
	case reflect.Map, reflect.Slice, reflect.Struct, reflect.Interface:
		castValue = val.Interface()
	default:
		err = errors.New("Invalid datatype: " + val.Kind().String())
	}
	return
}
