package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB
var gameServer GameServer

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

type Column struct {
	Name string `json:"Name"`
}

type Sql struct {
	Sql string `json:"sql"`
}

// this map stores the users sessions. For larger scale applications, you can use a database or cache for this purpose
var sessions = map[string]session{}

func getRoot(w http.ResponseWriter, r *http.Request) {

	file, err := os.Open("static/login.html")

	fileContents, err := ioutil.ReadAll(file)
	if err == nil {
		w.Write(fileContents)
	}
}

func rowsToJSON(rows *sql.Rows) ([]byte, error) {
	columnTypes, err := rows.ColumnTypes()

	if err != nil {
		return nil, err
	}

	count := len(columnTypes)
	finalRows := []interface{}{}

	for rows.Next() {

		scanArgs := make([]interface{}, count)

		for i, v := range columnTypes {

			switch v.DatabaseTypeName() {
			case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
				scanArgs[i] = new(sql.NullString)
				break
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
				break
			case "INT4":
				scanArgs[i] = new(sql.NullInt64)
				break
			default:
				scanArgs[i] = new(sql.NullString)
			}
		}

		err := rows.Scan(scanArgs...)

		if err != nil {
			return nil, err
		}

		masterData := map[string]interface{}{}

		for i, v := range columnTypes {

			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
				masterData[v.Name()] = z.Bool
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				masterData[v.Name()] = z.String
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				masterData[v.Name()] = z.Int64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				masterData[v.Name()] = z.Float64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				masterData[v.Name()] = z.Int32
				continue
			}

			masterData[v.Name()] = scanArgs[i]
		}

		finalRows = append(finalRows, masterData)
	}

	z, err := json.Marshal(finalRows)
	return z, err
}

// POST /sql
func PostSql(w http.ResponseWriter, r *http.Request) {
	// Get Sql.
	var sql Sql
	// Get the JSON body and decode into Sql
	err := json.NewDecoder(r.Body).Decode(&sql)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	fmt.Println("This would run SQL: " + sql.Sql)

	rows, err := DB.Query(sql.Sql)
	defer rows.Close()
	message := ""
	rowJSON := ""
	headerJSON := ""
	if err != nil {
		message = "Error: " + err.Error()
	} else {
		message = " success"

		cols, err := rows.ColumnTypes()
		if err != nil {
			log.Fatal(err)
		}

		var columns []Column
		for _, col := range cols {
			print("Column: " + col.Name())
			columns = append(columns, Column{Name: col.Name()})
		}

		headerJSONBytes, err := json.Marshal(columns)
		if err != nil {
			log.Fatal(err)
		}

		rowJSONBytes, err := rowsToJSON(rows)
		if err != nil {
			log.Fatal(err)
		}
		headerJSON = string(headerJSONBytes)
		rowJSON = string(rowJSONBytes)
	}

	encodedMessage := new(strings.Builder)
	json.NewEncoder(encodedMessage).Encode(message)
	jsonResponse := []byte(fmt.Sprintf(`{"message":%s,"headers":%s,"rows":%s}`, encodedMessage.String(), headerJSON, rowJSON))
	w.Write(jsonResponse)
}

// GET /admin
func Admin(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("secure/admin.html")

	fileContents, err := ioutil.ReadAll(file)
	if err == nil {
		w.Write(fileContents)
	}

}

func postCommand(w http.ResponseWriter, r *http.Request) {
	var cmd Command
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&cmd)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Parse command.
	commandPieces := ParseCommand(cmd.Command)

	if commandPieces[0] == "time" {
		t := time.Now()
		io.WriteString(w, t.Format("20060102150405"))
	} else if commandPieces[0] == "add-photo-folder" {
		folderToAdd := commandPieces[1]
		AddPhotoFolder(folderToAdd)
	} else {
		io.WriteString(w, "You said: "+cmd.Command+"\n")
	}

}

