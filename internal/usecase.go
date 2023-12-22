package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type UseCase struct {
	Name         string            `yaml:"name"`
	Method       string            `yaml:"method"`
	Endpoint     string            `yaml:"endpoint"`
	Body         string            `yaml:"body"`
	Headers      map[string]string `yaml:"headers"`

	WantStatus   int               `yaml:"wantStatus"`
	WantResponse string            `yaml:"wantResponse"`
	Vars         map[string]string `yaml:"vars"`
}

type APIUseCase struct {
	Name         string
	Method       string
	BaseURL      string
	Endpoint     string
	Body         io.Reader
	Headers      map[string]string
	Response     *http.Response
	Vars         map[string]string

	WantStatus   int
	WantResponse string
}

func NewAPIUseCase(baseUrl string, env interface{}, usecase *UseCase) (*APIUseCase, error) {
	headers := make(map[string]string)
	for key, header := range usecase.Headers {
		headers[key] = templateValueReplace(header, env)
	}

	endpoint := templateValueReplace(usecase.Endpoint, env)
	body := strings.NewReader(usecase.Body)

	return &APIUseCase{
		Name: usecase.Name,
		Method: usecase.Method,
		BaseURL: baseUrl,
		Endpoint: endpoint,
		Body: body,
		Headers: headers,
		Vars: usecase.Vars,
		WantStatus: usecase.WantStatus,
		WantResponse: usecase.WantResponse,
	}, nil
}

func (c *APIUseCase) Curl(env interface{}) (string, error) {
	err := c.request()
	if err != nil {
		return "", err
	}

	output, err := c.responseToString(env)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (c *APIUseCase) Test(env interface{}) error {
	err := c.request()
	if err != nil {
		return err
	}

	output, err := c.responseToString(env)
	if err != nil {
		return err
	}

	statusError := fmt.Sprintf("status mismatch: expected status %d, got %d", c.WantStatus, c.Response.StatusCode)
	if c.WantStatus != c.Response.StatusCode {
		return fmt.Errorf("%s", statusError)
	}

	lowerOutput := strings.ToLower(output)
	substring, _ := extractSubstringBetween(c.WantResponse, "{", "}")
	lowerSubstring := strings.ToLower(substring)

	responseError := fmt.Sprintf("response mismatch: expected '%s' is not contained in the output \n%s", substring, output)
	if !strings.Contains(lowerOutput, lowerSubstring) {
		return fmt.Errorf("%s", responseError)
	}

	return nil
}

func (c *APIUseCase) Prefix() string {
	var prefix = fmt.Sprintf("%s %s", c.Method, c.Endpoint)
	if c.Name != "" {
		prefix = fmt.Sprintf("[%s] %s %s", c.Name, c.Method, c.Endpoint)
	}
	return prefix
}

func (c *APIUseCase) request() (error) {
	request, err := http.NewRequest(c.Method, c.BaseURL + c.Endpoint, c.Body)
	if err != nil {
		return err
	}

	for k, v := range c.Headers {
		request.Header.Set(k, v)
	}

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	c.Response = res
	return nil
}

func (c *APIUseCase) responseToString(env interface{}) (string, error) {
	defer c.Response.Body.Close()

	body, err := io.ReadAll(c.Response.Body)
	if err != nil {
		return "", err
	}
	bodyStr := string(body)

	if len(c.Vars) == 0 {
		return bodyStr, nil
	}

	var output interface{}

	if err := json.Unmarshal([]byte(bodyStr), &output); err != nil {
		return bodyStr, err
	}

	err = c.seekAndSetEnv(output, env)
	if err != nil {
		return bodyStr, err
	}

	return bodyStr, nil
}

func (c *APIUseCase) seekAndSetEnv(output, env interface{}) error {
	for key, value := range c.Vars {
		varKeys := strings.Split(value, ",")

		var newEnvValue interface{}
		outputDepth := output
		for _, node := range varKeys {
			index, remaingSubstring := extractSubstringBetween(node, "[", "]")
			index = strings.ReplaceAll(index, " ", "")
			remaingSubstring = strings.ReplaceAll(remaingSubstring, " ", "")

			if index != "" {
				newIndex, err := strconv.Atoi(index)
				if err != nil {
					return err
				}

				outputDepthOne, err := getValueForKey(outputDepth, remaingSubstring)
				if err != nil {
					return fmt.Errorf("response does not contain %s", remaingSubstring)
				}
				v := reflect.ValueOf(outputDepthOne)

				if v.Kind() != reflect.Slice {
					return fmt.Errorf("%s picked in var is not an array, it is %s", outputDepthOne, v.Kind())
				}

				length := v.Len()
				if newIndex < 0 || newIndex >= length {
					return fmt.Errorf("index requested is out of bounds, length '%s' is %d", remaingSubstring, length)
				}

				newEnvValue = v.Index(newIndex).Interface()
				outputDepth = v.Index(newIndex).Interface()
			} else {
				keyValue, err := getValueForKey(outputDepth, node)
				if err != nil {
					return fmt.Errorf("response does not contain %s", node)
				}
				newEnvValue = keyValue
			}
		}

		envUpdated, err := setValueForKey(env, key, newEnvValue)
		if err != nil {
			return err
		}
		env = envUpdated
	}

	return nil
}

func templateValueReplace(str string, env interface{}) string {
	t, err := template.New("str").Parse(str)
	if err != nil {
		log.Fatalln(err)
	}
	var buff bytes.Buffer

	err = t.Execute(&buff, env)
	if err != nil {
		log.Fatalln(err)
	}

	return buff.String()
}

func extractSubstringBetween(input, startSubstring, endSubstring string) (string, string) {
	startIndex := strings.Index(input, startSubstring)
	endIndex := strings.Index(input, endSubstring)

	if startIndex == -1 || endIndex == -1 || startIndex >= endIndex {
		return "", ""
	}

	resultSubstring := input[startIndex+len(startSubstring):endIndex]
	remaingSubstring := input[0:startIndex]
	return resultSubstring, remaingSubstring
}

func getValueForKey(data interface{}, key string) (interface{}, error) {
	if m, ok := data.(map[string]interface{}); ok {
		if value, exists := m[key]; exists {
			return value, nil
		} else {
			return nil, fmt.Errorf("key %s not found", key)
		}
	}
	return nil, fmt.Errorf("value is not a map")
}

func setValueForKey(data interface{}, key string, value interface{}) (interface{}, error) {
	if m, ok := data.(map[string]interface{}); ok {
		if _, ok := m[key]; !ok {
			return nil, fmt.Errorf("var '%s' was not defined in the config", key)
		}
		m[key] = value
		return m, nil
	}
	return nil, fmt.Errorf("value is not a map")
}
