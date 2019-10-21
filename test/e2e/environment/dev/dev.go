package main

import (
	"fmt"
	"os"

	"github.com/jetstack/cert-manager-csi/test/e2e/environment"
	"github.com/jetstack/cert-manager-csi/test/kind"
)

func main() {
	if len(os.Args) != 2 {
		errExit(fmt.Errorf("expecting 2 arguments, got=%d",
			len(os.Args)))
	}

	switch os.Args[1] {
	case "create":
		create()
	case "destroy":
		destroy()
	default:
		errExit(fmt.Errorf("unexpected argument %q, expecting %q or %q",
			os.Args[1], "create", "destroy"))
	}
}

func create() {
	env, err := environment.Create(1, 1)
	errExit(err)

	fmt.Printf("dev environment created.\nexport KUBECONFIG=%s\n",
		env.KubeConfigPath())
}

func destroy() {
	errExit(kind.DeleteFromName("cert-manager-csi-e2e"))
	fmt.Printf("dev environment destroyed.\n")
}

func errExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
