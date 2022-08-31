package name_checker

// Import Packages
import (
	"bufio"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	Global "rapid_name_claimer/global"
	Queue "rapid_name_claimer/queue"

	"github.com/gookit/color"
	"github.com/valyala/fasthttp"
)

// Variables
var (
	// Counter Variables
	errorCount, nameCount, claimedNameCount, availableNameCount int = 0, 0, 0, 0

	// Request Variables
	requestTempAmount, totalRequests int = 0, 1

	// Queues
	UrlQueue   *Queue.ItemQueue = Queue.Create()
	TokenQueue *Queue.ItemQueue = Global.AddToQueue(Queue.Create(), "data/tokens/tokens.txt")
)

// The CheckedTotalAdd() function is used to spoof
// the total request count. This helps with hiding
// how the names are checked.
func CheckedTotalAdd(requestTempAmount *int, nameAmount int) int {
	var randNum int = 0
	*requestTempAmount += nameAmount
	if *requestTempAmount > 1 {
		randNum = rand.Intn(*requestTempAmount-1) + 1
		*requestTempAmount -= randNum
	}
	return randNum
}

// The LiveCounter() function is used to display all of
// the stats for the claimer. This includes checks per second,
// available/claimed names, errors, proxy count, etc.
func LiveCounter(programStartTime *int64, threadCount *int) {
	var reqsPerSecond int64 = int64(totalRequests) / ((time.Now().Unix() - *programStartTime) + 1)
	color.Printf(
		"\033[H\033[2J\033[0;0H%s\n\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Names \033[1;97m[\033[1;33m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Threads \033[1;97m[\033[1;35m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Proxies \033[1;97m[\u001b[30m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m Available \033[1;97m[\033[1;32m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m Claimed \033[1;97m[\033[1;32m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m CPS \033[1;97m[\033[1;36m%d/s\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m\033[1;37m\033[1;34m Checked \033[1;97m[\033[1;33m%d\033[1;97m]\033[1;34m\n\033[1;37m ┃ \033[1;34m Errors \033[1;97m[\033[1;31m%d\033[1;97m]\033[1;34m\n\n \033[1;31m%s",
		Global.RapidLogoString, nameCount, *threadCount, Global.ProxyQueue.Size(), availableNameCount, claimedNameCount, reqsPerSecond, totalRequests, errorCount, Global.CurrentError)
}

// The CheckNamesRequest() function is used to send the http
// request to the ubisoft api. It will return the response of
// the provided url.
func CheckNamesRequest(RequestClient *fasthttp.Client, url *string, token *string) (*fasthttp.Response, error) {
	Global.SetProxy(RequestClient)

	// Request object
	var req *fasthttp.Request = Global.SetRequest("GET")
	defer fasthttp.ReleaseRequest(req)

	// Request data
	req.Header.Set("Authorization", *token)
	req.SetRequestURI(*url)

	// Acquire response and do request
	var (
		resp *fasthttp.Response = Global.SetResponse(false)
		err  error              = RequestClient.DoTimeout(req, resp, time.Second*6)
	)
	return resp, err
}

// The GenerateNameUrls() is used to generate the api urls which
// will be used for sending requests to.
//
// Once each url has 50 nameOnPlatform args, it will add the url
// to the url queue.
func GenerateNameUrls() {
	var (
		// The url endpoint
		url string = fmt.Sprintf("%sv3/profiles?platformType=uplay", Global.GetCustomUrl())

		// The amount of names in the names.txt file
		fileNameCount int64          = Global.FileNewLineCount("data/name_checker/names.txt")
		nameFile      *bufio.Scanner = Global.ReadFile("data/name_checker/names.txt")

		// Temp Variables
		tempNameCount       int64  = 0
		nameFileReplacement string = ""
	)

	// Iterate through the file
	for nameFile.Scan() {
		var name string = nameFile.Text()
		tempNameCount++
		if len(name) > 2 {

			// If the name is valid
			if Global.IsValidUplayName(name) {
				nameFileReplacement += fmt.Sprintf("%s\n", name)
				nameCount++
				url += fmt.Sprintf("&nameOnPlatform=%s", name)

				// Add the url the queue if 50 names has been reached
				if nameCount%50 == 0 || tempNameCount == fileNameCount {
					UrlQueue.Put(url)
					url = fmt.Sprintf("%sv3/profiles?platformType=uplay", Global.GetCustomUrl())
				}
			}
		}
	}
	// Replace all the invalid names from names.txt
	// with only the valid ones
	if len(nameFileReplacement) > 0 {
		go Global.OverwriteToFile("data/name_checker/names.txt", &nameFileReplacement)
	}
}

