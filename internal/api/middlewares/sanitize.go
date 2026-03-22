package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"restapi/pkg/utils"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func XSSProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		sanitizedPath, err := clean(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		params := r.URL.Query()
		sanitizedQuery := make(map[string][]string)
		for key, values := range params {
			sanitizedKey, err := clean(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var sanitizedValues []string
			for _, value := range values {
				sanitizedValue, err := clean(value)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				sanitizedValues = append(sanitizedValues, sanitizedValue.(string))
			}
			sanitizedQuery[sanitizedKey.(string)] = sanitizedValues
		}

		r.URL.Path = sanitizedPath.(string)
		r.URL.RawQuery = url.Values(sanitizedQuery).Encode() //""

		if r.Header.Get("Content-Type") == "application/json" {
			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, utils.HandleError(err, "Error reading request body").Error(), http.StatusBadRequest)
					return
				}
				bodyString := strings.TrimSpace(string(bodyBytes))
				r.Body = io.NopCloser(bytes.NewReader([]byte(bodyString)))
				if bodyString != "" {
					var inputData interface{}
					err = json.NewDecoder(bytes.NewReader([]byte(bodyString))).Decode(&inputData)
					if err != nil {
						http.Error(w, utils.HandleError(err, "Error parsing JSON request body").Error(), http.StatusBadRequest)
						return
					}
					sanitizedBody, err := clean(bodyString)
					if err != nil {
						http.Error(w, utils.HandleError(err, "Error sanitizing request body").Error(), http.StatusBadRequest)
						return
					}
					sanitizedBodyBytes, err := json.Marshal(sanitizedBody)
					if err != nil {
						http.Error(w, utils.HandleError(err, "Error marshaling sanitized request body").Error(), http.StatusBadRequest)
						return
					}
					r.Body = io.NopCloser(bytes.NewReader(sanitizedBodyBytes))
				} else {
					http.Error(w, utils.HandleError(fmt.Errorf("empty request body"), "Empty request body").Error(), http.StatusBadRequest)
					return
				}
			} else {
				http.Error(w, utils.HandleError(fmt.Errorf("empty request body"), "Empty request body").Error(), http.StatusBadRequest)
				return
			}
		} else {
			log.Printf("Non-JSON content type: %s", r.Header.Get("Content-Type"))
			http.Error(w, utils.HandleError(fmt.Errorf("unsupported content type"), "Unsupported Content-Type").Error(), http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Clean sanitizes input data by trimming spaces and escaping HTML characters
func clean(data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case string:
		return sanitizeString(v), nil
	case map[string]interface{}:
		for key, value := range v {
			cleanedValue, err := sanitizeValue(reflect.ValueOf(value))
			if err != nil {
				return nil, err
			}
			v[key] = cleanedValue.Interface()
		}
		return v, nil
	case []interface{}:
		for i, item := range v {
			cleanedValue, err := sanitizeValue(reflect.ValueOf(item))
			if err != nil {
				return nil, err
			}
			v[i] = cleanedValue.Interface()
		}
		return v, nil
	default:
		return nil, utils.HandleError(fmt.Errorf("unsupported type: %T", data), fmt.Sprintf("unsupported type: %T", data))
	}
}

func sanitizeValue(value reflect.Value) (reflect.Value, error) {
	switch value.Kind() {
	case reflect.String:
		cleanedValue := sanitizeString(value.String())
		return reflect.ValueOf(cleanedValue).Convert(value.Type()), nil
	case reflect.Struct:
		cleanedStruct := reflect.New(value.Type()).Elem()
		for i := 0; i < value.NumField(); i++ {
			fieldValue, err := sanitizeValue(value.Field(i))
			if err != nil {
				return reflect.Value{}, err
			}
			cleanedStruct.Field(i).Set(fieldValue)
		}
		return cleanedStruct, nil
	case reflect.Slice:
		cleanedSlice := reflect.MakeSlice(value.Type(), value.Len(), value.Cap())
		for i := 0; i < value.Len(); i++ {
			cleanedItem, err := sanitizeValue(value.Index(i))
			if err != nil {
				return reflect.Value{}, err
			}
			cleanedSlice.Index(i).Set(cleanedItem)
		}
		return cleanedSlice, nil
	default:
		return value, nil
	}
}

func sanitizeString(input string) string {
	policy := bluemonday.UGCPolicy()
	policy.AllowStandardURLs()
	return policy.Sanitize(input)
}
