package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	razorpay "github.com/razorpay/razorpay-go"
)

type PageVariables struct {
	OrderId string
	Email   string
	Name    string
	Amount  string
	Contact string
}

func Caller(amount int) string {
	statusChannel := make(chan string)
	router := gin.Default()
	router.LoadHTMLGlob("*.html")
	router.GET("/", func(c *gin.Context) {
		App(c, amount*100)
	})
	router.GET("/payment-fail", func(c *gin.Context) {
		fmt.Println("Payment Failed")
		PaymentFaliure(c, statusChannel)
	})

	router.GET("/payment-complete", func(c *gin.Context) {
		fmt.Println("Payment Page")
		PaymentSuccess(c, statusChannel)
	})
	srv := &http.Server{
		Addr:    ":8089",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	fmt.Println("Server is running on port 8089")

	// create a timer and exit the server in 1 minute
	time.AfterFunc(45*time.Second, func() {
		fmt.Println("Shutting down server...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Server gracefully stopped")
	})
	// Wait for a payment status to be sent to the channel or for a timeout
	select {
	case <-time.After(45 * time.Second):
		select {
		case status := <-statusChannel:
			return status
		default:
			return "no-status"
		}
	case <-time.After(1 * time.Minute):
		return "payment-failed"
	}

}

func App(c *gin.Context, amount int) {

	page := &PageVariables{}
	page.Amount = strconv.Itoa(amount)
	page.Email = "rahul.c@prograd.org"
	page.Name = "Rahul Chhabra"
	page.Contact = "7906936789"
	//Create order_id from the server
	client := razorpay.NewClient("rzp_test_p4bR8DXSKNi8tJ", "JDcoQd2EgIcodZR1vGparqCq")

	data := map[string]interface{}{
		"amount":   page.Amount,
		"currency": "INR",
		"receipt":  "some_receipt_id",
	}
	body, err := client.Order.Create(data, nil)
	fmt.Println("////////////////reciept", body)
	if err != nil {
		fmt.Println("Problem getting the repository information", err)
		os.Exit(1)
	}

	value := body["id"]

	str := value.(string)
	fmt.Println("str////////////////", str)
	HomePageVars := PageVariables{ //store the order_id in a struct
		OrderId: str,
		Amount:  page.Amount,
		Email:   page.Email,
		Name:    page.Name,
		Contact: page.Contact,
	}

	c.HTML(http.StatusOK, "app.html", HomePageVars)
}

func PaymentSuccess(c *gin.Context, statusChannel chan<- string) {
	fmt.Println("Payment Successfull")
	paymentid := c.Query("paymentid")
	orderid := c.Query("orderid")
	signature := c.Query("signature")

	fmt.Println(paymentid, "paymentid")
	fmt.Println(orderid, "orderid")
	fmt.Println(signature, "signature")

	statusChannel <- "Payment Successfull"
	fmt.Println("Payment Successfull")
	c.Redirect(http.StatusFound, "http://localhost:3000/orders")
}

func PaymentFaliure(c *gin.Context, statusChannel chan<- string) {
	statusChannel <- "Payment Failed"
	c.Redirect(http.StatusFound, "http://localhost:3000/orders")
}
