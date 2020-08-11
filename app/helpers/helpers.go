package helpers

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"time"
)

// InArray - check if val exist in array
func InArray(val interface{}, array interface{}) (exists bool) {
	exists = false

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				exists = true
				return
			}
		}
	case reflect.Map:
		m := reflect.ValueOf(array)
		for _, e := range m.MapKeys() {
			if reflect.DeepEqual(val, m.MapIndex(e).Interface()) {
				exists = true
				return
			}
		}
	}

	return
}

// KeyInMap - Check if ID is in map
func KeyInMap(a uint32, list map[uint32]int32) bool {
	for k := range list {
		if k == a {
			return true
		}
	}
	return false
}

// StringInSlice - Verifica che sia presente una string un uno slice di stringhe
func StringInSlice(v string, a []string) bool {
	for _, e := range a {
		if e == v {
			return true
		}
	}
	return false
}

// GetEndTime - Aggiunge un tempo di durata T.
func GetEndTime(hours, minutes, seconds int) (t time.Time) {
	t = time.Now().Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds) + time.Second)
	return
}

// NewContext - Recupero nuovo context per effettuare le chiamate
func NewContext(seconds time.Duration) context.Context {
	TTLRPC, err := strconv.Atoi(os.Getenv("TTL_RPC"))
	if err != nil {
		TTLRPC = 1
	}

	d := time.Now().Add(seconds * time.Second * time.Duration(TTLRPC))
	// nolint:govet // Escludo il check sul defer del cancel
	ctx, _ := context.WithDeadline(context.Background(), d)

	return ctx
}
