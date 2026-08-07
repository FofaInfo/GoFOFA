package gofofa

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type apiErrorEnvelope struct {
	Error  bool   `json:"error"`
	Errmsg string `json:"errmsg"`
}

func decodeAPIErrorEnvelope(body []byte) (apiErrorEnvelope, error) {
	var response apiErrorEnvelope
	err := json.Unmarshal(body, &response)
	return response, err
}

func apiResponseError(failed bool, errmsg, fallback string) error {
	if !failed {
		return nil
	}
	if errmsg != "" {
		return errors.New(errmsg)
	}
	return errors.New(fallback)
}

func apiErrorCode(errmsg string) (int, bool) {
	errmsg = strings.TrimSpace(errmsg)
	if len(errmsg) < 3 || errmsg[0] != '[' {
		return 0, false
	}
	end := strings.IndexByte(errmsg, ']')
	if end < 2 {
		return 0, false
	}
	code, err := strconv.Atoi(errmsg[1:end])
	if err != nil {
		return 0, false
	}
	return code, true
}
