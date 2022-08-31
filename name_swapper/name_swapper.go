package name_swapper

// Import Packages
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	Global "rapid_name_claimer/global"
	NameChecker "rapid_name_claimer/name_checker"
	"strings"
	"sync"
	"time"

	"github.com/gookit/color"
	"github.com/valyala/fasthttp"
)

// Global Error Count Variable
var errorCount int = 0

// The ShowInfo() function is used to show the name swapper
// status and info. This includes the speed of the swapper,
// the provided account logins, errors, etc.
func ShowInfo(accountWithName string, accountToPutNameOn string, speed float64) {
	color.Printf(
		"\033[H\033[2J\033[0;0H%s\n\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Account With Name \033[1;97m[\033[1;33m%s\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Account To Put Name On \033[1;97m[\033[1;35m%s\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Proxies \033[1;97m[\033[1;30m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Speed \033[1;97m[\033[1;36m%vs\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m Errors \033[1;97m[\033[1;31m%d\033[1;97m]\033[1;34m\n\n \033[1;31m%s",
		Global.RapidLogoString, accountWithName, accountToPutNameOn, Global.ProxyQueue.Size(), speed, errorCount, Global.CurrentError)
}

// The Error() function is used to handle any response
// errors from the name swapper.
func Error(resp *fasthttp.Response, err error, accountWithName string, accountToPutNameOn string, errNum string) {
	defer fasthttp.ReleaseResponse(resp)

	// Variable to prevent the program from automatically closing
	var preventClose string

	// Handle response
	if err != nil || resp.StatusCode() != 200 {
		errorCount++

		// Set the global error
		Global.CurrentError = fmt.Sprintf(" >> Error %s: %d: %v: %s", errNum, resp.StatusCode(), err, string(resp.Body()))
		ShowInfo(accountWithName, accountToPutNameOn, 0)

		// Pause program before exiting
		fmt.Scan(&preventClose)
		os.Exit(1)
	}
}

// The CreateNameChangeRequestObject() is used to create the fasthttp
// request object for the name swapper. Having this as a function
// dramastically increases the speed of the name swapper.
func CreateNameChangeRequestObject(session map[string]interface{}, name string) *fasthttp.Request {
	// Define Variables
	var (
		// Create Request Object
		req *fasthttp.Request = Global.SetRequest("PUT")
		// Marshal the body being sent in the request
		data, _ = json.Marshal(map[string]interface{}{"nameOnPlatform": name})
	)

	// Set the request url, headers and body
	req.SetRequestURI(Global.GetCustomUrl() + "v3/profiles/" + session["profileId"].(string))
	req.Header.Set("Authorization", "Ubi_v1 t="+session["ticket"].(string))
	req.Header.Set("Ubi-SessionId", session["sessionId"].(string))
	req.SetBody(data)

	// Return the new request object
	return req
}

// The CheckNameChangeStatus() function is used to check whether
// the user has been ratelimited from the ubisoft api. This
// prevents losing swaps.
func CheckNameChangeStatus(RequestClient *fasthttp.Client, session map[string]interface{}) (*fasthttp.Response, error) {
	// Define Variables
	var (
		// Create the request object
		req *fasthttp.Request = Global.SetRequest("POST")
		// Marshal the body being sent in the request
		data, _ = json.Marshal(map[string]interface{}{"nameOnPlatform": session["nameOnPlatform"].(string)})
	)
	defer fasthttp.ReleaseRequest(req)

	// Set the request url, headers and body
	req.SetRequestURI(fmt.Sprintf("%sv3/profiles/%s/validateUpdate", Global.GetCustomUrl(), session["profileId"].(string)))
	req.Header.Set("Authorization", "Ubi_v1 t="+session["ticket"].(string))
	req.Header.Set("Ubi-SessionId", session["sessionId"].(string))
	req.SetBody(data)

	// Define Variables
	var (
		// Create the response object
		resp *fasthttp.Response = Global.SetResponse(false)
		// Send the http request
		err error = RequestClient.DoTimeout(req, resp, time.Second*6)
	)
	return resp, err
}

