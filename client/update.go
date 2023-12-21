package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Mise à jour de l'état du jeu en fonction des entrées au clavier.
func (g *game) Update() error {

	g.stateFrame++

	switch g.gameState {
	case titleState:
		if g.titleUpdate() {
			g.gameState++
			g.otherReady = false //il n'a pas select sa couleur
		}
	case colorSelectState:
		if g.colorSelectUpdate() {
			g.gameState++
			g.otherReady = false //il ne sait pas qui il est
		}
	case playState:
		g.tokenPosUpdate()
		var lastXPositionPlayed int
		var lastYPositionPlayed int
		if g.turn == p1Turn {
			lastXPositionPlayed, lastYPositionPlayed = g.p1Update()
		} else {
			lastXPositionPlayed, lastYPositionPlayed = g.p2Update()
		}
		if lastXPositionPlayed >= 0 {
			finished, result := g.checkGameEnd(lastXPositionPlayed, lastYPositionPlayed)
			if finished {
				g.result = result
				g.gameState++
			}
		}
	case resultState:
		if g.resultUpdate() {
			//dit au serveur qui a gagné
			if g.result == p1wins {
				g.writeChan <- "gagné\n"
			} else if g.result == p2wins {
				g.writeChan <- "perdu\n"
			} else {
				g.writeChan <- "égalisé\n"
			}
			g.reset()
			g.gameState = playState
		}
	}

	return nil
}

// Mise à jour de l'état du jeu à l'écran titre.
func (g *game) titleUpdate() bool {
	g.stateFrame = g.stateFrame % globalBlinkDuration

	if !g.otherReady {
		select {
		case <-g.readChan:
			g.otherReady = true
		default:
		}
	}

	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) && g.otherReady
}

// Mise à jour de l'état du jeu lors de la sélection des couleurs.
func (g *game) colorSelectUpdate() bool {

	if !g.colorSelected {
		col := g.p1Color % globalNumColorCol
		line := g.p1Color / globalNumColorLine

		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			col = (col + 1) % globalNumColorCol
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			col = (col - 1 + globalNumColorCol) % globalNumColorCol
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			line = (line + 1) % globalNumColorLine
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			line = (line - 1 + globalNumColorLine) % globalNumColorLine
		}

		precol := g.p1Color //envoie si new color selected
		g.p1Color = line*globalNumColorLine + col
		if precol != g.p1Color {
			g.writeChan <- strconv.Itoa(g.p1Color)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			// transmet couleur ici
			g.colorSelected = true
			g.writeChan <- "couleur ok"
		}
	}

	if !g.otherReady {

		//LIT COULEUR AUTRE (SI YA) + p1 p2
		select {
		case m := <-g.readChan:
			if m != "tu es 2\n" && m != "tu es 1\n" {
				//lit ici là
				mc := strings.TrimSuffix(m, "\n")
				col, _ := strconv.Atoi(mc)

				g.p2Color = col

				log.Println("l'autre à la couleur n°" + mc)

				if g.p2Color == g.p1Color {
					g.p2Color = (g.p2Color + 1) % globalNumColor
				}
			} else {
				log.Println("j'ai pos reçu une couleur")
				if m == "tu es 1\n" {
					g.otherReady = true
					log.Println("Jsuis le joueur 1, c'est à mon tour")
					g.turn = p1Turn
				} else if m == "tu es 2\n" {
					g.otherReady = true
					log.Println("Jsuis le joueur 2, c'est pas à mon tour")
					g.turn = p2Turn
				}
			}
		default:
		}
	}

	return g.otherReady && g.colorSelected
}

// Gestion de la position du prochain pion à jouer par le joueur 1.
func (g *game) tokenPosUpdate() {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.tokenPosition = (g.tokenPosition - 1 + globalNumTilesX) % globalNumTilesX
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.tokenPosition = (g.tokenPosition + 1) % globalNumTilesX
	}
}

// Gestion du moment où le prochain pion est joué par le joueur 1.

