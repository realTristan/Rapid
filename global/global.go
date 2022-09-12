package global

// Import Packages
import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	Queue "rapid_name_claimer/queue"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// Define Global Variables
var (
	// Current Error
	CurrentError string

	// Client for sending discord webhooks
	WebhookRequestClient *fasthttp.Client = SetClient()

	// All the banned names from banned_names.txt
	BannedNames, _ = ioutil.ReadFile("data/name_checker/banned_names.txt")

	// Custom Url Queue
	CustomUrlQueue *Queue.ItemQueue = Queue.Create()

	// Mapped data from the data.json file
	JsonData map[string]interface{} = ReadJsonFile("data/data.json")

	// Rapid Variables
	RapidServerWebhook string = "Webhook URL"
	RapidLogoString    string = "\033[1;34m\n ┃ ██████╗  █████╗ ██████╗ ██╗██████╗\n ┃ ██╔══██╗██╔══██╗██╔══██╗██║██╔══██╗\n ┃ ██████╔╝███████║██████╔╝██║██║  ██║\n ┃ ██╔══██╗██╔══██║██╔═══╝ ██║██║  ██║\n ┃ ██║  ██║██║  ██║██║     ██║██████╔╝\n ┃ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚═════╝"
)

// Function to check if char is alpha
func IsAlphaChar(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

// Function to convert bool to int
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Function to check if uplay name is valid
func IsValidUplayName(name string) bool {
	var bannedNames []byte = bytes.ToLower(BannedNames)
	return len(name) > 2 && IsAlphaChar(rune(name[0])) && !bytes.Contains(bannedNames, []byte(name))
}

// The AddToQueue() function is used to read each line in
// the provided file and add each line to it's corresponding
// queue. (which was provided in the func params)
func AddToQueue(q *Queue.ItemQueue, fileName string) *Queue.ItemQueue {
	var file *bufio.Scanner = ReadFile(fileName)

	// Iterate over the files lines
	for file.Scan() {
		var value string = file.Text()

		// If value is not invalid add it to the queue
		if len(value) > 0 {
			q.Put(value)
		}
	}
	return q
}

// The FileNewLineCount() function is used to count
// how many lines are in the provided file
func FileNewLineCount(fileName string) int {
	// Define Variables
	var (
		// Counter variable
		count int = 0
		// File data scanner
		file *bufio.Scanner = ReadFile(fileName)
	)

	// Iterate over the files liones
	for file.Scan() {
		count++
	}
	return count
}

// The ReadJsonFile() function is used to read the provided
// json file then unmarshal it's data into a readable map
func ReadJsonFile(fileName string) map[string]interface{} {
	// Define Variables
	var (
		// Result map
		result map[string]interface{}
		// Read the json file
		jsonFile, _  = os.Open(fileName)
		byteValue, _ = ioutil.ReadAll(jsonFile)
	)
	// Close the json file once the function returns
	defer jsonFile.Close()

	// Unmarshal the data to the result map
	// then return said map
	json.Unmarshal([]byte(byteValue), &result)
	return result
}

// The ReadFile() function is used for reading all
// the content within the provided file.
func ReadFile(fileName string) *bufio.Scanner {
	// Open the file
	var file, _ = os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)

	// Return a scanner for the file
	return bufio.NewScanner(file)
}

// The WriteToFile() function is used to append the
// provided data to the provided file
func WriteToFile(fileName string, data string) {
	// Open the file
	var file, err = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)

	// Set the CurrentError if any errors have occured
	if err != nil {
		CurrentError = fmt.Sprintf(" >> Write To File Error: %s: %s: %v", fileName, data, err)
	} else {
		// Write the data to the file
		file.WriteString("\n" + data)
	}
	// Close the file
	file.Close()
}

// The OverWriteToFile() function is used to replace
// all the data in the provided file with the provided data
func OverwriteToFile(fileName string, data *string) error {
	return ioutil.WriteFile(fileName, []byte(*data), 0644)
}

// The RandomString() function is used to generate
// a random string using the characters defined below
func RandomString(length int) string {
	var (
		// Characters used in the random string
		chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
		// byte result
		b []byte = make([]byte, length)
	)
	// For the provided range
	for i := 0; i < length; i++ {
		// Set the index in the byte to a random character
		b[i] = chars[rand.Intn(len(chars))]
	}
	// Return the randomized string
	return string(b)
}

// The SetClient() function is used to establish a new
// fasthttp request client used for sending
// any http request
func SetClient() *fasthttp.Client {
	return &fasthttp.Client{
		Dial:                (&fasthttp.TCPDialer{Concurrency: 4096}).Dial,
		TLSConfig:           &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:     4096,
		ReadTimeout:         time.Second * 5,
		WriteTimeout:        time.Second * 5,
		MaxIdleConnDuration: time.Second * 6,
	}
}

// The SetResponse() function is used to create
// a new fasthttp response object
func SetResponse(skipBody bool) *fasthttp.Response {
	var resp *fasthttp.Response = fasthttp.AcquireResponse()
	resp.SkipBody = skipBody
	return resp
}

// The SetRequest() function is used to return a new
// fasthttp request object. This request object is used
// for sending any ubisoft-api-based requests
func SetRequest(method string) *fasthttp.Request {
	var req *fasthttp.Request = fasthttp.AcquireRequest()

	// Set the default headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 5_1 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9B179 Safari/7534.48.3")
	req.Header.Set("Ubi-AppId", "5e814037-2687-4164-9cd2-cc1b28652a9a")
	req.Header.Set("GenomeId", "de726b45-417f-476f-a3ba-d0c032a9ef2e")
	req.Header.Set("Ubi-RequestedPlatformType", "uplay")
	req.Header.Set("Connection", "keep-alive")
	// Set the content type and method
	req.Header.SetContentType("application/json")
	req.Header.SetMethod(method)
	// Return the request
	return req
}

