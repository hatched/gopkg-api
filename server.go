// Primary package for the gopkg api server.
package gopkg

import (
	"appengine"
	"appengine/datastore"
	"appengine/search"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Package struct {
	Url  string
	Date time.Time
}

type SearchResult struct {
	Id  string
	Url string
}

type Duplicate struct {
	Err string
}

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/add", addHandler)
	r.HandleFunc("/search/{query}", searchHandler)
	r.HandleFunc("/show", showHandler)
	http.Handle("/", r)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index, err := search.Open("Package")
	if err != nil {
		log.Print(err)
	}
	c := appengine.NewContext(r)
	packages := []*SearchResult{}
	for t := index.Search(c, vars["query"], nil); ; {
		var doc Package
		id, err2 := t.Next(&doc)
		if err2 == search.Done {
			break
		}
		if err2 != nil {
			log.Print(w, "Search error: %v\n", err2)
			break
		}
		log.Print(id)
		packages = append(packages, &SearchResult{
			Id:  id,
			Url: doc.Url,
		})
	}

	jsonPackages, err3 := json.Marshal(packages)
	if err3 != nil {
		http.Error(w, err3.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonPackages)
}

func showHandler(w http.ResponseWriter, r *http.Request) {

}

// Callback handler for the /add api endpoint.
func addHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
	}

	packageUrl := string(body)

	log.Print("adding: " + packageUrl)

	p := &Package{
		Url:  packageUrl,
		Date: time.Now(),
	}

	noneFound, err := searchForDuplicate(p)

	if err != nil {
		log.Print(err)
	}

	if noneFound {
		addToDatabase(p, r)
		addToIndex(p, w, r)
	} else {
		returnDuplicateWarning(p, w)
	}
}

// Returns a warning to the user if the package already exists in the index.
func returnDuplicateWarning(p *Package, w http.ResponseWriter) {
	w.WriteHeader(200)
	d := Duplicate{
		Err: "The package " + p.Url + " is already in the index.",
	}
	js, err := json.Marshal(d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Searches the database for the package they are trying to add so that there
// are not any duplicate entries.
func searchForDuplicate(p *Package) (bool, error) {

	return true, nil
}

// Adds a Package struct into the App Engine datastore.
func addToDatabase(p *Package, r *http.Request) {
	c := appengine.NewContext(r)
	key := datastore.NewIncompleteKey(c, "Package", nil)
	datastore.Put(c, key, p)
}

func addToIndex(p *Package, w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	index, err1 := search.Open("Package")
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}
	id, err2 := index.Put(c, "", p)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	log.Print("index id " + id)
}
