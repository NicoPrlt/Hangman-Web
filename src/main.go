package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"unicode"
)

type Game struct {
	WordToFind   string
	RandomLetter rune
	LettersFound []rune
	LettersTried []rune
	Attempts     int
	DisplayWord  string
	Image        string
	GameOver     bool
	GameResult   string
}

var game Game

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", handler)
	http.HandleFunc("/replay", replay)
	fmt.Println("Server is running on port localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		handlePost(w, r)
		return
	}

	if game.WordToFind == "" {
		word, randomLetter, err := chooseOneWord()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal Server Error: %v", err)
			return
		}

		game = Game{
			WordToFind:   word,
			RandomLetter: randomLetter,
			LettersFound: []rune{randomLetter},
			LettersTried: []rune{},
			Attempts:     10,
			DisplayWord:  "",
		}
		updateDisplayWord(&game)
	}

	data := struct {
		Game Game
	}{game}

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error: %v", err)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error: %v", err)
		return
	}
}

func updateImage(game *Game) {
	imagePath := fmt.Sprintf("./static/img/hangman-%d.png", 10-game.Attempts-1)
	game.Image = imagePath
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	if game.GameOver {
		http.Error(w, "The game is over. You cannot submit more letters.", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	letter := strings.TrimSpace(r.Form.Get("letter"))

	if letter == "" || len(letter) != 1 || !unicode.IsLetter(rune(letter[0])) {
		http.Error(w, "Invalid input. Please enter a single letter.", http.StatusBadRequest)
		return
	}

	guessedLetter := unicode.ToLower(rune(letter[0]))

	if containsLetter(game.LettersTried, guessedLetter) {
		http.Error(w, "You've already tried this letter.", http.StatusBadRequest)
		return
	}

	game.LettersTried = append(game.LettersTried, guessedLetter)

	wordRunes := []rune(game.WordToFind)

	if containsLetter(wordRunes, guessedLetter) {
		if !containsLetter(game.LettersFound, guessedLetter) {
			game.LettersFound = append(game.LettersFound, guessedLetter)
		}

		updateDisplayWord(&game)

		if strings.Index(game.DisplayWord, "_") == -1 {
			game.GameOver = true
			game.GameResult = "You won!"
		}

	} else {
		game.Attempts--

		updateImage(&game)

		if game.Attempts == 0 {
			game.GameOver = true
			game.GameResult = "You lost!"
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func updateDisplayWord(game *Game) {
	displayWord := ""
	for _, char := range game.WordToFind {
		if char == ' ' {
			displayWord += "  "
		} else if containsLetter(game.LettersFound, unicode.ToLower(char)) {
			displayWord += string(char) + " "
		} else {
			displayWord += "_ "
		}
	}
	displayWord = strings.TrimSpace(displayWord)
	game.DisplayWord = displayWord
}

func chooseOneWord() (string, rune, error) {
	wordsData, err := ioutil.ReadFile("words.txt")
	if err != nil {
		return "", 0, err
	}

	wordsList := strings.Split(string(wordsData), "\n")
	if len(wordsList) == 0 || (len(wordsList) == 1 && wordsList[0] == "") {
		return "", 0, fmt.Errorf("Aucun mot dans words.txt")
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(wordsList))
	randomWord := wordsList[randomIndex]

	randomLetter := rune(randomWord[0])

	return randomWord, randomLetter, nil
}

func containsLetter(letters []rune, letter rune) bool {
	for _, l := range letters {
		if l == letter {
			return true
		}
	}
	return false
}

func replay(w http.ResponseWriter, r *http.Request) {
	game = Game{}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
