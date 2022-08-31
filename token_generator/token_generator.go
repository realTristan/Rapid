package token_generator

// Import Packages
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	Global "rapid_name_claimer/global"
	Queue "rapid_name_claimer/queue"

	"github.com/gookit/color"
	"github.com/valyala/fasthttp"
)

// Define Global Variables
var (
	// Counter Variables
	errorCount, generatedCount int = 0, 0

	// Previous Accounts Queue
	TokenAccountQueue *Queue.ItemQueue = Global.AddToQueue(Queue.Create(), "data/tokens/token_accounts.txt")
)

// The LiveCounter() function is used to display all of
// the stats for the token generator. This includes
// amount of tokens generated, errors, etc.
func LiveCounter(tokenCount int) {
	color.Printf(
		"\033[H\033[2J\033[0;0H%s\n\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Tokens \033[1;97m[\033[1;33m%d\033[1;97m]\033[1;34m   \033[1;37m\n ┃ \033[1;34m\033[1;37m\033[1;34m Proxies \033[1;97m[\u001b[30m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m Generated \033[1;97m[\033[1;32m%d\033[1;97m]\033[1;34m\033[1;37m\n ┃ \033[1;34m Errors \033[1;97m[\033[1;31m%d\033[1;97m]\033[1;34m\n\n \033[1;31m%s",
		Global.RapidLogoString, tokenCount, Global.ProxyQueue.Size(), generatedCount, errorCount, Global.CurrentError)
}

// The GetAuthTokenFromExistingAccountRequest() function is used to get the tokens
// from the existing accounts. (accounts found in token_accounts.txt)
func GetAuthTokenFromExistingAccountRequest(RequestClient *fasthttp.Client, account string) (*fasthttp.Response, error) {
	Global.SetProxy(RequestClient)

	// Define Variables
	var (
		// Marshal the request body being sent
		data, _ = json.Marshal(map[string]interface{}{"rememberMe": false})

		// Base64 encode the accounts email:password
		auth string = base64.StdEncoding.EncodeToString([]byte(account))

		// Create a new request opbject
		req *fasthttp.Request = Global.SetRequest("POST")
	)
	defer fasthttp.ReleaseRequest(req)

	// Set the request url, authorization header and body
	req.SetRequestURI(fmt.Sprintf("%sv3/profiles/sessions", Global.GetCustomUrl()))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	req.SetBody(data)

	// Define Variables
	var (
		// Create a new response object
		resp *fasthttp.Response = Global.SetResponse(false)

		// Send the http request
		err error = RequestClient.DoTimeout(req, resp, time.Second*6)
	)
	return resp, err
}

// The UsePreviousAccounts() function is used to send requests to the
// ubisoft api and get the new session token for each of the accounts
// in the token_accounts.txt file.
func UsePreviousAccounts(RequestClient *fasthttp.Client, tokenCount int) {
	for i := 0; i < TokenAccountQueue.Size(); i++ {
		// Define variables
		var (
			// Get the acocunt
			account string = fmt.Sprint(*TokenAccountQueue.Get())

			// Create a json response map variable
			respJson map[string]interface{}

			// Send the request to the ubi api
			resp, err = GetAuthTokenFromExistingAccountRequest(RequestClient, account)
		)
		defer fasthttp.ReleaseResponse(resp)

		// If no errors occured and the response
		// status code is 200 (success)
		if err == nil && resp.StatusCode() == 200 {

			// Unmarhsal the response body to the respJson
			// variable created above
			json.Unmarshal(resp.Body(), &respJson)

			// Get the accounts session ticket
			if respJson["ticket"] != nil {

				// Establish the authentication token using the ticket
				var authToken string = fmt.Sprintf("Ubi_v1 t=%s", respJson["ticket"])

				// Write the token to the tokens.txt file
				Global.WriteToFile("data/tokens/tokens.txt", &authToken)
			}
			generatedCount++
		} else {

			// Set the current error and increase the error count
			Global.CurrentError = fmt.Sprintf(" >> Token Error: %d: %v: %s", resp.StatusCode(), err, string(resp.Body()))
			errorCount++
		}

		// Show the token generator status and info
		LiveCounter(tokenCount)
	}
}

// The Start() function is used to start all of the functions
// that are going to be used in generating the tokens
//
// If the user types a "y" in the "Use Existing Accounts?" input,
// the UsePreviousAccounts function is called in a goroutine for
// iterating over the previously created token accounts, getting
// their authentication tokens.
func Start(tokenCount int) {
	// Define Variables
	var (
		// Client for sending http requests
		RequestClient *fasthttp.Client = Global.SetClient((&fasthttp.TCPDialer{Concurrency: 4096}).Dial)

		// Waitgroup for goroutines
		waitGroup sync.WaitGroup = sync.WaitGroup{}

		// Token Counter
		tokenCountTotal int = tokenCount
	)
	waitGroup.Add(1)

	// Option to use existing accounts for tokens
	var tokenGenUseExistingAccounts string
	color.Printf("\033[H\033[2J%s\n\n\033[97m ┃ Use Existing Accounts? \033[1;34m(y/n)\033[1;97m ", Global.RapidLogoString)
	fmt.Scan(&tokenGenUseExistingAccounts)

	// If the user said yes to using previous accounts
	if strings.Contains(tokenGenUseExistingAccounts, "y") && !TokenAccountQueue.IsEmpty() {

		// Update the token counter
		tokenCountTotal = tokenCount + TokenAccountQueue.Size()

		// Start the goroutine for getting previous account tokens
		go UsePreviousAccounts(RequestClient, tokenCountTotal)
	}

	// Iterate over the provided Token Count
	for i := 0; i < tokenCount; i++ {

		// Run everything below in a goroutine
		go func() {
			// Define Variables
			var (
				// The Randomized Name
				name string = fmt.Sprintf("rapd%s", Global.RandomString(11))

				// Create a new account
				resp, account, err = Global.CreateUplayAccount(RequestClient, name, "")
			)
			defer fasthttp.ReleaseResponse(resp)

			// If no errors occured and the response status code is 200 (succes)
			if err == nil && resp.StatusCode() == 200 {

				// Write the created account to the token_accounts.txt file
				go Global.WriteToFile("data/tokens/token_accounts.txt", &account)

				// Update generated counter
				generatedCount++
			} else {

				// Set the current error and increase the error count
				Global.CurrentError = fmt.Sprintf(" >> Token Error: %d: %v: %s", resp.StatusCode(), err, string(resp.Body()))
				errorCount++
			}

			// Show the token generator status and info
			LiveCounter(tokenCountTotal)
		}()
	}
	waitGroup.Wait()
}
