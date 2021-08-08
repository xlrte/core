package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}
	fmt.Println(boo)

	fmt.Println("Terraform planned")

	fmt.Println(*state.Values.Outputs["cloud-run-url"])

	fmt.Println(state.FormatVersion) // "0.1"

}
