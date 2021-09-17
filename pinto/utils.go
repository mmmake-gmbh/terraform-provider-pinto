package pinto

import (
	"io/ioutil"
	"log"
	"net/http"
)

const (
	// using name "pinto_provider" because "provider" is a reserved word for resources and data sources
	schemaProvider = "pinto_provider"
	// using name "pinto_environment" to keep the same naming schema as schemaProvider
	schemaEnvironment = "pinto_environment"
)

func handleClientError(op string, errorString string, httpResponse *http.Response) string {
	bodyBytes, err := ioutil.ReadAll(httpResponse.Body)
	bodyString := ""
	if err == nil {
		bodyString = string(bodyBytes)
		log.Printf("[ERROR] Unable to perform operation %s. \n Reason: %s \n Details: %s", op, errorString, bodyString)
		return errorString + ": " + bodyString
	} else {
		log.Printf("[ERROR] Unable to perform operation %s. \n Reason: %s", op, errorString)
		return errorString
	}
}

// TODO: Clarify missing struct in client
type AccessOptions struct {
	Provider      string `json:"provider"`
	Environment   string `json:"environment"`
	CredentialsId string `json:"credentials_id"`
}
type XApiOptions struct {
	AccessOptions AccessOptions `json:"access_options"`
}
