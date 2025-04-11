package structo

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/Lucifer07/Structo/errdefs"
)

// These flags define options for tag handling
const (
	// Denotes that a destination field must be copied to. If copying fails then a panic will ensue.
	tagMust uint8 = 1 << iota

	// Denotes that the program should not panic when the must flag is on and
	// value is not copied. The program will return an error instead.
	tagNoPanic

	// Ignore a destination field from being copied to.
	tagIgnore

	// Denotes that the value as been copied
	hasCopied
)

// Tag Flags
type flags struct {
	BitFlags  map[string]uint8
	SrcNames  tagNameMapping
	DestNames tagNameMapping
}

// Field Tag name mapping
type tagNameMapping struct {
	FieldNameToTag map[string]string
	TagToFieldName map[string]string
}

// parseTags Parses struct tags and returns uint8 bit flags.
func parseTags(tag string) (flg uint8, name string, err error) {
	for _, t := range strings.Split(tag, ",") {
		switch t {
		case "-":
			flg = tagIgnore
			return
		case "must":
			flg = flg | tagMust
		case "nopanic":
			flg = flg | tagNoPanic
		default:
			if unicode.IsUpper([]rune(t)[0]) {
				name = strings.TrimSpace(t)
			} else {
				err = errdefs.ErrFieldNameTagStartNotUpperCase
			}
		}
	}
	return
}

// getTagFlags Parses struct tags for bit flags, field name.
func getFlags(dest, src reflect.Value, toType, fromType reflect.Type) (flags, error) {
	flgs := flags{
		BitFlags: map[string]uint8{},
		SrcNames: tagNameMapping{
			FieldNameToTag: map[string]string{},
			TagToFieldName: map[string]string{},
		},
		DestNames: tagNameMapping{
			FieldNameToTag: map[string]string{},
			TagToFieldName: map[string]string{},
		},
	}
	var toTypeFields, fromTypeFields []reflect.StructField
	if dest.IsValid() {
		toTypeFields = deepFields(toType)
	}
	if src.IsValid() {
		fromTypeFields = deepFields(fromType)
	}

	// Get a list dest of tags
	for _, field := range toTypeFields {
		tags := field.Tag.Get("copier")
		if tags != "" {
			var name string
			var err error
			if flgs.BitFlags[field.Name], name, err = parseTags(tags); err != nil {
				return flags{}, err
			} else if name != "" {
				flgs.DestNames.FieldNameToTag[field.Name] = name
				flgs.DestNames.TagToFieldName[name] = field.Name
			}
		}
	}

	// Get a list source of tags
	for _, field := range fromTypeFields {
		tags := field.Tag.Get("copier")
		if tags != "" {
			var name string
			var err error
			if _, name, err = parseTags(tags); err != nil {
				return flags{}, err
			} else if name != "" {
				flgs.SrcNames.FieldNameToTag[field.Name] = name
				flgs.SrcNames.TagToFieldName[name] = field.Name
			}
		}
	}
	return flgs, nil
}