// The GetCustomUrl() function is used to return
// the api endpoint url required for sending
// the http requests.
//
// If there are no custom urls, return the default
// ubisoft api endpoint
func GetCustomUrl() string {
	if CustomUrlQueue.IsNotEmpty() {
		return CustomUrlQueue.Get().(string)
	}
	return "https://public-ubiservices.ubi.com/"
}

// The GetCustomUrls() function is used to get all
// the custom urls provided in the data.json file
func GetCustomUrls() {
	var urls []interface{} = JsonData["custom_api_urls"].([]interface{})

	// For each of the custom urls
	for i := 0; i < len(urls); i++ {
		// Make sure the url is valid
		if strings.Contains(urls[i].(string), "https") {
			// Add a slash to the end of the url
			// if there isn't one already
			var tempUrl string = urls[i].(string)
			if string([]rune(tempUrl)[len(tempUrl)-1]) != "/" {
				tempUrl += "/"
			}
			// Add the url to the custom url queue
			CustomUrlQueue.Put(tempUrl)
		}
	}
}

// The GenerateUplayAccountJSON() function is used to establish the
// request body map that is required for creating a new
// uplay account
func GenerateUplayAccountJSON(name string, customEmail string) ([]byte, string) {
	// If the custom email is invalid
	if !strings.Contains(customEmail, "@") {
		// Create a random one
		customEmail = RandomString(25-len(name)) + "@gmail.com"
	}

	// Define Variables
	var (
		// New Account Email
		email string = name + "." + customEmail
		// New Account Password
		password string = "rapd" + RandomString(10)
		// The request body map
		data, _ = json.Marshal(map[string]interface{}{
			"age":               "19",
			"confirmedEmail":    email,
			"email":             email,
			"country":           "CA",
			"firstName":         "Rapid Claimer",
			"lastName":          "tristan#2230",
			"nameOnPlatform":    name,
			"password":          password,
			"preferredLanguage": "en",
			"legalOptinsKey":    "eyJ2dG91IjoiNC4wIiwidnBwIjoiNC4wIiwidnRvcyI6IjIuMCIsImx0b3UiOiJlbi1DQSIsImxwcCI6ImVuLUNBIiwibHRvcyI6ImVuLUNBIn0",
		})
	)

	// Return the data and the account combo
	return data, email + ":" + password
}

// The CreateUplayAccount() function is used to create a new uplay account
// using the provided, request client, name, and custom email.
//
// Once the account has been created, it will write it's auth token
// to the tokens.txt file so it can be used later for name checking
func CreateUplayAccount(RequestClient *fasthttp.Client, name string, customEmail string) (*fasthttp.Response, string, error) {
	// Request Proxy
	SetProxy(RequestClient)

	// Define Variables
	var (
		// Create new request object
		req *fasthttp.Request = SetRequest("POST")
		// Get the request body and the account being created
		bodyData, account = GenerateUplayAccountJSON(name, customEmail)
	)
	defer fasthttp.ReleaseRequest(req)

	// Set the Request url and body
	req.SetRequestURI(GetCustomUrl() + "v3/users")
	req.SetBody(bodyData)

	// Acquire response and do request
	var (
		// Create a new response object
		resp *fasthttp.Response = SetResponse(false)
		// Send the http request
		err error = RequestClient.DoTimeout(req, resp, time.Second*6)
		// Store the request body in a string variable
		body string = strings.ToLower(string(resp.Body()))
	)

	// If the Name is Banned
	if strings.Contains(body, "severity") {
		// Write it to the banned_names.txt file
		WriteToFile("data/name_checker/banned_names.txt", name)
	} else

	// If the status code is 200 (success) and no errors occured
	if resp.StatusCode() == 200 && err == nil {
		// Write the token to the data/tokens/tokens.txt file
		var jsonData map[string]interface{}
		json.Unmarshal([]byte(body), &jsonData)

		// If the ticket isn't empty/nil
		if jsonData["ticket"] != nil {
			// Write it to the tokens.txt file
			WriteToFile("data/tokens/tokens.txt", "Ubi_v1 t="+jsonData["ticket"].(string))
		}
	}
	// Return the response object, the account with the name on it, and any errors
	return resp, account, err
}

// The AccountValidationRequest() function is used to check whether
// the provided email is valid or not
func AccountValidationRequest(RequestClient *fasthttp.Client, email string) (*fasthttp.Response, error) {
	// Request Proxy
	SetProxy(RequestClient)

	// Define Variables
	var (
		// Function for generating the request body
		GenerateBody func(email string) []byte = func(email string) []byte {
			var data, _ = json.Marshal(map[string]interface{}{
				"email":             email,
				"legalOptinsKey":    "eyJ2dG91IjoiNC4wIiwidnBwIjoiNC4wIiwidnRvcyI6IjIuMCIsImx0b3UiOiJlbi1DQSIsImxwcCI6ImVuLUNBIiwibHRvcyI6ImVuLUNBIn0",
				"age":               "19",
				"password":          "RapidCheck!",
				"country":           "CA",
				"preferredLanguage": "en",
			})
			return data
		}
		// Create a new request object
		req *fasthttp.Request = SetRequest("POST")
	)
	defer fasthttp.ReleaseRequest(req)

	// Set the request url and body
	req.SetRequestURI(GetCustomUrl() + "v3/users/validatecreation")
	req.SetBody(GenerateBody(email))

	// Define Variables
	var (
		// Create new response object
		resp *fasthttp.Response = SetResponse(false)
		// Send the http request
		err error = RequestClient.DoTimeout(req, resp, time.Second*6)
	)
	return resp, err
}
