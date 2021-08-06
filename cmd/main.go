package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"plugin"

	"github.com/go-playground/validator/v10"
	"github.com/xlrte/core/pkg/api"
	"gopkg.in/yaml.v2"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-exec/tfinstall"
)

// a Runtime has a Driver, that supports potentially resources and network stanzas?
// for resources not supported by driver, fall back on environment config?
// As in static resource defined per env? or have some way of mixing in other resources?

func main() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	// tmpDir, err := ioutil.TempDir("", "tfinstall")
	// if err != nil {
	// 	log.Fatalf("error creating temp dir: %s", err)
	// }
	// defer os.RemoveAll(tmpDir)

	execPath, err := tfinstall.Find(context.Background(), &tfinstall.LookPathOption{}) //tfinstall.LatestVersion(tmpDir, false))
	if err != nil {
		log.Fatalf("error locating Terraform binary: %s", err)
	}
	fmt.Println("found Terraform")

	workingDir := "deployment"
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	fmt.Println("Terraform started")

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	fmt.Println("Terraform init")

	state, err := tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	fmt.Println("Terraform state read")
	boo, err := tf.Plan(context.Background())
	fmt.Println(boo)

	fmt.Println("Terraform planned")

	fmt.Println(*state.Values.Outputs["cloud-run-url"])

	fmt.Println(state.FormatVersion) // "0.1"

}

func readYa() {
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
