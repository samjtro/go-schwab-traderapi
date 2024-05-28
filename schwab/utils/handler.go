package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Handler is the general purpose request function for the td-ameritrade api, all functions will be routed through this handler function, which does all of the API calling work
// It performs a GET request after adding the apikey found in the .APIKEY file in the same directory as the program calling the function,
// then returns the body of the GET request's return.
// It takes one parameter:
// req = a request of type *http.Request
func Handler(req *http.Request) (string, error) {
	var (
		m      sync.Mutex
		tokens TOKEN
	)

	m.Lock()

	if _, err := os.Stat("~/.foo/bar.json"); errors.Is(err, os.ErrNotExist) {
		tokens = oAuthInit()
	} else {
		tokens = readDB()
	}

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	if !time.Now().Before(tokens.BearerExpiration) {
		req.Header = http.Header{
			"Authorization": {fmt.Sprintf("Bearer %s", tokens.Bearer)},
		}
	} else {
		newBearerToken := oAuthRefresh()
		req.Header = http.Header{
			"Authorization": {fmt.Sprintf("Bearer %s", newBearerToken)},
		}
	}

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	errorCode := resp.StatusCode
	bodyBytes, err := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	fmt.Println(body)

	if err != nil {
		return "", err
	}

	if errorCode < 200 || errorCode > 300 {
		log.Fatalf("Error %d - %s", errorCode, body)
	}

	m.Unlock()
	return body, nil
}
