package main

import (
	"bufio"
	"log"
)

func read_server_message(reader *bufio.Reader, c chan string) {
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		log.Println("message reçu : " + message)
		c <- message
	}
}

func send_message_server(writer *bufio.Writer, c chan string) {
	for {

		message := <-c

		if message != "" && message != "\n" {
			_, err := writer.WriteString(message + "\n")
			if err != nil {
				log.Fatal(err)
			}

			err = writer.Flush()
			if err != nil {
				log.Fatal(err)
			}

			log.Println("message envoyé : " + message)
		} else {
			log.Println("j'ai essayé d'envoyé du rien")
		}
	}
}