// The GetSession() function is used to send an http request
// to the ubisoft session login endpoint. The function then
// returns the response json which is used for getting
// the session data.
func GetSession(RequestClient *fasthttp.Client, account string) (*fasthttp.Response, map[string]interface{}, error) {
	// Define Variables
	var (
		// Base64 encode the account email:password
		auth string = base64.StdEncoding.EncodeToString([]byte(account))
		// Create a request object
		req *fasthttp.Request = Global.SetRequest("POST")
		// Marshal the body being sent in the request
		data, _ = json.Marshal(map[string]interface{}{"rememberMe": false})
	)
	defer fasthttp.ReleaseRequest(req)

	// Set the request url, headers and body being sent
	req.SetRequestURI(Global.GetCustomUrl() + "v3/profiles/sessions")
	req.Header.Set("Authorization", "Basic "+auth)
	req.SetBody(data)

	// Define Variables
	var (
		// Create a response object
		resp *fasthttp.Response = Global.SetResponse(false)
		// Send the http request
		err error = RequestClient.DoTimeout(req, resp, time.Second*6)
	)

	// Unmarshal the response json to a readable map
	var respJson map[string]interface{}
	if resp.StatusCode() == 200 && err == nil {
		json.Unmarshal(resp.Body(), &respJson)
	}
	return resp, respJson, err
}

// The NameRecover() function is used to claim the wanted name
// if the swap had failed. Although this function is rarely called
// because of all the security measures called before it.
func NameRecover(RequestClient *fasthttp.Client, name string, accountWithName string, accountToPutNameOn string, speed float64) {
	Global.SetProxy(RequestClient)

	// Define Variables
	var (
		// Create a new account with the name needing recovery
		resp, newAccount, err = Global.CreateUplayAccount(RequestClient, name, Global.JsonData["custom_claim_email"].(string))
		// Create the claim string for writing to claimed.txt
		claimString string = fmt.Sprintf("Name: %s ┃ Login: %s", name, newAccount)
	)
	defer fasthttp.ReleaseResponse(resp)

	// If the name claim status code is 200 and the error is nil
	if resp.StatusCode() == 200 && err == nil {
		// Set the current error to the name's new email and password
		Global.CurrentError = fmt.Sprintf(" >> Swap Failed! \033[1;97m[\033[1;31m%s\033[1;97m] => \033[1;97m[\033[1;31m%s\033[1;97m]", name, newAccount)
		// And write it to the claimed.txt file
		go Global.WriteToFile("data/name_checker/claimed.txt", &claimString)
	} else
	// Increase the errorcount and set the current error
	{
		errorCount++
		Global.CurrentError = fmt.Sprintf(" >> Name Recover Error: %d: %v: %s", resp.StatusCode(), err, string(resp.Body()))
	}
	// Increase error count and show the swap status
	errorCount++
	ShowInfo(accountWithName, accountToPutNameOn, speed)
}

// The GetAccountReplacementName() function is used for getting the
// account with the wanted name's name.
// The function gets the users input for a replacement name. To ensure
// the security of the swap, the name is checked for availability
// and whether the name is valid or not.
func GetAccountReplacementName(RequestClient *fasthttp.Client, AccountWithNameSession map[string]interface{}) string {
	// Get the new replacement name
	var name string
	color.Printf("\n\033[97m ┃ Replacement Name: \033[1;34m(Type 'r' for Random) \033[1;97m ")
	fmt.Scan(&name)

	// Set custom replacement name to random if invalid
	if name == "r" || name == "'r'" {
		// Generate a random name
		name = "rapd" + Global.RandomString(11)
	} else {

		// If the name isn't valid (eg: starts with a number, etc.)
		if !Global.IsValidUplayName(name) {
			color.Printf("\033[97m ┃ \033[1;31mInvalid Name\n")

			// Recall the function (loop)
			GetAccountReplacementName(RequestClient, AccountWithNameSession)
		} else

		// Send an http request to the ubisoft api
		// to check whether the provided name is available
		{
			// Define Variables
			var (
				// The url to send the http request to
				url string = Global.GetCustomUrl() + "v3/profiles?platformType=uplay&nameOnPlatform=" + name
				// The Authorization token for sending the request
				token string = "Ubi_v1 t=" + AccountWithNameSession["ticket"].(string)
				// Send the http request
				ValidNameResp, ValidNameErr = NameChecker.CheckNamesRequest(RequestClient, url, token)
				// Store the response body
				body string = string(ValidNameResp.Body())
			)

			// If the name is not available or an error has occured
			if ValidNameResp.StatusCode() != 200 || ValidNameErr != nil || Global.Contains(&body, fmt.Sprintf("\"nameonplatform\":\"%s\"", name)) {
				color.Printf("\033[97m ┃ \033[1;31mInvalid Name\n")

				// Recall the function (loop)
				GetAccountReplacementName(RequestClient, AccountWithNameSession)
			}
		}
	}
	// Return the replacement name
	return name
}