/* Database initialisation */
func dbInit() {
	fmt.Println("Initiating database...")
	var err error
	DB, err = sql.Open("sqlite3", "./db.db")
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()

	var sqlStatements []string
	sqlStatements = append(sqlStatements, `CREATE TABLE pair_request_opening (url text, created, expires, key);`)
	sqlStatements = append(sqlStatements, `CREATE TABLE photo_folder(folder TEXT, date_added DATETIME, date_last_scanned DATETIME, photo_count INT, state TEXT);`)
	sqlStatements = append(sqlStatements, `CREATE TABLE photo(photo_id INTEGER PRIMARY KEY,full_path TEXT,filename TEXT,bytes INT,parent_folder TEXT,date_taken DATETIME)`)
	sqlStatements = append(sqlStatements, `CREATE TABLE game(game_id INTEGER PRIMARY KEY, game_name TEXT,map_width INT,map_height INT)`)
	sqlStatements = append(sqlStatements, `CREATE TABLE player(player_id INTEGER PRIMARY KEY,player_name,game_id INT REFERENCES game(game_id),food INT,wood INT)`)
	sqlStatements = append(sqlStatements, `CREATE TABLE person(person_id INTEGER PRIMARY KEY,player_id INT REFERENCES player(player_id),role TEXT,gender TEXT,health INT)`)
	sqlStatements = append(sqlStatements, `CREATE TABLE map_object(map_object_id INTEGER PRIMARY KEY,game_id INT REFERENCES game(game_id),x INT,y INT,type TEXT,health INT)`)
	sqlStatements = append(sqlStatements, `CREATE TABLE player_map(player_map_id INTEGER PRIMARY KEY,player_id INT REFERENCES player(player_id),x INT,y INT,type TEXT)`)
	for _, sqlStatement := range sqlStatements {
		_, err = DB.Exec(sqlStatement)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStatement)
			return
		}
	}

}

func main() {
	var port int
	flag.IntVar(&port, "port", 3333, "Port number. Default is 3333")
	flag.Parse()

	fmt.Printf("Welcome to IT\n")

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Current working directory: %s\n", dir)

	dbInit()

	gameServer := NewGameServer()
	go gameServer.Run()
	go PhotoWatch()

	// If the data folder doesn't exist, create it.
	path := "data"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	r := chi.NewRouter()
	r.Get("/", getRoot)
	r.Post("/command", postCommand)
	r.Post("/login", Login)
	r.Get("/logout", Logout)
	r.Post("/start-pairing", startPairing)
	r.Post("/pair", pair)
	r.Post("/api/pairing-request-opening", PairRequestOpenings)

	r.Route("/photos", func(r chi.Router) {
		r.Get("/", PhotosHandler)
		r.Get("/{photoID}", ImageServeHandler)
		r.Post("/folders", NewPhotoFolderHandler)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(authenticatedPageMiddleware)
		r.Get("/", Admin)
	})

	r.Route("/sql", func(r chi.Router) {
		r.Use(authenticateMiddleware)
		r.Post("/", PostSql)
	})

	/* Game routes */
	r.Post("/game/create", gameServer.CreateGameHandler)
	r.Post("/game/join", gameServer.JoinGameHandler)
	r.Get("/game/games", gameServer.GetGamesHandler)
	r.Get("/game/{gameID}/{playerID}/status", gameServer.GetGameStatusHandler)
	r.Post("/game/buy", gameServer.BuyHandler)
	r.Post("/game/attack", gameServer.AttackHandler)
	r.Get("/game/gameObjectLibrary", gameServer.GetGameObjectLibraryHandler)
	r.Post("/game/message", gameServer.SendMessageHandler)

	r.Post("/refresh", Refresh)
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	fmt.Printf("IT is running at localhost:%d\n", port)
	err = http.ListenAndServe(":"+strconv.Itoa(port), r)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}

	DB.Close()
}
