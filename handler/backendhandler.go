package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"github.com/skiarn/madmock/filesys"
	"github.com/skiarn/madmock/model"
)

// ViewDataHandler handles functionality to view mock data.
type ViewDataHandler struct {
	DataDirPath string
	Fs          filesys.FileSystem
}

//NewViewDataHandler handles initzialisation of ViewDataHandler.
func NewViewDataHandler(path string) ViewDataHandler {
	return ViewDataHandler{DataDirPath: path, Fs: filesys.LocalFileSystem{}}
}

//ViewDataHandler load resource data.
func (h *ViewDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/mock/api/mock/data/"):]

	d, err := h.Fs.ReadResource(h.DataDirPath + "/" + name + filesys.ContentEXT)
	if err != nil {
		http.Error(w, "Was not able to find resource: "+r.URL.String()+" Failed with error: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = io.Copy(w, d)
	if err != nil {
		http.Error(w, "Internal error while wringing response: "+r.URL.String()+" Failed with error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// MockCURDHandler handles create update read delete calls.
type MockCURDHandler struct {
	DataDirPath string
	Fs          filesys.FileSystem
}

//NewMockCURDHandler handles initzialisation of MockCURDHandler.
func NewMockCURDHandler(path string) MockCURDHandler {
	return MockCURDHandler{DataDirPath: path, Fs: filesys.LocalFileSystem{}}
}

func (h *MockCURDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
	case "POST":
		h.save(w, r)
	case "PUT":
		h.save(w, r)
	case "DELETE":
		h.delete(w, r)
	default:
		http.Error(w, "Unknown request method.", 400)
	}
}

func (h *MockCURDHandler) get(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/mock/api/mock/"):]

	d, err := h.Fs.ReadResource(h.DataDirPath + "/" + name + filesys.ConfEXT)
	if err != nil {
		http.Error(w, "Was not able to find resource: "+r.URL.String()+" Failed with error: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = io.Copy(w, d)
	if err != nil {
		http.Error(w, "Internal error while writing response: "+r.URL.String()+" Failed with error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

type mockDTO struct {
                        URI         string            `json:"uri"`
                        Method      string            `json:"method"`
                        ContentType string            `json:"contenttype"`
                        StatusCode  int               `json:"status"`
                        Body        string            `json:"body"`
                }

//Save a new mock entity on serverside. Will create new resource if not existing and if it already exist it will update the resource.
func (h *MockCURDHandler) save(w http.ResponseWriter, r *http.Request) {
	var c model.MockConf
	var body string
	paramErrors := make(map[string]string)
	if r.Header.Get("Content-Type") == "application/json" {
		var dto mockDTO
		err := json.NewDecoder(r.Body).Decode(&dto)
		if err != nil {
			paramErrors["json"] = err.Error()
		}
		c = model.MockConf{URI: dto.URI, Method: dto.Method, ContentType: dto.ContentType, StatusCode: dto.StatusCode}
		body = dto.Body
		defer r.Body.Close()
	} else {
		uri := r.FormValue("URI")
		method := r.FormValue("Method")
		contentType := r.FormValue("ContentType")
		body = r.FormValue("body")
		statuscodeFV := r.FormValue("StatusCode")
		statuscode, _ := strconv.Atoi(statuscodeFV)
		c = model.MockConf{URI: uri, Method: method, ContentType: contentType, StatusCode: statuscode}
	}
	isValid, paramErrors := c.Valid()

	if !isValid {	
		validationErrors, err := json.Marshal(paramErrors)
		if err != nil {
			log.Println(err)
		}
		http.Error(w, string(validationErrors), 400)
		return
	}

	err := h.Fs.WriteMock(c, []byte(body), h.DataDirPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/mock", http.StatusFound)
}

//Delete delete mock entity.
func (h *MockCURDHandler) delete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/mock/api/mock/"):]
	log.Println("Delete resource: ", name)
	//validate input!
	reg := regexp.MustCompile("[0-9A-Za-z_]+")
	match := reg.FindAllStringSubmatch(name, -1)
	if len(match) != 1 {
		http.Error(w, "Invalid request, name may only be [0-9A-Za-z_]: "+r.URL.String(), http.StatusBadRequest)
		return
	}

	dir := h.DataDirPath
	err := os.Remove(dir + "/" + name + filesys.ConfEXT)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	err = os.Remove(dir + "/" + name + filesys.ContentEXT)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	return
}