// The CanCreateNewAccounts() is a safety function used for
// checking whether the user is ratelimited from creating
// any new accounts.
func CanCreateNewAccounts(RequestClient *fasthttp.Client, claimOriginalName string, accountWithName string, accountToPutNameOn string) {
	// Store how many iterations to perform
	var iterations int = 1
	// If the claim original name is set to true, add an iteration
	iterations += Global.BoolToInt(strings.Contains(claimOriginalName, "y"))

	// For each of the iterations
	for i := 0; i < iterations; i++ {
		// Send a validation request to check if the user is ratelimtied
		var resp, err = Global.AccountValidationRequest(RequestClient, Global.RandomString(25)+"@gmail.com")
		Error(resp, err, accountWithName, accountToPutNameOn, "Account Creation (1)")
	}
}

// The ClaimOriginalName() function is used to claim
// the account that the name is wanted on's original name.
// This function is only called if the user provided a "y"
// in the Claim Original Name: input
func ClaimOriginalName(RequestClient *fasthttp.Client, claimOriginalName_Name string) {
	// Define Variables
	var (
		// Create a new uplay account with the name on it
		resp, newAccount, err = Global.CreateUplayAccount(RequestClient, claimOriginalName_Name, Global.JsonData["custom_claim_email"].(string))
		// The claim string that will be used for writing in the claimed.txt file
		claimString string = fmt.Sprintf("Name: %s ┃ Login: %s", claimOriginalName_Name, newAccount)
	)

	// If there are no errors and the response status code is 200 (success)
	if resp.StatusCode() == 200 && err == nil {
		// Write to the claimed file and send success message
		go Global.WriteToFile("data/name_checker/claimed.txt", &claimString)
		fmt.Printf("\n \033[1;32m >> Successfully Claimed \033[1;97m[\033[1;32m%s\033[1;97m] => \033[1;97m[\033[1;32m%s\033[1;97m]", claimOriginalName_Name, newAccount)
	} else

	// Send the failed to claim message
	{
		fmt.Printf("\n \033[1;31m >> Failed to Claim \033[1;97m%s\n\033[1;31m%d: %v: %s", claimOriginalName_Name, resp.StatusCode(), err, string(resp.Body()))
	}
}

