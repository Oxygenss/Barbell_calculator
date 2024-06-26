package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Plate struct {
	Id int
	Weight float64
	Amount int
}


func HomePage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/Home_page/Home_page.html")
	tmpl.ExecuteTemplate(w, "index", nil)
}

func Add(w http.ResponseWriter, r *http.Request) {
	weightStr := r.FormValue("weight")
	amountStr := r.FormValue("amount")

	weight, err := strconv.ParseFloat(weightStr, 64)
	if err!= nil {
		http.Error(w, "Неверный формат веса", http.StatusBadRequest)
		return
	}

	amount, err := strconv.Atoi(amountStr)
	if err!= nil {
		http.Error(w, "Неверный формат количества", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./data")
	if err!= nil {
		panic(err)
	}
	defer db.Close()

	

	var plates = []Plate{}
	
	rows, err := db.Query("SELECT * FROM `Plates`")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var plate Plate 
		err = rows.Scan(&plate.Id, &plate.Weight, &plate.Amount)
		if err != nil {
			panic(err)
		}
		plates = append(plates, plate)
	}

	flag := false

	for i := 0; i < len(plates); i++ {
		if weight == plates[i].Weight {
			flag = true
			plates[i].Amount += amount
			_, err = db.Exec(fmt.Sprintf("UPDATE `Plates` SET `amount` = %d WHERE id = %d", plates[i].Amount, plates[i].Id))
			if err!= nil {
				panic(err)
			}
		}
	}

	if !flag {
		_, err = db.Exec(fmt.Sprintf("INSERT INTO `Plates` (`weight`, `amount`) VALUES ('%f', '%d')", weight, amount))
		if err!= nil {
			panic(err)
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func Calculate(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/Calculate/Calculate.html")
	weightStr := r.FormValue("weight")
	handleStr := r.FormValue("handle")

	weight, err := strconv.ParseFloat(weightStr, 64)
	if err!= nil {
		http.Error(w, "Неверный формат веса", http.StatusBadRequest)
		return
	}

	handle, err := strconv.ParseFloat(handleStr, 64)
	if err!= nil {
		http.Error(w, "Неверный формат количества", http.StatusBadRequest)
		return
	}

	if handle < 0 || weight <= 0 {
		http.Error(w, "Отрицательные значения", http.StatusBadRequest)
		return
	}

	if handle + 1 > weight {
		http.Error(w, "Слишком большая рукоять", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./data")
	if err!= nil {
		panic(err)
	}
	defer db.Close()

	

	var plates = []Plate{}
	
	rows, err := db.Query("SELECT * FROM `Plates`")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var plate Plate 
		err = rows.Scan(&plate.Id, &plate.Weight, &plate.Amount)
		if err != nil {
			panic(err)
		}
		plates = append(plates, plate)
	}

	sort.Slice(plates, func(i, j int) bool {
		return plates[i].Weight > plates[j].Weight
	})

	result := make(map[float64]int, len(plates))

	for i := 0; i < len(plates); i++ {
		result[plates[i].Weight] = 0

	}

	var sum float64

	value := (weight - handle - 1) / 2

	for i := 0; i < len(plates); i++ {
		for plates[i].Weight <= value && plates[i].Amount > 1 {
			result[plates[i].Weight] += 2
			value -= plates[i].Weight
			plates[i].Amount -= 2
			sum += plates[i].Weight

		}
	}

	var minWeight float64
	flag := false

	for i := 0; i < len(plates); i++ {
		if plates[i].Amount > 1 {
			flag = true
			minWeight = plates[i].Weight
			break
		}
	}

	for i := 0; i < len(plates); i++ {
		if plates[i].Weight < minWeight &&  plates[i].Amount > 1 {
			minWeight = plates[i].Weight
		}
	}

	if !flag {
		minWeight = -1
	}


	over := sum * 2 + minWeight * 2 + 1

	fmt.Println(over, weight)

	tmpl.ExecuteTemplate(w, "result", map[string]interface{}{
        "result": result,
        "value": value * 2,
		"minWeight": minWeight,
		"sum": sum * 2 + 1,
		"over": over - (weight - handle),
    })
}

func Plates(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/Plates/Plates.html")

	var plates = []Plate{}

	db, err := sql.Open("sqlite3", "./data")
	if err!= nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM `Plates`")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var plate Plate 
		err = rows.Scan(&plate.Id, &plate.Weight, &plate.Amount)
		if err != nil {
			panic(err)
		}
		plates = append(plates, plate)
	}
	
	tmpl.ExecuteTemplate(w, "plates", plates)
}


func handleFunc() {
    r := mux.NewRouter()

	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir("./templates"))))

	r.HandleFunc("/", HomePage)
	r.HandleFunc("/Add", Add)
	r.HandleFunc("/Calculate", Calculate)
	r.HandleFunc("/Plates", Plates)
	http.Handle("/", r)

	http.ListenAndServe("localhost:8080", r)
}


func main() {
	handleFunc()
}

	