package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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

// We also need to get and update the AccountBalance field in the DB after recording a new transaction
func handler(context context.Context, request events.APIGatewayProxyRequest)(events.APIGatewayProxyResponse, error){
	type Transaction struct {
		TransactionID 		string	`json:"TransactionID"`
		TransactionType 	string 	`json:"TransactionType"`
		Amount      		string 	`json:"TransactionAmount"`
		TransactionStatus	string 	`json:"TransactionStatus"`
		TransactionDate   	string 	`json:"TransactionDate"`
		AccountBalance		float64	`json:"AccountBalance"`
	}

	type TransactionRecord struct {
		UserID     		string  		`json:"UserId"`
		NewTransaction 	Transaction 	`json:"NewTransaction"`
	}

	var newTransaction TransactionRecord

	err := json.Unmarshal([]byte(request.Body), &newTransaction)
	if err != nil {
        log.Fatal("Error marshalling user struct to JSON: ", err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500, // internal server error
            Body:       err.Error(),
        }, nil
    }

	// Get the current account balance from the User table
	var accountBalance float64
	query := "SELECT AccountBalance FROM Users Where UserId = ?"
	var err2 = db.QueryRow(query, newTransaction.UserID).Scan(&accountBalance)
	if err2 != nil {
        log.Fatal("Error executing query: ", err)
    }

	//Update the new transaction before updating the table
	newTransaction.NewTransaction.TransactionID = uuid.New().String()
	newTransaction.NewTransaction.AccountBalance = accountBalance
	newTransaction.NewTransaction.TransactionStatus = "Pending"
	newTransaction.NewTransaction.TransactionDate = time.Now().Format("2006-01-02 15:04:05")

    stmt, err := db.Prepare("INSERT INTO TransactionHistory(TransactionId, UserId, TransactionType, Amount, TransactionStatus, TransactionDate, AccountBalance) VALUES (?,?,?,?,?,?,?)")

    if err != nil {
        return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error preparing query: %v", err)
    }
    defer stmt.Close()

    _, err = stmt.Exec(newTransaction.NewTransaction.TransactionID, newTransaction.UserID, newTransaction.NewTransaction.TransactionType, 
		newTransaction.NewTransaction.Amount, newTransaction.NewTransaction.TransactionStatus, newTransaction.NewTransaction.TransactionDate, 
		newTransaction.NewTransaction.AccountBalance)
    if err != nil {
        return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error executing query: %v", err)
    }

    // Successful response
    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Body:       fmt.Sprintf("Successfully recorded a new transaction for userID: %s", newTransaction.UserID),
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