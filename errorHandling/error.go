package errorHandling

import (
	"log"
	"os"
	"reflect"
	"runtime"
)

type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return e.Message
}

func NewAPIError(status int, i interface{}, message string) error {
	log.Printf("\nGIN_MODE: %s, Code: %d, %s: %s", os.Getenv("GIN_MODE"), status, GetFunctionName(i), message)
	return APIError{Status: status, Message: message}
}

func GetFunctionName(i interface{}) string {
	funcName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	return funcName
}

func LogToFile(file string, statement string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err = f.WriteString(statement); err != nil {
		log.Fatal(err)
	}
}
