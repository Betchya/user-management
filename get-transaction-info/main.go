package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	query := "SELECT * from TransactionHistory where UserID = ?"
	userID := request.PathParameters["userId"]

	type Transaction struct {
		TransactionId 	string
		UserId			string
		TransactionType	string
		Amount			string
		Status 			string
		TransactionDate	string
		AccountBalance	string
	}

	rows, err := db.Query(query, userID)

	if err != nil {
        log.Fatal("Error executing query: ", err)
        return events.APIGatewayProxyResponse{
            StatusCode: http.StatusInternalServerError,
            Body:       err.Error(),
        }, nil
    }
    defer rows.Close()

    var usersTransactions []Transaction
 
	for rows.Next() {
        var userTransaction Transaction
        err = rows.Scan(
				&userTransaction.TransactionId,
				&userTransaction.UserId,
				&userTransaction.TransactionType,
				&userTransaction.Amount,
				&userTransaction.Status,
				&userTransaction.TransactionDate,
				&userTransaction.AccountBalance,
            )
        
        if err != nil {
            log.Fatal("Error scanning row: ", err)
            return events.APIGatewayProxyResponse{
                StatusCode: http.StatusInternalServerError,
                Body:       err.Error(),
            }, nil
        }

        usersTransactions = append(usersTransactions, userTransaction)
    }

    // Check for errors from iterating over rows
    if err = rows.Err(); err != nil {
        log.Fatal("Error iterating rows: ", err)
        return events.APIGatewayProxyResponse{
            StatusCode: http.StatusInternalServerError,
            Body:       err.Error(),
        }, nil
    }

    // Serialize (marshal) the users slice to JSON
    usersJSON, err := json.Marshal(usersTransactions)
    if err != nil {
        log.Fatal("Error marshalling users slice to JSON: ", err)
        return events.APIGatewayProxyResponse{
            StatusCode: http.StatusInternalServerError,
            Body:       err.Error(),
        }, nil
    }

    // Convert the JSON bytes to a string and use it as the response body
    return events.APIGatewayProxyResponse{
        StatusCode: http.StatusOK,
        Body:       string(usersJSON),
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