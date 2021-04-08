package provider

import (
	"io/ioutil"
	"log"
	"net/http"
)

func handleClientError(op string, errorString string, httpResponse *http.Response) {
	bodyBytes, err := ioutil.ReadAll(httpResponse.Body)
	bodyString := ""
	if err == nil {
		bodyString = string(bodyBytes)
		log.Printf("[ERROR] Unable to perform operation %s. \n Reason: %s \n Details: %s", op, errorString, bodyString)
	} else {
		log.Printf("[ERROR] Unable to perform operation %s. \n Reason: %s", op, errorString)
	}
}
