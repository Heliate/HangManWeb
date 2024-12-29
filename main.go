package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// Structures nécessaires
type Jeu struct {
	MotSecret       string
	MotAffiche      string
	LettresEssayees map[string]bool
	ViesRestantes   int
	Score           int
	Pseudo          string
}

type Score struct {
	Pseudo string
	Score  int
}

var (
	mots         = []string{"fromage", "chocolat", "ordinateur", "programmation", "soleil", "lune", "ciel", "mer", "montagne", "voiture", "train", "avion", "livre", "stylo", "table", "chaise", "fenetre", "porte", "mur", "plafond", "sol", "jardin", "parc", "foret", "animal", "oiseau", "poisson", "chien", "chat", "maison", "villa", "appartement", "ecole", "universite", "bureau", "hôpital", "pharmacie", "magasin", "supermarche", "restaurant", "cafe", "cinema", "theatre", "musee", "concert", "hotel", "plage", "camping", "village", "ville", "capitale"}
	jeuActuel    *Jeu
	scores       = []Score{}
	templates    = template.Must(template.ParseGlob("templates/*.html"))
	scoreFichier = "scores.json"
)

func init() {
	// Charger les scores
	fichier, err := os.ReadFile(scoreFichier)
	if err == nil {
		json.Unmarshal(fichier, &scores)
	}
	// Initialisation aléatoire
	rand.Seed(time.Now().UnixNano())
}

func main() {
	http.HandleFunc("/", pageAccueil)
	http.HandleFunc("/jeu", pageJeu)
	http.HandleFunc("/fin", pageFin)
	http.HandleFunc("/scores", pageScores)
	http.HandleFunc("/action", gestionAction)
	http.HandleFunc("/continuer", continuerJeu)

	fmt.Println("Serveur démarré sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Page d'accueil
func pageAccueil(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.ExecuteTemplate(w, "accueil.html", nil)
	}
	if r.Method == "POST" {
		pseudo := r.FormValue("pseudo")
		jeuActuel = &Jeu{
			MotSecret:       mots[rand.Intn(len(mots))],
			LettresEssayees: make(map[string]bool),
			ViesRestantes:   7,
			Pseudo:          pseudo,
		}
		jeuActuel.MotAffiche = strings.Join(strings.Split(strings.Repeat("_", len(jeuActuel.MotSecret)), ""), " ")
		http.Redirect(w, r, "/jeu", http.StatusSeeOther)
	}
}

// Page de jeu
func pageJeu(w http.ResponseWriter, r *http.Request) {
	if jeuActuel == nil || jeuActuel.ViesRestantes <= 0 || !strings.Contains(jeuActuel.MotAffiche, "_") {
		http.Redirect(w, r, "/fin", http.StatusSeeOther)
		return
	}
	templates.ExecuteTemplate(w, "jeu.html", jeuActuel)
}

// Page de fin
func pageFin(w http.ResponseWriter, r *http.Request) {
	if jeuActuel == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	gagne := !strings.Contains(jeuActuel.MotAffiche, "_")
	if gagne {
		jeuActuel.Score += jeuActuel.ViesRestantes * 10
	} else {
		ajouterScore(jeuActuel.Pseudo, jeuActuel.Score)
		jeuActuel.Score = 0 // Reset le score en cas de défaite
	}
	templates.ExecuteTemplate(w, "fin.html", map[string]interface{}{
		"Gagne": gagne,
		"Jeu":   jeuActuel,
		"Mot":   jeuActuel.MotSecret, // Ajouter le mot en cas de défaite
	})
	if !gagne {
		jeuActuel = nil
	}
}

// Continuer à jouer
func continuerJeu(w http.ResponseWriter, r *http.Request) {
	jeuActuel.MotSecret = mots[rand.Intn(len(mots))]
	jeuActuel.MotAffiche = strings.Join(strings.Split(strings.Repeat("_", len(jeuActuel.MotSecret)), ""), " ")
	jeuActuel.LettresEssayees = make(map[string]bool)
	jeuActuel.ViesRestantes = 7
	http.Redirect(w, r, "/jeu", http.StatusSeeOther)
}

// Page des scores
func pageScores(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "scores.html", scores)
}

// Gestion des actions (entrer une lettre ou deviner le mot entier)
func gestionAction(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		proposition := strings.ToLower(r.FormValue("lettre"))
		if len(proposition) > 1 { // Proposition d'un mot entier
			if proposition == jeuActuel.MotSecret {
				jeuActuel.MotAffiche = jeuActuel.MotSecret
			} else {
				jeuActuel.ViesRestantes--
			}
		} else { // Proposition d'une lettre
			lettre := proposition
			if len(lettre) != 1 || lettre < "a" || lettre > "z" || jeuActuel.LettresEssayees[lettre] {
				http.Redirect(w, r, "/jeu", http.StatusSeeOther)
				return
			}
			jeuActuel.LettresEssayees[lettre] = true
			if strings.Contains(jeuActuel.MotSecret, lettre) {
				motTemp := []rune(jeuActuel.MotAffiche)
				for i, c := range jeuActuel.MotSecret {
					if string(c) == lettre {
						motTemp[i*2] = c // Modifier la lettre correspondante
					}
				}
				jeuActuel.MotAffiche = string(motTemp)
			} else {
				jeuActuel.ViesRestantes--
			}
		}
		http.Redirect(w, r, "/jeu", http.StatusSeeOther)
	}
}

// Ajouter un score à la fin de la session
func ajouterScore(pseudo string, score int) {
	scores = append(scores, Score{Pseudo: pseudo, Score: score})
	sauvegarderScores()
}

// Sauvegarder les scores dans un fichier
func sauvegarderScores() {
	fichier, _ := json.MarshalIndent(scores, "", "  ")
	os.WriteFile(scoreFichier, fichier, 0644)
}
