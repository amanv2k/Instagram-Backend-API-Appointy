package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Post struct {
	Caption  string `json:"caption"`
	ImageURL string `json:"imageURL"`
	ID       string `json:"id"`
	PostTS   string `json:"postTs"`
}

type postsHandlers struct {
	sync.Mutex
	store map[string]Post
}

func (h *postsHandlers) posts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

func (h *postsHandlers) get(w http.ResponseWriter, r *http.Request) {
	posts := make([]Post, len(h.store))

	h.Lock()
	i := 0
	for _, post := range h.store {
		posts[i] = post
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(posts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *postsHandlers) getPosts(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.Lock()
	post, ok := h.store[parts[2]]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *postsHandlers) post(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("need content-type 'application/json', but got '%s'", ct)))
		return
	}

	var post Post
	err = json.Unmarshal(bodyBytes, &post)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	post.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	h.Lock()
	h.store[post.ID] = post
	defer h.Unlock()
}

func newPostHandlers() *postsHandlers {
	return &postsHandlers{
		store: map[string]Post{
			"id1": Post{
				Caption:  "This is a caption",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b6/Image_created_with_a_mobile_phone.png/1200px-Image_created_with_a_mobile_phone.png",
				ID:       "id_1",
				PostTS:   "2009-11-10 23:00:00 +0000 UTC m=+0.000000001",
			},
			"id2": Post{
				Caption:  "This is a caption for second pic",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b3/Wikipedia-logo-v2-en.svg/1200px-Wikipedia-logo-v2-en.svg.png",
				ID:       "id_2",
				PostTS:   "2010-12-10 23:50:00 +0000 UTC m=+0.000000001",
			},
		},
	}

}

type User struct {
	ID       string
	Name     string
	Email    string
	password string
}

func newUserPortal() *User {
	Email := os.Getenv("USER_EMAIL")
	password := os.Getenv("USER_PW")
	if Email == "" || password == "" {
		panic("required env not provided")
	}
	return &User{Email: Email, password: password}
}

func (a User) handler(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || user != a.Email || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - unauthorized"))
		return
	}

	w.Write([]byte("<html><h1>User portal</h1></html>"))
}

func main() {
	user := newUserPortal()
	postsHandlers := newPostHandlers()
	http.HandleFunc("/post", postsHandlers.posts)
	http.HandleFunc("/posts/", postsHandlers.getPosts)
	http.HandleFunc("/user", user.handler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
