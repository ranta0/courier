package courier

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarValueReplace(t *testing.T) {
	env := map[string]string{
		"url": "http://localhost:8080/api/v1",
	}

	endpointWithVar := "{{ .url }}/users"
	if result, err := templateValueReplace(endpointWithVar, env); err == nil {
		assert.Equal(t, "http://localhost:8080/api/v1/users", result, "failed replacing template value")
	} else {
		assert.Fail(t, fmt.Sprintf("failed replacing template value, error: `%s`", err))
	}

	endpointWithVar = "/users"
	if result, err := templateValueReplace(endpointWithVar, env); err == nil {
		assert.Equal(t, "/users", result, "string with no placeholder failed")
	} else {
		assert.Fail(t, fmt.Sprintf("string with no placeholder failed, error: `%s`", err))
	}

	endpointWithVar = "{{ url }}/users"
	if result, err := templateValueReplace(endpointWithVar, env); err != nil {
		assert.Error(t, err)
	} else {
		assert.Fail(t, fmt.Sprintf("Not dotted variable did not throw an error, the resulting output is `%s`", result))
	}

	endpointWithVar = "{{ .token_value }}/users"
	if result, err := templateValueReplace(endpointWithVar, env); err == nil {
		assert.Equal(t, "/users", result, "string with undefined placeholder failed")
	} else {
		assert.Fail(t, fmt.Sprintf("string with undefined placeholder failed, error: `%s`", err))
	}
}

func TestGetValueForKeyJSON(t *testing.T) {
	output := `{ "data": [{ "id": 1 }, { "id": 2 }] }`
	var jsonData interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		assert.Fail(t, "error parsing json")
	}

	_, err := getValueForKey(jsonData, "data")
	if err != nil {
		assert.Fail(t, fmt.Sprintf("failed setting value from request, %s", err))
	}
}

func TestSeekAndSetJSON(t *testing.T) {
	env := map[string]interface{}{
		"result": 0,
		"id":     "",
		"item3":  nil,
	}
	usecase := APIUseCase{
		Name:     "get users standard",
		Method:   "GET",
		Endpoint: "{{ .url }}/users",
	}
	output := `{ "data": [{ "id": 1 }, { "id": 2 }], "result": "ok" }`

	var jsonData interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		assert.Fail(t, "error parsing json")
	}

	// string in var result, pulled from data[result]
	usecase.Vars = map[string]string{
		"result": "result",
	}
	if err := usecase.seekAndSetEnv(jsonData, env); err != nil {
		assert.Fail(t, fmt.Sprintf("failed setting value from request, %s", err))
	} else {
		assert.Equal(t, "ok", env["result"], "value set from request however it does not match")
	}

	// id value in var item, pulled from the first object in the data array
	// data[0] = first object
	// data[0].id = id of first object
	usecase.Vars = map[string]string{
		"id": "data[0].id",
	}
	if err := usecase.seekAndSetEnv(jsonData, env); err != nil {
		assert.Fail(t, fmt.Sprintf("failed setting value from request, %s", err))
	} else {
		assert.Equal(t, 1.0, env["id"], "value set from request however it does not match")
	}

	// TODO: do the object now...
}
