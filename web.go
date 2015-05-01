package rabbitmonit

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-datculescu/rabbit-hole"
	"github.com/gorilla/mux"
)

func InitWebServer() {
	router := mux.NewRouter()
	router.HandleFunc("/", overview)
	router.HandleFunc("/queue/{vhostName}/{queueName}", queue)
	router.HandleFunc("/vhost/{vhostName}", vhost)
	http.ListenAndServe(":8080", router)
}

type OverviewStruct struct {
	Nodes  []*NodeProperties
	Queues []QueueProperties
	Queue  QueueProperties
	Vhost  rabbithole.VhostInfo
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

func queue(writer http.ResponseWriter, request *http.Request) {
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

	vars := mux.Vars(request)

	over := &OverviewStruct{
		Queue: ops.getQueue(vars["vhostName"], vars["queueName"]),
	}

	err := templates.ExecuteTemplate(writer, "queueMain", &over)
	if err != nil {
		panic(err.Error())
	}

	log.Println("Request served")
}
func vhost(writer http.ResponseWriter, request *http.Request) {
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

	vars := mux.Vars(request)

	over := &OverviewStruct{
		Queues: ops.getQueues(vars["vhostName"]),
		Vhost:  ops.Vhost(vars["vhostName"]),
	}

	err := templates.ExecuteTemplate(writer, "vhostMain", &over)
	if err != nil {
		panic(err.Error())
	}

	log.Println("Request served")
}
