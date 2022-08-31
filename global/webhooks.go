package global

// Import Packages
import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/valyala/fasthttp"
)

// The SendWebhook() function is used to send a webhook
// using the data found in data.json
func SendWebhook(name string, speed float64, opt string, url string) {
	// Establish a new request object.
	// Release it once the function returns
	var req *fasthttp.Request = fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Set the request headers and method
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 5_1 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9B179 Safari/7534.48.3")
	req.Header.Set("Content-Type", "application/json")
	req.Header.SetMethod("POST")

	// Marshal the webhook data from the data.json file
	var _jsonData, err = json.Marshal(JsonData)
	if err != nil {
		CurrentError = fmt.Sprintf(" >> Send Webhook Error (1): %s: %v", name, err)
		return
	}

	// Convert the jsonData to a string
	var jsonData string = string(_jsonData)

	// Add a timestamp to the webhook
	if gjson.Get(jsonData, fmt.Sprintf("name_%s_webhook_data.embeds.0.timestamp", opt)).Bool() {
		var newData, err = sjson.Set(jsonData, fmt.Sprintf("name_%s_webhook_data.embeds.0.timestamp", opt), time.Now().Format(time.RFC3339))

		// If an error has occured, set the current error
		// to said error
		if err != nil {
			CurrentError = fmt.Sprintf(" >> Send Webhook Error (2): %s: %v", name, err)
			return
		}

		// Replace the jsonData
		jsonData = newData
	}

	// Replace Embed Name
	if strings.Contains(jsonData, "{name}") {
		jsonData = strings.ReplaceAll(jsonData, "{name}", name)
	}

	// Replace Embed Speed
	if strings.Contains(jsonData, "{speed}") {
		jsonData = strings.ReplaceAll(jsonData, "{speed}", fmt.Sprintf("%vs", speed))
	}

	// Set body and request url
	req.SetBody([]byte(gjson.Get(jsonData, fmt.Sprintf("name_%s_webhook_data", opt)).String()))
	if len(url) > 1 {
		req.SetRequestURI(url)
	} else {
		req.SetRequestURI(gjson.Get(jsonData, fmt.Sprintf("name_%s_webhook_url", opt)).String())
	}

	// Set Response
	var resp *fasthttp.Response = SetResponse(true)
	defer fasthttp.ReleaseResponse(resp)

	// Send Request
	WebhookRequestClient.DoTimeout(req, resp, time.Second*6)
}
