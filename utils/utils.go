package utils

import (
	"reflect"

	"github.com/bwmarrin/discordgo"
)

func Italics(s string) string {
	return "*" + s + "*"
}

func Bold(s string) string {
	return "**" + s + "**"
}

func Under(s string) string {
	return "__" + s + "__"
}

func Spoil(s string) string {
	return "||" + s + "||"
}

// Strlen Recursively searches for strings and counts up the total length
func Strlen(e *interface{}) int {
	count := 0
	v := reflect.ValueOf(e)
	for i := range v.NumField() {
		// iterate over fields
		switch f := v.Field(i); f.Kind() {
		case reflect.String:
			// count string field lengths
			count += len(f.String())
		case reflect.Ptr:
			// unroll pointers
			el := f.Elem()
			if !el.IsValid() {
				return 0
			}
			count += CheckEmbed(el)
		case reflect.Struct:
			// recurse over struct fields
			count += CheckEmbed(f.Interface())
		default:
			continue
		}
	}
	return count
}
