package PetitBac

import (
	"net/http"
)

type struct_PB struct {
	Code    string
	Actual  bool
	Players []string
	Round   int
	current int
	Letter  string
	Reponse map[string]map[string]string
}

var PetitBacGames = make(map[string]*struct_PB)

func Start(w http.ResponseWriter, r *http.Request) {

	var code string // generateCode() converti en string apres le merge

	PetitBacGames[code] = &struct_PB{
		Code:    code,
		Actual:  false,
		Players: []string{},
		Round:   5,
		current: 0,
		Letter:  "", //generateLetter()
		Reponse: make(map[string]map[string]string),
	}
	http.Redirect(w, r, "/petit-bac?code="+code, http.StatusSeeOther)
}

//generer une lettre
// creer une verification que le mot commence par la bonne lettre
// creer les catégories de mots...
