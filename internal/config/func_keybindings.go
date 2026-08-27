package config

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
)

func ShortHelp(keymap any) []key.Binding {
	return bindingsWithHelp(keymap, "short")
}

func pageFullHelp(keymap any) []key.Binding {
	return bindingsWithHelp(keymap, "short", "full")
}

func bindingsWithHelp(keymap any, levels ...string) []key.Binding {
	v := reflect.ValueOf(keymap)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	t := v.Type()

	var out []key.Binding
	for i := range v.NumField() {
		field := t.Field(i)
		if field.Type != bindingType {
			continue
		}
		if slices.Contains(levels, field.Tag.Get("help")) {
			out = append(out, v.Field(i).Interface().(key.Binding))
		}
	}
	return out
}

type HelpKeyMap struct {
	pages []any
}

func Help(pages ...any) HelpKeyMap {
	return HelpKeyMap{pages: pages}
}

func (h HelpKeyMap) ShortHelp() []key.Binding {
	var out []key.Binding
	for _, p := range h.pages {
		out = append(out, ShortHelp(p)...)
	}
	return out
}

func (h HelpKeyMap) FullHelp() [][]key.Binding {
	cols := make([][]key.Binding, 0, len(h.pages))
	for _, p := range h.pages {
		if col := pageFullHelp(p); len(col) > 0 {
			cols = append(cols, col)
		}
	}
	return cols
}

var bindingType = reflect.TypeFor[key.Binding]()

var validHelpLevels = map[string]bool{"none": true, "short": true, "full": true}

func stringToKeyBindingHook(from reflect.Type, to reflect.Type, data any) (any, error) {
	if to != bindingType || from.Kind() != reflect.String {
		return data, nil
	}

	var keys []string
	for part := range strings.SplitSeq(data.(string), ",") {
		if k := strings.TrimSpace(part); k != "" {
			keys = append(keys, k)
		}
	}
	return key.NewBinding(key.WithKeys(keys...)), nil
}

func fillHelp(v reflect.Value) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		fv := v.Field(i)

		if field.Type == bindingType {
			b := fv.Addr().Interface().(*key.Binding)
			b.SetHelp(strings.Join(b.Keys(), ","), field.Tag.Get("desc"))

			if level := field.Tag.Get("help"); !validHelpLevels[level] {
				panic(fmt.Sprintf(
					"config: %s.%s has invalid `help` tag %q, want \"none\", \"short\" or \"full\"",
					t.Name(), field.Name, level,
				))
			}
			continue
		}
		if field.Type.Kind() == reflect.Struct {
			fillHelp(fv)
		}
	}
}
