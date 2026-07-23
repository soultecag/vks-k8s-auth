package k8s_utils_test

import (
	"fmt"

	"github.com/soultecag/vks-k8s-auth/pkg/k8s_utils"
	"k8s.io/client-go/rest"
)

func ExampleConvertRESTConfigToKubeconfig() {
	kubeconfig, err := k8s_utils.ConvertRESTConfigToKubeconfig(
		"example-cluster",
		"example-user",
		"example-context",
		&rest.Config{
			Host:        "https://10.0.0.10:6443",
			BearerToken: "example-token",
			TLSClientConfig: rest.TLSClientConfig{
				CAData: []byte("example-ca-data"),
			},
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(kubeconfig != "")
	// Output:
	// true
}