// Envoie à l'autre la position joué
func (g *game) p1Update() (int, int) {
	lastXPositionPlayed := -1
	lastYPositionPlayed := -1
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if updated, yPos := g.updateGrid(p1Token, g.tokenPosition); updated {
			g.turn = p2Turn
			lastXPositionPlayed = g.tokenPosition
			lastYPositionPlayed = yPos
			//ici
			p := strconv.Itoa(g.tokenPosition) + "\n"
			g.writeChan <- p
			//
		}
	}
	return lastXPositionPlayed, lastYPositionPlayed
}

// Gestion de la position du prochain pion joué par le joueur 2 et
// du moment où ce pion est joué.

// Lecture de la position de l'autre
func (g *game) p2Update() (int, int) {
	var position int
	var err error
	select {
	case pos := <-g.readChan:
		readpos := strings.TrimSuffix(pos, "\n")
		position, err = strconv.Atoi(readpos)

		if err != nil {
			log.Fatal(err)
		}
	default:
		return -1, -1
	}
	updated, yPos := g.updateGrid(p2Token, position)
	for ; !updated; updated, yPos = g.updateGrid(p2Token, position) {
		position = (position + 1) % globalNumTilesX
	}
	g.turn = p1Turn
	return position, yPos
}

// Mise à jour de l'état du jeu à l'écran des résultats.
func (g game) resultUpdate() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

// Mise à jour de la grille de jeu lorsqu'un pion est inséré dans la
// colonne de coordonnée (x) position.
func (g *game) updateGrid(token, position int) (updated bool, yPos int) {
	for y := globalNumTilesY - 1; y >= 0; y-- {
		if g.grid[position][y] == noToken {
			updated = true
			yPos = y
			g.grid[position][y] = token
			return
		}
	}
	return
}

// Vérification de la fin du jeu : est-ce que le dernier joueur qui
// a placé un pion gagne ? est-ce que la grille est remplie sans gagnant
// (égalité) ? ou est-ce que le jeu doit continuer ?
func (g game) checkGameEnd(xPos, yPos int) (finished bool, result int) {

	tokenType := g.grid[xPos][yPos]

	// horizontal
	count := 0
	for x := xPos; x < globalNumTilesX && g.grid[x][yPos] == tokenType; x++ {
		count++
	}
	for x := xPos - 1; x >= 0 && g.grid[x][yPos] == tokenType; x-- {
		count++
	}

	if count >= 4 {
		if tokenType == p1Token {
			return true, p1wins
		}
		return true, p2wins
	}

	// vertical
	count = 0
	for y := yPos; y < globalNumTilesY && g.grid[xPos][y] == tokenType; y++ {
		count++
	}

	if count >= 4 {
		if tokenType == p1Token {
			return true, p1wins
		}
		return true, p2wins
	}

	// diag haut gauche/bas droit
	count = 0
	for x, y := xPos, yPos; x < globalNumTilesX && y < globalNumTilesY && g.grid[x][y] == tokenType; x, y = x+1, y+1 {
		count++
	}

	for x, y := xPos-1, yPos-1; x >= 0 && y >= 0 && g.grid[x][y] == tokenType; x, y = x-1, y-1 {
		count++
	}

	if count >= 4 {
		if tokenType == p1Token {
			return true, p1wins
		}
		return true, p2wins
	}

	// diag haut droit/bas gauche
	count = 0
	for x, y := xPos, yPos; x >= 0 && y < globalNumTilesY && g.grid[x][y] == tokenType; x, y = x-1, y+1 {
		count++
	}

	for x, y := xPos+1, yPos-1; x < globalNumTilesX && y >= 0 && g.grid[x][y] == tokenType; x, y = x+1, y-1 {
		count++
	}

	if count >= 4 {
		if tokenType == p1Token {
			return true, p1wins
		}
		return true, p2wins
	}

	// egalité ?
	if yPos == 0 {
		for x := 0; x < globalNumTilesX; x++ {
			if g.grid[x][0] == noToken {
				return
			}
		}
		return true, equality
	}

	return
}
