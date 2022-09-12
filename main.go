package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"runtime"
	"time"

	AccountChecker "rapid_name_claimer/account_checker"
	Global "rapid_name_claimer/global"
	NameChecker "rapid_name_claimer/name_checker"
	NameSwapper "rapid_name_claimer/name_swapper"
	TokenGenerator "rapid_name_claimer/token_generator"

	"github.com/gookit/color"
)

// Within the init() function, check the users authentication
// and initalize the rand seed and max goroutines
func init() {
	// Set rand seed and max goroutines
	rand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())
}

////////////////////////////////////////////////
// PUT THIS INSIDE init() IF USING THE MONGODB
///////////////////////////////////////////////
/*
	// Check if user is authenticated
	var hwid, getHashedHWIDError = UserAuth.GetHashedHWID()
	if getHashedHWIDError != nil {
		os.Exit(2)
	}

	// Define Variables
	var (
		// Mongo client
		client *mongo.Client = UserAuth.ConnectToClient()

		// Mongo database
		database *mongo.Database = client.Database("hwid_data")

		// Mongo collection
		hwidCollection *mongo.Collection = database.Collection("ids")
	)

	// Check if the user is in the database
	if !UserAuth.ExistsInCollection(hwidCollection, "hwid", hwid) {
		// Access tokens
		var accessTokenCollection *mongo.Collection = database.Collection("access_tokens")

		// Get the authentication token from the user
		var token string
		color.Printf("\033[H\033[2J%s\n\n\033[97m ┃ Authentication Token: \033[1;34m", Global.RapidLogoString)
		fmt.Scanln(&token)

		// Check if the authentication token is valid
		if !UserAuth.AuthenticationByToken(accessTokenCollection, hwidCollection, hwid, token) {
			os.Exit(3)
		} else {
			// Successfully authenticated
			color.Printf("\033[H\033[2J%s\n\n\033[97m ┃ Visit \033[1;34mReadme.md\033[97m Before Pressing \033[1;34mEnter \033[97m", Global.RapidLogoString)
			fmt.Scanln()
		}
	}
*/

// Main function
func main() {
	Global.SetProxiesToCorrectFormat()

	// Define Variables
	var (
		// Thread Count
		threadCount int = 0
		// Selection
		option int
	)

	// Change Terminal Title
	exec.Command("cmd", "/C", "title", "Developed by tristan#2230").Run()

	// Print Rapid Logo
	color.Printf("\033[H\033[2J%s\n\n", Global.RapidLogoString)
	Global.GetCustomUrls()

	// User Select Option
	color.Print("\033[97m ┃  [\033[1;34m1\033[97m] Name Claimer\n ┃\n ┃  \033[97m[\033[1;34m2\033[97m] Name Swapper\n ┃\n ┃  \033[97m[\033[1;34m3\033[97m] Account Checker\n ┃\n ┃  \033[97m[\033[1;34m4\033[97m] Token Generator\n ┃\n\n   >> \033[1;34m")
	fmt.Scan(&option)

	// Get Thread Count
	if option == 1 || option == 3 {
		// If the inputted threadcount is 0 or greater than 100
		for threadCount > 1000 || threadCount < 1 {
			color.Printf("\033[H\033[2J%s\n\033[1;34m ┃ \033[1;31mMax 100 Threads\n\n\033[97m ┃ How many threads?\033[1;34m ", Global.RapidLogoString)
			fmt.Scan(&threadCount)
		}
	} else

	// Get Token Count
	if option == 4 {
		color.Printf("\033[H\033[2J%s\n\n\033[97m ┃ How many tokens?\033[1;34m ", Global.RapidLogoString)
		fmt.Scan(&threadCount)
	}

	// Use option response
	switch option {
	case 1:
		NameChecker.Start(threadCount)
	case 2:
		NameSwapper.Start()
	case 3:
		AccountChecker.Start(threadCount)
	case 4:
		TokenGenerator.Start(threadCount)
	default:
		main()
	}
}
