package utils

import (
	"reflect"
)

// Italics Encloses string in italics tags
func Italics(s string) string {
	return "*" + s + "*"
}

// Bold Encloses string in bold tags
func Bold(s string) string {
	return "**" + s + "**"
}

// Under Encloses string in underline tags
func Under(s string) string {
	return "__" + s + "__"
}

// Spoil Encloses string in spoiler tags
func Spoil(s string) string {
	return "||" + s + "||"
}

// Strlen Recursively searches for strings and counts up the total length
func Strlen(e interface{}) int {
	count := 0
	// type assert e
	ev, ok := e.(reflect.Value)
	if !ok {
		ev = reflect.ValueOf(e)
	}

	switch ev.Kind() {
	case reflect.String:
		// count string lengths
		count += len(ev.String())

	case reflect.Ptr:
		// unroll pointers
		elv := ev.Elem()
		if !elv.IsValid() {
			return 0
		}
		count += Strlen(elv)

	case reflect.Struct:
		// iterate over fields
		for i := 0; i < ev.NumField(); i++ {
			switch f := ev.Field(i); f.Kind() {
			case reflect.String:
				// count string field lengths
				count += len(f.String())

			case reflect.Ptr:
				// unroll pointers
				el := f.Elem()
				if !el.IsValid() {
					return 0
				}
				count += Strlen(el)

			case reflect.Struct:
				// recurse over struct fields
				count += Strlen(f.Interface())

			default:
				// do nothing
			}
		}
	default:
		// do nothing
	}
	return count
}
