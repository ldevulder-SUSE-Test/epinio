package models

import (
	"github.com/epinio/epinio/internal/names"
)

const (
	EpinioStageIDPrevious   = "epinio.suse.org/previous-stage-id"
	EpinioStageIDLabel      = "epinio.suse.org/stage-id"
	EpinioStageBlobUIDLabel = "epinio.suse.org/blob-uid"

	ApplicationCreated = "created"
	ApplicationStaging = "staging"
	ApplicationRunning = "running"
	ApplicationError   = "error"
)

type ApplicationStatus string

type GitRef struct {
	Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
	URL      string `json:"repository"         yaml:"url"`
}

// App has all the application's properties, for at rest (Configuration), and active (Workload).
// The main structure has identifying information.
// It is used in the CLI and API responses.
// If an error is hit while constructing the app object, the Error attribute
// will be set to that.
type App struct {
	Meta          AppRef                   `json:"meta"`
	Configuration ApplicationUpdateRequest `json:"configuration"`
	Origin        ApplicationOrigin        `json:"origin"`
	Workload      *AppDeployment           `json:"deployment,omitempty"`
	Status        ApplicationStatus        `json:"status"`
	StatusMessage string                   `json:"statusmessage"`
	StageID       string                   `json:"stage_id,omitempty"` // staging id, last run
}

type PodInfo struct {
	Name        string `json:"name"`
	MemoryBytes int64  `json:"memoryBytes"`
	MilliCPUs   int64  `json:"millicpus"`
	CreatedAt   string `json:"createdAt,omitempty"`
	Restarts    int32  `json:"restarts"`
	Ready       bool   `json:"ready"`
}

// AppDeployment contains all the information specific to an active
// application, i.e. one with a deployment in the cluster.
type AppDeployment struct {
	// TODO: Readiness and Liveness fields?
	Active          bool                `json:"active,omitempty"` // app is > 0 replicas
	CreatedAt       string              `json:"createdAt,omitempty"`
	DesiredReplicas int32               `json:"desiredreplicas"`
	ReadyReplicas   int32               `json:"readyreplicas"`
	Replicas        map[string]*PodInfo `json:"replicas"`
	Username        string              `json:"username,omitempty"` // app creator
	StageID         string              `json:"stage_id,omitempty"` // staging id, running app
	Status          string              `json:"status,omitempty"`   // app replica status
	Routes          []string            `json:"routes,omitempty"`   // app routes
}

// NewApp returns a new app for name and namespace
func NewApp(name string, namespace string) *App {
	return &App{
		Meta: AppRef{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// AppRef returns a reference to the app (name, namespace)
func (a *App) AppRef() AppRef {
	return a.Meta
}

// AppList is a collection of app references
type AppList []App

// Implement the Sort interface for application slices

// Len (Sort interface) returns the length of the AppList
func (al AppList) Len() int {
	return len(al)
}

// Swap (Sort interface) exchanges the contents of specified indices
// in the AppList
func (al AppList) Swap(i, j int) {
	al[i], al[j] = al[j], al[i]
}

// Less (Sort interface) compares the contents of the specified
// indices in the AppList and returns true if the condition holds, and
// else false.
func (al AppList) Less(i, j int) bool {
	return (al[i].Meta.Namespace < al[j].Meta.Namespace) ||
		((al[i].Meta.Namespace == al[j].Meta.Namespace) &&
			(al[i].Meta.Name < al[j].Meta.Name))
}

// AppRef references an App by name and namespace
type AppRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// NewAppRef returns a new reference to an app
func NewAppRef(name string, namespace string) AppRef {
	return AppRef{Name: name, Namespace: namespace}
}

// App returns a fresh app model for the reference
func (ar *AppRef) App() *App {
	return NewApp(ar.Name, ar.Namespace)
}

// MakeEnvSecretName returns the name of the kube secret holding the
// environment variables of the referenced application
func (ar *AppRef) MakeEnvSecretName() string {
	// TODO: This needs tests for env operations on an app with a long name
	return names.GenerateResourceName(ar.Name + "-env")
}

// MakeServiceSecretName returns the name of the kube secret holding the
// bound services of the referenced application
func (ar *AppRef) MakeServiceSecretName() string {
	// TODO: This needs tests for service operations on an app with a long name
	return names.GenerateResourceName(ar.Name + "-svc")
}

// MakeScaleSecretName returns the name of the kube secret holding the number
// of desired instances for referenced application
func (ar *AppRef) MakeScaleSecretName() string {
	// TODO: This needs tests for service operations on an app with a long name
	return names.GenerateResourceName(ar.Name + "-scale")
}

// MakePVCName returns the name of the kube pvc to use with/for the referenced application.
func (ar *AppRef) MakePVCName() string {
	return names.GenerateResourceName(ar.Namespace, ar.Name)
}

// StageRef references a staging run by ID, currently randomly generated
// for each POST to the staging endpoint
type StageRef struct {
	ID string `json:"id,omitempty"`
}

// NewStage returns a new reference to a staging run
func NewStage(id string) StageRef {
	return StageRef{id}
}

// ImageRef references an upload
type ImageRef struct {
	ID string `json:"id,omitempty"`
}

// NewImage returns a new image ref for the given ID
func NewImage(id string) ImageRef {
	return ImageRef{id}
}
