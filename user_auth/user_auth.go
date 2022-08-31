package user_auth

// Import Packages
import (
	"context"
	"crypto/tls"
	"os"
	"time"

	Global "rapid_name_claimer/global"

	"github.com/denisbrodbeck/machineid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// The ConnectToClient() function is used to
// establish a connection to the mongodb
func ConnectToClient() *mongo.Client {
	// Define Variables
	var (
		// Your mongo db uri
		uri string = "mongo client"

		// MongoDB Client Options
		clientOptions = options.Client().ApplyURI(uri).SetTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS12,
		})

		// Connect to the client
		client, err = mongo.Connect(context.TODO(), clientOptions)
	)

	// Ping the client to ensure it's working
	if client.Ping(context.TODO(), readpref.Primary()) != nil || err != nil {
		os.Exit(1)
	}

	// Return the client connection
	return client
}

// The SearchCollection function is used to search for a key
// and a value within the mongodb collection
func SearchCollection(coll *mongo.Collection, key string, value string) (bson.M, error) {
	var (
		res bson.M
		err = coll.FindOne(context.TODO(), bson.D{primitive.E{Key: key, Value: value}}).Decode(&res)
	)
	return res, err
}

// The ExistsInCollection() is used to check if key
// and value exist in a collection
func ExistsInCollection(coll *mongo.Collection, key string, value string) bool {
	var res, err = SearchCollection(coll, key, value)
	return err == nil && len(res) > 1
}

// The InsertIntoCollection() is used to insert data into a collection
func InsertIntoCollection(coll *mongo.Collection, document bson.D) (*mongo.InsertOneResult, error) {
	return coll.InsertOne(context.TODO(), document)
}

// The DeleteFromCollection() function is used to delete
// data from a collection by key and value
func DeleteFromCollection(coll *mongo.Collection, key string, value string) (*mongo.DeleteResult, error) {
	return coll.DeleteOne(context.TODO(), bson.D{primitive.E{Key: key, Value: value}})
}

// The GetHashedHWID() function is used to get
// he users hashed hwid
func GetHashedHWID() (string, error) {
	var hwid, err = machineid.ProtectedID("Rapid")
	return hwid, err
}

// The AuthenticationByToken() function is used to take
// in an authentication token provided by the user which
// will be used for adding the users hwid and data into
// the mongodb database.
func AuthenticationByToken(accessTokenCollection *mongo.Collection, hwidCollection *mongo.Collection, hwid string, token string) bool {

	// Check if the users hwid doesn't already exist in the database
	if !ExistsInCollection(hwidCollection, "hwid", hwid) {

		// Check if the token is valid
		if ExistsInCollection(accessTokenCollection, "token", token) {

			// Insert the token
			var _, tokenInsertionError = InsertIntoCollection(accessTokenCollection, bson.D{primitive.E{Key: "token", Value: Global.RandomString(24)}})
			if tokenInsertionError == nil {

				// Delete previous token from the collection
				var _, tokenRemovalError = DeleteFromCollection(accessTokenCollection, "token", token)
				if tokenRemovalError == nil {

					// Insert the new user and all their info
					var _, hwidInsertionError = InsertIntoCollection(hwidCollection, bson.D{
						primitive.E{Key: "hwid", Value: hwid},
						primitive.E{Key: "discord", Value: "unknown: token authentication"},
						primitive.E{Key: "token_used", Value: token},
						primitive.E{Key: "date", Value: time.Now().Format("Jan-02-2006")},
					})
					// Return whether an error has occured
					return hwidInsertionError == nil
				}
			}
		}
	}
	// Return false, (results in os.Exit())
	return false
}
