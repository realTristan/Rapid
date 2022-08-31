package global

// Import Packages
import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	Queue "rapid_name_claimer/queue"
	"strings"

	"github.com/valyala/fasthttp"
)

//////////////////////////////////////////////////////
// For the official queue system,
// visit https://github.com/realTristan/FastProxy
//////////////////////////////////////////////////////

// Global Proxy Queue Variable
var ProxyQueue *Queue.ItemQueue = Queue.Create()

// The ProxyDial struct holds 3 keys
// - Proxy *proxy -> The proxy to use
// - Address *string -> The incoming dial address
type proxyDial struct {
	proxy   string
	address string
}

// Base64 encode a string
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Function to return a fasthttp response object
//   - sB bool -> SkipBody
func SetProxyResponse(sB bool) *fasthttp.Response {
	// Acquire the response object
	var r *fasthttp.Response = fasthttp.AcquireResponse()

	// Skip the response body, makes request faster
	r.SkipBody = sB
	return r
}

// The GenerateConnectUrl() function will create the Dial Url for the proxy
func GenerateConnectUrl(proxyDial *proxyDial) (string, string) {
	var url string = fmt.Sprintf("CONNECT %s HTTP/1.1\r\n", proxyDial.address)

	// If the Proxy Contains an @ (user:pass authentication)
	if strings.Contains(proxyDial.proxy, "@") {

		// Split the proxy and encode the auth
		var (
			AuthProxySplit []string = strings.Split(proxyDial.proxy, "@")
			Auth           string   = Base64Encode(AuthProxySplit[0])
		)

		// Append the proxy authentication to the url
		url += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", Auth)
		proxyDial.proxy = AuthProxySplit[1]
	}
	url += "\r\n"
	return url, proxyDial.proxy
}

// The EstablishConnection() function is the dial function that returns
// a fasthttp.Dial object
//
// This function will create a connection url then send a connection request
func EstablishConnection(proxyDial *proxyDial) (*net.Conn, error) {
	// Generate a connection url and create a connection to the proxy
	var (
		ConnectionUrl, proxy = GenerateConnectUrl(proxyDial)
		Connection, err      = fasthttp.Dial(proxy)
	)

	// Connection Error
	if err != nil {
		return nil, err
	}

	// Write to the connection url
	if _, err := Connection.Write([]byte(ConnectionUrl)); err != nil {
		return nil, err
	}

	// Set the connection response, release it once resp is no longer needed
	var resp *fasthttp.Response = SetResponse(true)

	// Release response once the connection has been established
	defer fasthttp.ReleaseResponse(resp)

	// Connection Reader Error
	if err := resp.Read(bufio.NewReader(Connection)); err != nil {
		Connection.Close() // Close Connection
		return nil, err
	}

	// Establish Connection Failed
	if resp.Header.StatusCode() != 200 {
		Connection.Close() // Close Connection
		return nil, fmt.Errorf("unable to establish a connection to %s", proxy)
	}
	// Return Connection and no error
	return &Connection, nil
}

// Function to use a proxy dial [user:pass@proxy:port]
func HttpProxyDial(proxy *string) fasthttp.DialFunc {
	// Return the dial function
	return func(addr string) (net.Conn, error) {
		// Create ProxyDial struct object
		var proxyDial *proxyDial = &proxyDial{
			address: addr,
			proxy:   *proxy,
		}

		// Return the connection
		var Connection, err = EstablishConnection(proxyDial)
		return *Connection, err
	}
}

// The SetProxiesToCorrectFormat() function is used to
// set the each proxy in the proxies.txt file to the
// correct user:pass@ip:port format
func SetProxiesToCorrectFormat() {
	// Define Variables
	var (
		// Proxy file data
		file *bufio.Scanner = ReadFile("data/proxies.txt")

		// The string to replace the current proxies.txt
		// data with
		result string = ""
	)

	// For each line in the proxies.txt file
	for file.Scan() {

		// Check whether the proxy has authentication
		var proxy string = file.Text()
		if strings.Contains(proxy, "@") {

			// Define Variables
			var (
				// Split the proxy
				splitString []string = strings.Split(proxy, "@")

				// Variables for
				userPass string
				ipPort   string
			)
			// If the first split is the IP:Port
			if ContainsAmount(splitString[0], ".") == 3 {
				ipPort = splitString[0]
				userPass = splitString[1]
			} else {
				ipPort = splitString[1]
				userPass = splitString[0]
			}

			// Store formatted proxy in string
			proxy = userPass + "@" + ipPort
		}

		// Add the proxy to the result
		result += proxy + "\n"
		// Add the proxy to the queue
		ProxyQueue.Put(proxy)
	}

	// Replace all the proxies in the proxies file
	// with the newly formatted ones
	if len(result) > 0 {
		go OverwriteToFile("data/proxies.txt", &result)
	}
}
