package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/joho/godotenv"
)

var rabbit_host string
var rabbit_port string
var rabbit_user string
var rabbit_password string

func main() {

	godotenv.Load()

	rabbit_host = os.Getenv("RABBIT_HOST")
	rabbit_port = os.Getenv("RABBIT_PORT")
	rabbit_user = os.Getenv("RABBIT_USERNAME")
	rabbit_password = os.Getenv("RABBIT_PASSWORD")

	router := httprouter.New()

	router.POST("/publish/:message", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		submit(w, r, p)
	})

	fmt.Println("Running...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func submit(writer http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	message := p.ByName("message")

	fmt.Println("Received message: " + message)

	conn, err := amqp.Dial("amqp://" + rabbit_user + ":" + rabbit_password + "@" + rabbit_host + ":" + rabbit_port + "/")

	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to RabbitMQ", err)
	}

	defer conn.Close()

	ch, err := conn.Channel()

	if err != nil {
		log.Fatalf("%s: %s", "Failed to open a channel", err)
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		"publisher", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare a queue", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})

	if err != nil {
		log.Fatalf("%s: %s", "Failed to publish a message", err)
	}

	fmt.Println("publish success!")
	var response string
	if err != nil {
		log.Fatalf("Failed to publish a message: %s", err)
		response = fmt.Sprintf("Failed to publish message: %s\n", message)
	} else {
		// Create the response body with RabbitMQ host and message details
		response = fmt.Sprintf("Message published: %s\nRabbitMQ Host: %s\n", message, rabbit_host)

	}

	// Send the response back to the client
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(response)) // Send the response to the client
}
