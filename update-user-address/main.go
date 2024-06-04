package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

	// Verify the connectionSDsdads
	err = db.Ping()
	if err != nil {          
		log.Fatal(err)
	}
	
	fmt.Println("Connected to the MySQL database successfully!")
}

func handler(context context.Context, request events.APIGatewayProxyRequest)(events.APIGatewayProxyResponse, error){
	
	type Address struct {
		AddressID int    `json:"AddressID"`
		UserID    int    `json:"UserID"`
		Street    string `json:"Street"`
		City      string `json:"City"`
		State     string `json:"State"`
		ZipCode   string `json:"ZipCode"`
		Country   string `json:"Country"`
	}

	type AddressUpdate struct {
		UserID     string  `json:"UserId"`
		NewAddress Address `json:"NewAddress"`
	}

	var updateAddress AddressUpdate

	err := json.Unmarshal([]byte(request.Body), &updateAddress)
	if err != nil {
        log.Fatal("Error marshalling user struct to JSON: ", err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500, // internal server error
            Body:       err.Error(),
        }, nil
    }
	

	// Prepare update statement
    stmt, err := db.Prepare("UPDATE UserAddress SET Street=?, City=?, State=?, ZipCode=?, Country=? WHERE userId=?")
    if err != nil {
        return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error preparing query: %v", err)
    }
    defer stmt.Close()

    // Execute update statement
    _, err = stmt.Exec(updateAddress.NewAddress.Street, updateAddress.NewAddress.City, updateAddress.NewAddress.State, updateAddress.NewAddress.ZipCode, updateAddress.NewAddress.Country, updateAddress.UserID)
    if err != nil {
        return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error executing query: %v", err)
    }

    // Successful response
    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Body:       fmt.Sprintf("User %s address updated successfully", updateAddress.UserID),
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