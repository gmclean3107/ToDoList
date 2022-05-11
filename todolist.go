package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"
)

var db, _ = gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct {
	Id          int `gorm: "primary_key; auto_increment; not_null"`
	Description string
	Completed   bool
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	description := r.FormValue("description")
	log.WithFields(log.Fields{"description": description}).Info("Add new TodoItem. Saving to database.")
	todo := &TodoItemModel{Description: description, Completed: false}
	db.Create(&todo)
	result := db.Last(&todo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Value)
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	err := GetItemById(id)
	if !err {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "Record not found"}`)
	} else {
		description := r.FormValue("description")
		completed, _ := strconv.ParseBool(r.FormValue("completed"))
		log.WithFields(log.Fields{"Id": id, "Completed": completed}).Info("Updating item")
		todoitem := &TodoItemModel{}
		db.First(&todoitem, id)
		todoitem.Completed = completed
		todoitem.Description = description
		db.Save(&todoitem)
		w.Header().Set("Content-Type", "json/application")
		io.WriteString(w, `{"updated": true}`)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	err := GetItemById(id)
	if !err {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": false, "error": "Record not found"}`)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Deleting Item")
		todoitem := &TodoItemModel{}
		db.First(&todoitem, id)
		db.Delete(&todoitem)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
	}
}

func GetItemById(Id int) bool {
	todoitem := &TodoItemModel{}
	result := db.First(&todoitem, Id)
	if result.Error != nil {
		log.Warn("Item not found in database!")
		return false
	}
	return true
}

func GetToDoItems(completed bool) interface{} {
	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func GetCompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Getting completed items")
	completedToDos := GetToDoItems(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedToDos)
}

func GetIncompleteItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Getting incomplete items")
	incompleteToDos := GetToDoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incompleteToDos)
}

func Health(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func main() {
	defer db.Close()

	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})

	log.Info("Starting Todolist API server")
	router := mux.NewRouter()
	router.HandleFunc("/health", Health).Methods("GET")
	router.HandleFunc("/createitem", CreateItem).Methods("POST")
	router.HandleFunc("/deleteitem", DeleteItem).Methods("POST")
	router.HandleFunc("/updateitem", UpdateItem).Methods("POST")
	router.HandleFunc("/getcompleted", GetCompletedItems).Methods("GET")
	router.HandleFunc("/getincompleted", GetIncompleteItems).Methods("GET")
	http.ListenAndServe(":8000", router)
}
