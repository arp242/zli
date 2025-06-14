package zli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (f *Flags) fromEnv(prefix string) error {
	if prefix != "" {
		prefix = strings.ToUpper(strings.TrimRight(prefix, "_")) + "_"
	}

	var unknown []string
	for _, e := range os.Environ() {
		k, v, _ := strings.Cut(e, "=")
		k = strings.ReplaceAll(strings.ToUpper(k), "-", "_")
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		key := k /// For errors.
		k = k[len(prefix):]
		if len(k) < 2 {
			continue
		}

		flag, ok := f.match(k, true)
		if !ok {
			unknown = append(unknown, key)
			continue
		}
		err := setFromEnv(flag, k, v)
		if err != nil {
			return fmt.Errorf("environment variable %q: %w", key, err)
		}
	}
	if len(unknown) > 0 {
		return ErrUnknownEnv{prefix, unknown}
	}
	return nil
}

func setFromEnv(flag flagValue, k, val string) error {
	switch v := flag.value.(type) {
	case flagBool:
		x, ok := parseEnvBool(val)
		*v.s, *v.e, *v.v = true, true, x
		if !ok {
			return fmt.Errorf("invalid value %q for boolean %q", val, k)
		}
	case flagString:
		*v.s, *v.e, *v.v = true, true, val
	case flagInt:
		x, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			if nErr := errors.Unwrap(err); nErr != nil {
				err = nErr
			}
			return ErrFlagInvalid{k, err, "number"}
		}
		*v.s, *v.e, *v.v = true, true, int(x)
	case flagInt64:
		x, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			if nErr := errors.Unwrap(err); nErr != nil {
				err = nErr
			}
			return ErrFlagInvalid{k, err, "number"}
		}
		*v.s, *v.e, *v.v = true, true, x
	case flagFloat64:
		x, err := strconv.ParseFloat(val, 64)
		if err != nil {
			if nErr := errors.Unwrap(err); nErr != nil {
				err = nErr
			}
			return ErrFlagInvalid{k, err, "number"}
		}
		*v.s, *v.e, *v.v = true, true, x
	case flagIntCounter:
		b, ok := parseEnvBool(val)
		switch {
		case ok && b:
			*v.s, *v.e, *v.v = true, true, 1
		case ok && !b:
			*v.s, *v.e, *v.v = true, true, 0
		default:
			n, err := strconv.ParseInt(val, 0, 64)
			if err != nil {
				if nErr := errors.Unwrap(err); nErr != nil {
					err = nErr
				}
				return ErrFlagInvalid{k, err, "number"}
			}
			*v.s, *v.e, *v.v = true, true, int(n)
		}
	case flagStringList:
		*v.s, *v.e, *v.v = true, true, strings.Split(val, ",")
	case flagIntList:
		*v.s, *v.e, *v.v = true, true, nil
		for _, n := range strings.Split(val, ",") {
			x, err := strconv.ParseInt(n, 0, 64)
			if err != nil {
				if nErr := errors.Unwrap(err); nErr != nil {
					err = nErr
				}
				return ErrFlagInvalid{k, err, "number"}
			}
			*v.v = append(*v.v, int(x))
		}
	}
	return nil
}

func parseEnvBool(val string) (bool, bool) {
	switch strings.ToLower(val) {
	case "1", "true", "t", "":
		return true, true
	case "0", "false", "f":
		return false, true
	default:
		return false, false
	}
}