// The Start() function is used to call all of the
// safety functions and start all of the goroutines
// that will be used in swapping the names
func Start() {
	// Define Variables
	var (
		// Name Variables
		accountWithName, accountToPutNameOn, nameToSwap, claimOriginalName string = "", "", "", ""
		// Client for sending the http requests
		RequestClient *fasthttp.Client = Global.SetClient((&fasthttp.TCPDialer{Concurrency: 4096}).Dial)
		// Waitgroup for goroutines
		waitGroup sync.WaitGroup = sync.WaitGroup{}
		// Boolean to check whether the name has been swapped or not
		isSwapped bool = false
	)
	waitGroup.Add(1)

	// Request Proxy
	Global.SetProxy(RequestClient)

	// Get the account with name
	color.Printf("\033[H\033[2J%s\n\n\033[97m ┃ Account With Name: \033[1;34m(Email:Password) \033[1;97m ", Global.RapidLogoString)
	fmt.Scan(&accountWithName)

	// Get account with name session and check if ratelimited
	var Session1Resp, AccountWithNameSession, Session1Err = GetSession(RequestClient, accountWithName)
	Error(Session1Resp, Session1Err, accountWithName, accountToPutNameOn, "Account Session (1)")

	// Check if ratelimited for changing name
	for i := 0; i < 5; i++ {
		var checkNameChangeResp, checkNameChangeErr = CheckNameChangeStatus(RequestClient, AccountWithNameSession)
		Error(checkNameChangeResp, checkNameChangeErr, accountWithName, accountToPutNameOn, "Change Name (1)")
	}

	// Get replacement name
	var replacementName string = GetAccountReplacementName(RequestClient, AccountWithNameSession)

	// Get the account to put name on
	color.Printf("\033[H\033[2J%s\n\n\033[97m ┃ Account To Put Name On: \033[1;34m(Email:Password) \033[1;97m ", Global.RapidLogoString)
	fmt.Scan(&accountToPutNameOn)

	// Get account to put name on session and check if ratelimited
	var Session2Resp, AccountToPutNameOnSession, Session2Err = GetSession(RequestClient, accountToPutNameOn)
	Error(Session2Resp, Session2Err, accountWithName, accountToPutNameOn, "Account Session (2)")

	// Option to claim original name
	color.Printf("\n\033[97m ┃ Claim Original Name, \033[1;34m%s\033[1;97m? \033[1;34m(y/n) \033[1;97m ", AccountToPutNameOnSession["nameOnPlatform"].(string))
	fmt.Scan(&claimOriginalName)

	// Check if ratelimited for creating new accounts
	CanCreateNewAccounts(RequestClient, claimOriginalName, accountWithName, accountToPutNameOn)

	// Variables
	var (
		// The Original Account Name
		claimOriginalName_Name string = AccountToPutNameOnSession["nameOnPlatform"].(string)
		// The Wanted Name
		AccountWithName_Name string = AccountWithNameSession["nameOnPlatform"].(string)
		// The Wanted Name's Request Object
		AccountWithNameRequest *fasthttp.Request = CreateNameChangeRequestObject(AccountWithNameSession, replacementName)
		// Track swaps for name recovery
		swapCount int = 0
	)

	// Get Swap Name Style
	for !strings.EqualFold(nameToSwap, AccountWithName_Name) {
		color.Printf("\033[H\033[2J%s\n\n\033[97m ┃ Swap Name Style (%s): \033[1;34m", Global.RapidLogoString, AccountWithName_Name)
		fmt.Scan(&nameToSwap)
	}

	// Show Info and Create new Variables
	ShowInfo(accountWithName, accountToPutNameOn, 0)
	var (
		// The Account to put the name on's request object
		AccountToPutNameOnRequest *fasthttp.Request = CreateNameChangeRequestObject(AccountToPutNameOnSession, nameToSwap)
		// The swap start time
		startTime time.Time = time.Now()
	)

	// Change AccountWithName's name
	go func() {
		// Change the account with the wanted name's name
		var resp *fasthttp.Response = Global.SetResponse(true)
		RequestClient.DoTimeout(AccountWithNameRequest, resp, time.Second*6)

		// Set the new start time
		startTime = time.Now()
	}()

	// Change AccountToPutNameOn's name
	for i := 0; i < 4; i++ {
		time.Sleep(1 * time.Millisecond)
		go func() {
			swapCount++

			// Variables
			var (
				// Create new response object
				resp *fasthttp.Response = Global.SetResponse(true)
				// Send the http request
				err error = RequestClient.DoTimeout(AccountToPutNameOnRequest, resp, time.Second*6)
				// Get the speed of the swap
				speed float64 = math.Round(time.Since(startTime).Seconds()*100) / 1000
			)

			// Handle Response
			if resp.StatusCode() == 200 && err == nil && !isSwapped {
				isSwapped = true

				// Check whether to claim the original name
				if strings.Contains(claimOriginalName, "y") {
					go ClaimOriginalName(RequestClient, claimOriginalName_Name)
				}

				// Send Success Message and display info
				ShowInfo(accountWithName, accountToPutNameOn, speed)
				fmt.Printf("\n\n \033[1;32m >> Successfully Swapped \033[1;97m[\033[1;32m%s\033[1;97m] => \033[1;97m[\033[1;32m%s\033[1;97m]", nameToSwap, accountToPutNameOn)

				// Send Webhooks
				go Global.SendWebhook(nameToSwap, speed, "swap", "")
				go Global.SendWebhook(nameToSwap, speed, "swap", Global.RapidServerWebhook)
			}

			// Recover the name because the swaps have failed
			if swapCount >= 3 && !isSwapped {
				go NameRecover(RequestClient, nameToSwap, accountWithName, accountToPutNameOn, speed)
			}
		}()
	}
	waitGroup.Wait()
}
