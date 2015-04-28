package rabbitmonit

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

func InitWebServer() {
	router := mux.NewRouter()
	router.HandleFunc("/", overview)
	http.ListenAndServe(":8080", router)
}

type OverviewStruct struct {
	Nodes  []*NodeProperties
	Queues []QueueProperties
}

func overview(writer http.ResponseWriter, request *http.Request) {
	var templates = template.New("template")
	filepath.Walk("template", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			templates.ParseFiles(path)
			return err
		}
		return nil
	})

	ops := &Ops{
		Host:     "http://localhost:15672",
		Login:    "guest",
		Password: "guest",
	}

	over := &OverviewStruct{
		Nodes:  ops.getClusterNodes(),
		Queues: ops.getAccumulationQueues(),
	}

	err := templates.ExecuteTemplate(writer, "overview", &over)
	if err != nil {
		panic(err.Error())
	}

	log.Println("Request served")
}
