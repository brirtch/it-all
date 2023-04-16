package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Create a struct that models the structure of a user in the request body
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type PairRequestOpening struct {
	URL      string `json:"url"`
	LifeSpan string `json:"lifeSpan"`
}

type PairRequest struct {
	SourceURL string `json:"source_url"`
	DestURL   string `json:"dest_url"`
	Key       string `json:"key"`
}

// each session contains the username of the user and the time at which it expires
type session struct {
	username string
	expiry   time.Time
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the expected password from our in memory map
	expectedPassword, ok := users[creds.Username]

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create a new random session token
	// we use the "github.com/google/uuid" library to generate UUIDs
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	// Set the token in the session map, along with the session information
	sessions[sessionToken] = session{
		username: creds.Username,
		expiry:   expiresAt,
	}

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 120 seconds
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		SameSite: 4,
		Secure:   true,
	})
}

func authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We can obtain the session token from the requests cookies, which come with every request
		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sessionToken := c.Value

		// We then get the session from our session map
		userSession, exists := sessions[sessionToken]
		if !exists {
			// If the session token is not present in session map, return an unauthorized error
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// If the session is present, but has expired, we can delete the session, and return
		// an unauthorized status
		if userSession.isExpired() {
			delete(sessions, sessionToken)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Authentication middleware for an HTML page. Failure of
// authentication does a redirect to the / page.
func authenticatedPageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We can obtain the session token from the requests cookies, which come with every request
		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sessionToken := c.Value

		// We then get the session from our session map
		userSession, exists := sessions[sessionToken]
		if !exists {
			// If the session token is not present in session map, return an unauthorized error
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		// If the session is present, but has expired, we can delete the session, and return
		// an unauthorized status
		if userSession.isExpired() {
			delete(sessions, sessionToken)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	// remove the users session from the session map
	delete(sessions, sessionToken)

	// We need to let the client know that the cookie is expired
	// In the response, we set the session token to an empty
	// value and set its expiry as the current time
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now(),
		SameSite: 4,
		Secure:   true,
	})
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	// (BEGIN) The code from this point is the same as the first part of the `Welcome` route
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	userSession, exists := sessions[sessionToken]
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// (END) The code until this point is the same as the first part of the `Admin` route

	// If the previous session is valid, create a new session token for the current user
	newSessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	// Set the token in the session map, along with the user whom it represents
	sessions[newSessionToken] = session{
		username: userSession.username,
		expiry:   expiresAt,
	}

	// Delete the older session token
	delete(sessions, sessionToken)

	// Set the new token as the users `session_token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    newSessionToken,
		Expires:  time.Now().Add(120 * time.Second),
		SameSite: 4,
		Secure:   true,
	})
}

// POST /start-pairing - Client asks this node to pair with another.
func startPairing(w http.ResponseWriter, r *http.Request) {
	var pairRequest PairRequest

	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&pairRequest)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Send a pair message to url with key
	jsonBytes, _ := json.Marshal(pairRequest)
	request, err := http.NewRequest("POST", pairRequest.DestURL+"/pair", bytes.NewBuffer(jsonBytes))
	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

}

// POST /pair - Another node is trying to pair with this one.
func pair(w http.ResponseWriter, r *http.Request) {
	var pairRequest PairRequest

	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&pairRequest)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Print the details of the pairing request to the console.
	fmt.Println("PAIRING REQUEST FROM " + pairRequest.SourceURL + " with key " + pairRequest.Key)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	sql := "SELECT COUNT(*) FROM pair_request_opening WHERE key = ? AND url = ? AND expires < NOW()"
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(pairRequest.SourceURL, pairRequest.Key)

	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func PairRequestOpenings(w http.ResponseWriter, r *http.Request) {
	var pairRequestOpening PairRequestOpening

	err := json.NewDecoder(r.Body).Decode(&pairRequestOpening)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Store in the table.
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO pair_request_opening(url, created, expires, key) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(pairRequestOpening.URL, time.Now(), time.Now(), "polo")
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte(`{"result":"success"}`))
}
