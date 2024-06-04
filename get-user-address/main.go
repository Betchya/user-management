package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	//"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init(){
	// Read database credentials from environment variables
    dbUser := "admin"
    dbPassword := "Capstone2024"
    dbName := "user_management"
    dbHost := "betchya.cjq8sw40g4n7.us-west-2.rds.amazonaws.com"
    dbPort := "3306"

    // Build the DSN (Data Source Name)
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

    // Open a SQL database connection
    var err error
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }

    // Verify the connection
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }
	
    fmt.Println("Connected to the MySQL database successfully!")
}

func handler(context context.Context, request events.APIGatewayProxyRequest)(events.APIGatewayProxyResponse, error){
	query := "SELECT * from UserAddress where UserID = ?"
	userID := request.PathParameters["userId"]

    fmt.Println(request.Body)
    

	type UserAddress struct {
		AddressId 	string
		UserId 		string
		Street		string
		City		string
		State		string
		ZipCode		string
		Country		string
	}

	var userAddress UserAddress

	err := db.QueryRow(query, userID).Scan(
		&userAddress.AddressId,
		&userAddress.UserId,
		&userAddress.Street,
		&userAddress.City,
		&userAddress.State,
		&userAddress.ZipCode,
		&userAddress.Country,
	)

	if err != nil {
        log.Fatal("Error executing query: ", err)
    }

    fmt.Printf("Address Found for UserID: %s", userAddress.UserId)

    userJSON, err := json.Marshal(userAddress)
    if err != nil {
        log.Fatal("Error marshalling user struct to JSON: ", err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500, // internal server error
            Body:       err.Error(),
        }, nil
    }

    // Convert the JSON bytes to a string and use it as the response body
    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Body:       string(userJSON),
        Headers:    map[string]string{"Content-Type": "application/json"},
    }, nil

}

func main(){
	//lambda.Start(handler)

    /*
    To test locally.
    Delete or comment code below before uploading to AWS 
	And uncomment lambda.Start(handler)
    */
    file, err := os.ReadFile("event.json")
    if err != nil {
        fmt.Printf("Failed to read file: %s\n", err)
        return
    }

    // Unmarshal the JSON into an APIGatewayProxyRequest
    var request events.APIGatewayProxyRequest
    err = json.Unmarshal(file, &request)
    if err != nil {
        fmt.Printf("Failed to unmarshal request: %s\n", err)
        return
    }

    // Call the handler with the unmarshalled request
    ctx := context.Background()
    response, err := handler(ctx, request)
    if err != nil {
        fmt.Printf("Handler error: %s\n", err)
        return
    }

    // Print the response
    fmt.Printf("Handler response: %+v\n", response)

    db.Close()

}
