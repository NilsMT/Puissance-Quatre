package main

import (
	"bufio"
	"log"
	"net"
)

func recup_transmet_color(otherwriter *bufio.Writer, reader *bufio.Reader, color *string, ok *bool) {
	var col string = ""
	for *ok == false { //envoie message tant que pas ok
		m, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if m != "couleur ok\n" && m != "tu es 2\n" && m != "tu es 1\n" { //assigne couleur select si couleur
			col = m
			log.Println("j'ai cte couleur pour l'autre " + col)
			_, err = otherwriter.WriteString(m)
			if err != nil {
				log.Println(err)
				return
			}
			otherwriter.Flush()
		} else if m == "couleur ok\n" { //exit
			*color = col
			*ok = true
		}
	}
	log.Println("jme casse lets go")
}

func main() {
	var msg string

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println("listen error:", err)
		return
	}
	defer listener.Close()

	//CLIENT 1
	conn1, err := listener.Accept()
	if err != nil {
		log.Println("accept error:", err)
		return
	}
	reader1 := bufio.NewReader(conn1)
	writer1 := bufio.NewWriter(conn1)

	log.Println("client 1 : connecté")
	//CLIENT 2
	conn2, err := listener.Accept()
	if err != nil {
		log.Println("accept error:", err)
		return
	}
	reader2 := bufio.NewReader(conn2)
	writer2 := bufio.NewWriter(conn2)

	log.Println("client 2 : connecté")

	//////////////////

	// defer conn1.Close()
	// defer conn2.Close()

	///////////////// CLIENT 1 OK

	_, err = writer1.WriteString("connexion ok\n")
	if err != nil {
		log.Println(err)
		return
	}
	writer1.Flush()

	log.Println("je dit à client 1 connexion OK")

	///////////////// CLIENT 2 OK

	_, err = writer2.WriteString("connexion ok\n")
	if err != nil {
		log.Println(err)
		return
	}
	writer2.Flush()

	log.Println("je dit à client 2 connexion OK")

	var col1 string
	var col2 string
	var ok1 bool
	var ok2 bool
	///////////////// SELECT COULEUR CLIENT 1

	go recup_transmet_color(writer1, reader2, &col1, &ok1)
	go recup_transmet_color(writer2, reader1, &col2, &ok2)

	tkt := false
	for tkt == false { //attend jusqua ce que les deux soit select
		if ok1 == true && ok2 == true {
			log.Println("Alléuia")
			tkt = true
		}
	}

	log.Println("les deux on select")

	///////////////// CLIENT 1 C JOUEUR 1
	_, err = writer1.WriteString("tu es 1\n")
	if err != nil {
		log.Println(err)
		return
	}
	writer1.Flush()

	log.Println("je dit à client 1 qu'il est le p1")

	///////////////// CLIENT 2 C JOUEUR 2
	_, err = writer2.WriteString("tu es 2\n")
	if err != nil {
		log.Println(err)
		return
	}
	writer2.Flush()

	log.Println("je dit à client 2 qu'il est le p2")

	//////////////////
	//////////////////
	//////////////////
	//IN-GAME
	//////////////////
	//////////////////
	//////////////////

	for {
		///////////////// RECUP POS P1
		msg, err = reader1.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if msg == "gagné\n" || msg == "perdu\n" || msg == "égalisé\n" {

			statutoppose := map[string]string{
				"perdu\n":   "gagné\n",
				"égalisé\n": "égalisé\n",
				"gagné\n":   "perdu\n",
			}

			log.Println("p1 a", msg, "et p2 a", statutoppose[msg])
		} else {
			///////////////// TRANSMET POS P2
			if msg != "\n" && msg != "" {
				log.Println("p1 me dit qu'il a joué en X = " + msg)
				_, err = writer2.WriteString(msg)
				if err != nil {
					log.Println(err)
					return
				}
				writer2.Flush()
				log.Println("je dit à p2 que p1 a joué en X = " + msg)
			} else {

			}
		}

		///////////////// RECUP POS P2
		msg, err = reader2.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if msg == "gagné\n" || msg == "perdu\n" || msg == "égalisé\n" {
			statutoppose := map[string]string{
				"perdu\n":   "gagné\n",
				"égalisé\n": "égalisé\n",
				"gagné\n":   "perdu\n",
			}

			log.Println("p2 a", msg, "et p1 a", statutoppose[msg])
		} else {
			///////////////// TRANSMET POS P1
			if msg != "\n" && msg != "" {
				log.Println("p2 me dit qu'il a joué en X = " + msg)
				_, err = writer1.WriteString(msg)
				if err != nil {
					log.Println(err)
					return
				}
				writer1.Flush()
				log.Println("je dit à p1 que p2 a joué en X = " + msg)
			} else {

			}
		}

	}
}
