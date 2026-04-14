package helper

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func DecodeRequestBody(request *http.Request, destination interface{}) error {
	jsonDecoder := json.NewDecoder(request.Body)
	jsonDecoder.DisallowUnknownFields()

	if err := jsonDecoder.Decode(destination); err != nil {
		return err
	}

	var trailingPayload map[string]interface{}
	if err := jsonDecoder.Decode(&trailingPayload); !errors.Is(err, io.EOF) {
		return errors.New("request body must only contain a single json object")
	}

	return nil
}
