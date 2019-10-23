package util

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

func ConvertToBytes(text string) (float64, error) {
	text = strings.ToLower(text)

	if strings.HasSuffix(text, "tib") {
		value := strings.ReplaceAll(text, "tib", "")
		value = strings.TrimSpace(value)
		ret, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "raw=%s", value)
		}
		return ret * 1000 * 1000 * 1000 * 1000, nil
	} else if strings.HasSuffix(text, "gib") {
		value := strings.ReplaceAll(text, "gib", "")
		value = strings.TrimSpace(value)
		ret, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "raw=%s", value)
		}
		return ret * 1000 * 1000 * 1000, nil
	} else if strings.HasSuffix(text, "mib") {
		value := strings.ReplaceAll(text, "mib", "")
		value = strings.TrimSpace(value)
		ret, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "raw=%s", value)
		}
		return ret * 1000 * 1000, nil
	} else if strings.HasSuffix(text, "kib") {
		value := strings.ReplaceAll(text, "kib", "")
		value = strings.TrimSpace(value)
		ret, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "raw=%s", value)
		}
		return ret * 1000, nil
	} else if strings.HasSuffix(text, "tb") {
		value := strings.ReplaceAll(text, "tb", "")
		value = strings.TrimSpace(value)
		ret, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "raw=%s", value)
		}
		return ret * 1024 * 1024 * 1024 * 1024, nil

	} else if strings.HasSuffix(text, "gb") {
		value := strings.ReplaceAll(text, "gib", "")
		value = strings.TrimSpace(value)
		ret, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "raw=%s", text)
		}
		return ret * 1024 * 1024 * 1024, nil
	} else if strings.HasSuffix(text, "mb") {
		value := strings.ReplaceAll(text, "mib", "")
		value = strings.TrimSpace(value)
		ret, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "raw=%s", text)
		}
		return ret * 1024 * 1024, nil
	} else if strings.HasSuffix(text, "kb") {
		value := strings.ReplaceAll(text, "kib", "")
		value = strings.TrimSpace(value)
		ret, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "raw=%s", text)
		}
		return ret * 1024, nil
	} else {
		return 0, errors.New("invalid format: " + text)
	}
}
