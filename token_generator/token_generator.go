package token_generator

// Import Packages
import (
	"fmt"
	"os"
	"strings"
	"sync"

	Global "rapid_name_claimer/global"

	"github.com/gookit/color"
	"github.com/valyala/fasthttp"
)

// Define Global Variables
var (
	// Counter Variables
	errorCount, generatedCount int = 0, 0

	// Files
	tokenFile, _         = os.OpenFile("data/tokens/tokens.txt", os.O_APPEND|os.O_WRONLY, 0644)
	tokenAccountsFile, _ = os.OpenFile("data/tokens/token_accounts.txt", os.O_APPEND|os.O_WRONLY, 0644)
)

// The LiveCounter() function is used to display all of
// the stats for the token generator. This includes
// amount of tokens generated, errors, etc.
func LiveCounter(tokenCount int) {
	color.Printf(
		"\033[H\033[2J\033[0;0H%s\n\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Tokens \033[1;97m[\033[1;33m%d\033[1;97m]\033[1;34m   \033[1;37m\n ┃ \033[1;34m\033[1;37m\033[1;34m Proxies \033[1;97m[\u001b[30m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m Generated \033[1;97m[\033[1;32m%d\033[1;97m]\033[1;34m\033[1;37m\n ┃ \033[1;34m Errors \033[1;97m[\033[1;31m%d\033[1;97m]\033[1;34m\n\n \033[1;31m%s",
		Global.RapidLogoString, tokenCount, Global.ProxyQueue.Size(), generatedCount, errorCount, Global.CurrentError)
}

// The UsePreviousAccounts() function is used to send requests to the
// ubisoft api and get the new session token for each of the accounts
// in the token_accounts.txt file.
func UsePreviousAccounts(RequestClient *fasthttp.Client, tokenCount int) {
	for i := 0; i < Global.TokenAccountQueue.Size(); i++ {
		// Send the request to the ubi api
		var resp, token, err = Global.GetAuthTokenFromExistingAccount(RequestClient)
		defer fasthttp.ReleaseResponse(resp)

		// Check the request status and errors
		if resp.StatusCode() == 200 && err == nil && len(token) > 15 {
			// Write the token to the tokens.txt file
			// and increase the generated count
			tokenFile.WriteString("\n" + token)
			generatedCount++
		} else {
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
		RequestClient *fasthttp.Client = Global.SetClient()
		// Waitgroup for goroutines
		waitGroup *sync.WaitGroup = &sync.WaitGroup{}
		// Token Counter
		tokenCountTotal int = tokenCount
	)
	waitGroup.Add(1)

	// Option to use existing accounts for tokens
	var tokenGenUseExistingAccounts string
	color.Printf("\033[H\033[2J%s\n\n\033[97m ┃ Use Existing Accounts? \033[1;34m(y/n)\033[1;97m ", Global.RapidLogoString)
	fmt.Scan(&tokenGenUseExistingAccounts)

	// If the user said yes to using previous accounts
	if strings.Contains(tokenGenUseExistingAccounts, "y") && Global.TokenAccountQueue.IsNotEmpty() {
		// Update the token counter
		tokenCountTotal = tokenCount + Global.TokenAccountQueue.Size()
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
				name string = "rapd" + Global.RandomString(11)
				// Create a new account
				resp, account, err = Global.CreateUplayAccount(RequestClient, name, "")
			)
			defer fasthttp.ReleaseResponse(resp)

			// If no errors occured and the response status code is 200 (succes)
			if err == nil && resp.StatusCode() == 200 {
				// Write the new account to the token_account.txt file
				tokenAccountsFile.WriteString("\n" + account)
				generatedCount++
			} else
			// Set the current error and increase the error count
			{
				Global.CurrentError = fmt.Sprintf(" >> Token Error: %d: %v: %s", resp.StatusCode(), err, string(resp.Body()))
				errorCount++
			}
			// Show the token generator status and info
			LiveCounter(tokenCountTotal)
		}()
	}
	waitGroup.Wait()
}
