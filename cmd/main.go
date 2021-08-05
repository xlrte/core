package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"plugin"

	"github.com/go-playground/validator/v10"
	"github.com/xlrte/core/pkg/api"
	"gopkg.in/yaml.v2"
)

// a Runtime has a Driver, that supports potentially resources and network stanzas?
// for resources not supported by driver, fall back on environment config?
// As in static resource defined per env? or have some way of mixing in other resources?

func main() {
	var service api.Service
	var http api.Http

	data, err := ioutil.ReadFile("examples/cloudrun-srv.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal([]byte(data), &service)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	bt, err := yaml.Marshal(service.Network["http"])
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal(bt, &http)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(http)

	validate := validator.New()

	if errs := validate.Struct(http); errs != nil {
		// values not valid, deal with errors here

		fmt.Println(errs)
	}

	if errs := validate.Struct(service.Artifact); errs != nil {
		// values not valid, deal with errors here

		fmt.Println(errs)
	}

	fmt.Println(os.Getwd())

	lang := "english"
	if len(os.Args) == 2 {
		lang = os.Args[1]
	}
	var mod string
	switch lang {
	case "english":
		mod = ".plugins/eng.so"
	case "chinese":
		mod = ".plugins/chi.so"
	default:
		fmt.Println("don't speak that language")
		os.Exit(1)
	}

	plug, err := plugin.Open(mod)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	symGreeter, err := plug.Lookup("Greeter")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var greeter Greeter
	greeter, ok := symGreeter.(Greeter)
	if !ok {
		fmt.Println("unexpected type from module symbol")
		os.Exit(1)
	}

	// 4. use the module
	greeter.Greet()
}

type Greeter interface {
	Greet()
}

// import (
// 	"fmt"
// 	"net/http"
// )

// func main() {
// 	http.HandleFunc("/", HelloServer)
// 	http.ListenAndServe(":8080", nil)
// }

// func HelloServer(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Gruzi mittenand, %s!", r.URL.Path[1:])
// }