// The ClaimName() function is used to send an http request to
// ubisoft's create account api endpoint.
//
// If the status is 200 and the name has been claimed, it will
// write the name, email and password to the claimed.txt file
func ClaimName(RequestClient *fasthttp.Client, name string) {
	var (
		// Custom Claim Email from data.json
		customClaimEmail string = Global.JsonData["custom_claim_email"].(string)

		// Send the http request to the ubi create account endpoint
		resp, account, err = Global.CreateUplayAccount(RequestClient, name, customClaimEmail)

		// The claim string to write to the claimed.txt file
		claimString string = fmt.Sprintf("Name: %s ┃ Login: %s", name, account)
	)
	defer fasthttp.ReleaseResponse(resp)

	// Handle Response
	if resp.StatusCode() == 200 && err == nil {
		claimedNameCount++

		// Send webhooks
		go Global.SendWebhook(name, 0, "claim", "")
		go Global.SendWebhook(name, 0, "claim", Global.RapidServerWebhook)

		// Write to the claimed.txt file, the claim string
		Global.WriteToFile("data/name_checker/claimed.txt", &claimString)
	} else

	// Set the error to the name claims response text and status code
	{
		Global.CurrentError = fmt.Sprintf(" >> Claim Name Error: %d: %v: %s", resp.StatusCode(), err, string(resp.Body()))
		errorCount++
	}
}

// The HandleResponse() function is used to handle the
// name check http response. If the response status was
// 200 and there were no errors, it will iterate
// over the provided names and check too see
// which ones are present in the response body.
//
// The names that are NOT present in the body are available.
func HandleResponse(RequestClient *fasthttp.Client, resp *fasthttp.Response, url *string, names *[]string) {
	defer fasthttp.ReleaseResponse(resp)

	// Handle Response
	var body string = string(resp.Body())
	for i := 0; i < len(*names); i++ {

		// Check if the body contains the name
		if !Global.Contains(&body, fmt.Sprintf("\"nameonplatform\":\"%s\"", (*names)[i])) {
			availableNameCount++

			// If it doesn't , claim the name and write the name
			// to the available.txt file
			go ClaimName(RequestClient, (*names)[i])
			Global.WriteToFile("data/name_checker/available.txt", &((*names)[i]))
		}
	}
}

// The Start() function is used to start all of the
// goroutines that will be used for sending the http
// requests to the ubisoft endpoints
func Start(threadCount int) {
	// Generate the api urls
	GenerateNameUrls()

	// Define Variables
	var (
		// sync.waitgroup for goroutines
		waitGroup sync.WaitGroup = sync.WaitGroup{}

		// Track the checks per second
		programStartTime int64 = time.Now().Unix()
	)
	waitGroup.Add(1)

	// Initialize the threads
	for i := 0; i < threadCount; i++ {
		go func() {
			// Create a requestclient for sending requests
			var RequestClient *fasthttp.Client = Global.SetClient((&fasthttp.TCPDialer{Concurrency: 4096}).Dial)

			// While loop
			for {
				// Define Variables
				var (
					// The ubi api endpoint
					url string = fmt.Sprint(*UrlQueue.Get())

					// The Authorization token
					token string = fmt.Sprint(*TokenQueue.Get())

					// Get the slice of names
					names []string = strings.Split(url, "&nameOnPlatform=")[1:]

					// Get the randum number created by the checked count spoofer
					randNum = CheckedTotalAdd(&requestTempAmount, len(names))

					// Send the http request to the ubi api endpoint
					resp, err = CheckNamesRequest(RequestClient, &url, &token)
				)

				// Update The Live Counter
				totalRequests += randNum
				LiveCounter(&programStartTime, &threadCount)

				// if the error is nil and the status code is 200
				if err == nil && resp.StatusCode() == 200 {
					// Handle the response
					go HandleResponse(RequestClient, resp, &url, &names)
				} else

				// Set the CurrentError message
				{
					// Define Variables
					var (
						// The response body
						body string = string(resp.Body())

						// The error converted to a string
						errString string = fmt.Sprint(err)
					)

					// Make sure the response body doesn't contain any secretive data (aka: the api endpoint data)
					if !Global.Contains(&body, "nameonplatform") && !Global.Contains(&errString, "nameonplatform") {
						Global.CurrentError = fmt.Sprintf(" >> Name Check Error: %d: %s: %s", resp.StatusCode(), errString, body)
					}

					// Release the response and increase the error count
					fasthttp.ReleaseResponse(resp)
					errorCount += randNum
				}

				// For Proxyless
				if threadCount <= 3 {
					time.Sleep(time.Millisecond * time.Duration(120/threadCount))
				}
			}
		}()
	}
	waitGroup.Wait()
}
