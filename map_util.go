package main

import (
	"maps"
	"reflect"
	"slices"
)

// Given a map, returns (first key, true) if non-empty, and
// (nil value, false) if empty
func FirstKey[K string | int32 | int64 | float32 | float64, V any](m map[K]V) (K, bool) {
	var nullKey K
	if len(m) == 0 {
		return nullKey, false
	}

	keySlice := KeySlice(m)
	slices.Sort(keySlice)

	return keySlice[0], true
}

// Returns a slice of the keys of a map
func KeySlice[K comparable, V any](m map[K]V) []K {
	keySlice := make([]K, 0, len(m))
	for k := range m {
		keySlice = append(keySlice, k)
	}
	return keySlice
}

// Returns a deep copy of `m`, a SettingsMap struct
func DeepCopySettings(m SettingsMap) SettingsMap {
	var copiedMap SettingsMap
	copiedMap.Script = maps.Clone(m.Script)
	copiedMap.General = maps.Clone(m.General)
	copiedMap.Slurm = DeepCopySlurm(m.Slurm)

	return copiedMap
}

// Returns a deep copy of the `Slurm` property of a `SettingsMap`
// The arbitrary structure of the `Slurm` map will be used in
// future versions of glurmo
func DeepCopySlurm(m map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(m))
	for k, v := range m {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			copy[k] = DeepCopySlurm(v.(map[string]interface{}))
		} else {
			copy[k] = v
		}
	}
	return copy
}

// Converts a `map[string]interface{}` to a `map[string]string`.
// Fails if `m` contains non-string values
func InterfaceToStringMap(m map[string]interface{}) (map[string]string, error) {
	stringMap := make(map[string]string)
	for k, v := range m {
		if reflect.TypeOf(v).Kind() != reflect.String {
			return nil, errorString{"map contains non-string elements"}
		}
		stringMap[k] = v.(string)
	}

	return stringMap, nil
}
