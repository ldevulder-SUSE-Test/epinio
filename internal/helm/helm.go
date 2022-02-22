package helm

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/epinio/epinio/internal/application"
	"github.com/epinio/epinio/pkg/api/core/v1/models"
	hc "github.com/mittwald/go-helm-client"
)

const (
	StandardChart = "https://github.com/epinio/helm-charts/releases/download/epinio-application-0.1.1/epinio-application-0.1.1.tgz"
	// ChartName:   "/path/to/stable/etcd-operator",			Local directory
	// ChartName:   "/path/to/stable/etcd-operator.tar.gz",			Local archive
	// ChartName:   "http://helm.whatever.com/repo/etcd-operator.tar.gz",	Remote archive
)

type ChartParameters struct {
	models.AppRef                                // Application: name & namespace
	Chart         string                         // Chart to use for deployment
	ImageURL      string                         // Application Image
	Username      string                         // User causing the (re)deployment
	Instances     int32                          // Number Of Desired Replicas
	Stage         models.StageRef                // Stage ID for ImageURL
	Owner         metav1.OwnerReference          // App CRD Owner Information
	Environment   models.EnvVariableMap          // App Environment
	Services      application.AppServiceBindList // Bound Services
	// TODO 1224 HELM : TODO: routes field
}

func Deploy(parameters ChartParameters) error {

	// TODO 1224 HELM : process routes
	//
	// YAML string - TODO ? Use unstructured as intermediary to
	// marshal yaml from ? Instead of direct generation of a
	// string ?

	serviceNames := "~"
	if len(parameters.Services) > 0 {
		serviceNames = fmt.Sprintf(`["%s"]`, strings.Join(parameters.Services.ToNames(), `","`))
	}

	environment := "~"
	if len(parameters.Environment) > 0 {
		environment = fmt.Sprintf(`{ %s }`, strings.Join(parameters.Environment.List().Assignments(), ","))
	}

	yamlParameters := fmt.Sprintf(`
epinio:
  replicaCount: %[1]d
  appUID: "%[2]s"
  stageID: "%[3]s"
  imageURL: "%[4]s"
  username: "%[5]s"
  routes: ~
  env: %[7]s
  services: %[6]s
`, parameters.Instances,
		parameters.Owner.UID,
		parameters.Stage.ID,
		parameters.ImageURL,
		parameters.Username,
		serviceNames,
		environment,
	)

	chartSpec := hc.ChartSpec{
		ReleaseName: parameters.Name,
		ChartName:   parameters.Chart,
		Namespace:   parameters.Namespace,
		Wait:        true,
		ValuesYaml:  yamlParameters,
	}

	client, err := hc.New(&hc.Options{
		Namespace:        parameters.Namespace, // Match chart spec
		RepositoryCache:  "/tmp/.helmcache",    // Hopefully reduces chart downloads.
		RepositoryConfig: "/tmp/.helmrepo",     // s.a.
	})
	if err != nil {
		return err
	}

	fmt.Printf("\nZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ\n")

	resolved, err := client.TemplateChart(&chartSpec)
	if err != nil {
		return err
	}

	fmt.Printf("ZZZ_______________________________\n%s_______________________________ZZZZ\n\n", string(resolved))

	// if _, err := client.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
	// 	return err
	// }

	return nil
}
