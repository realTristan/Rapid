package global

// Import Packages
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	Queue "rapid_name_claimer/queue"
	"time"

	"github.com/valyala/fasthttp"
)

// Token Account Queue Variable
var TokenAccountQueue *Queue.ItemQueue = AddToQueue(Queue.Create(), "data/tokens/token_accounts.txt")

// The GetAuthTokenFromExistingAccount() function is used to get the tokens
// from the existing accounts. (accounts found in token_accounts.txt)
func GetAuthTokenFromExistingAccount(RequestClient *fasthttp.Client) (*fasthttp.Response, string, error) {
	SetProxy(RequestClient)

	// Define Variables
	var (
		// Marshal the request body being sent
		data, _ = json.Marshal(map[string]interface{}{"rememberMe": false})

		// Base64 encode the accounts email:password
		auth string = base64.StdEncoding.EncodeToString([]byte((*TokenAccountQueue.Get()).(string)))

		// Create a new request opbject
		req *fasthttp.Request = SetRequest("POST")
	)
	defer fasthttp.ReleaseRequest(req)

	// Set the request url, authorization header and body
	req.SetRequestURI(fmt.Sprintf("%sv3/profiles/sessions", GetCustomUrl()))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	req.SetBody(data)

	// Define Variables
	var (
		// Create a new response object
		resp *fasthttp.Response = SetResponse(false)

		// Send the http request
		err error = RequestClient.DoTimeout(req, resp, time.Second*6)

		// Create a json response map variable
		respJson map[string]interface{}

		// The authentication token grabbed from the account session
		authToken string
	)

	// If no errors occured and the response
	// status code is 200 (success)
	if err == nil && resp.StatusCode() == 200 {
		// Unmarhsal the response body to the respJson
		// variable created above
		json.Unmarshal(resp.Body(), &respJson)

		// Get the accounts session ticket
		if respJson["ticket"] != nil {
			// Establish the authentication token using the ticket
			authToken = fmt.Sprintf("Ubi_v1 t=%s", respJson["ticket"])
		}
	} else {
		// Set the current error and increase the error count
		CurrentError = fmt.Sprintf(" >> Token Error: %d: %v: %s", resp.StatusCode(), err, string(resp.Body()))
	}
	return resp, authToken, err
}
