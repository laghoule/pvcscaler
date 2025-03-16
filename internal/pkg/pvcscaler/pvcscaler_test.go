package pvcscaler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"laghoule/pvcscaler/internal/pkg/k8s"

	assert "github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	fake "k8s.io/client-go/kubernetes/fake"
)

const (
	namespace = "default"
)

func int32PTR(i int32) *int32 { return &i }

func NewFakeClient() kubernetes.Interface {
	return fake.NewSimpleClientset()
}

func createNamespace(c kubernetes.Interface) *corev1.Namespace {
	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
	}
	ns, _ := c.CoreV1().Namespaces().Create(context.Background(), nsObj, metav1.CreateOptions{})
	return ns
}

func createDeployment(c kubernetes.Interface) *appsv1.Deployment {
	depObj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "nginx-pvc",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "nginx-pvc",
								},
							},
						},
					},
				},
			},
			Replicas: int32PTR(1),
		},
	}
	dep, _ := c.AppsV1().Deployments(namespace).Create(context.Background(), depObj, metav1.CreateOptions{})
	return dep
}

func createStatefulSet(c kubernetes.Interface) *appsv1.StatefulSet {
	scName := "standard"
	stsObj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-statefulset",
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32PTR(1),
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "nginx-pvc",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName:       "nginx-pvc",
						StorageClassName: &scName,
					},
				},
			},
		},
	}
	sts, _ := c.AppsV1().StatefulSets(namespace).Create(context.Background(), stsObj, metav1.CreateOptions{})
	return sts
}

func createPVC(c kubernetes.Interface, name, namespace, storageClass string) *corev1.PersistentVolumeClaim {
	pvcObj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
		},
	}
	pvc, _ := c.CoreV1().PersistentVolumeClaims(namespace).Create(context.Background(), pvcObj, metav1.CreateOptions{})
	return pvc
}

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		kubeconfig   string
		namespace    []string
		storageClass string
		dryRun       bool
		error        bool
	}{
		{
			name:         "New PVCscaler",
			kubeconfig:   "testdata/kubeconfig.yaml",
			namespace:    []string{"default"},
			storageClass: "standard",
			dryRun:       false,
			error:        false,
		},

		{
			name:         "New PVCscaler with empty storageClass",
			kubeconfig:   "filenotfound",
			namespace:    []string{"default"},
			storageClass: "",
			dryRun:       false,
			error:        true,
		},
	}

	for _, tt := range tests {
		pvcscaler, err := New(tt.kubeconfig, tt.namespace, tt.storageClass, tt.dryRun)
		if tt.error {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, pvcscaler)
		}
	}
}

func TestGetWorkloads(t *testing.T) {
	type k8sResource[T any] struct {
		resource T
	}

	fakeClient := NewFakeClient()
	createNamespace(fakeClient)
	createPVC(fakeClient, "nginx-pvc", namespace, "standard")

	tests := []struct {
		name         string
		resource     k8sResource[any]
		namespace    []string
		storageClass string
		error        bool
	}{
		{
			name:         "GetWorkloads deployment for all namespaces",
			resource:     k8sResource[any]{resource: createDeployment(fakeClient)},
			namespace:    []string{"all"},
			storageClass: "standard",
			error:        false,
		},
		{
			name:         "GetWorkloads statefulset for all namespaces",
			resource:     k8sResource[any]{resource: createStatefulSet(fakeClient)},
			namespace:    []string{"all"},
			storageClass: "standard",
			error:        false,
		},
	}

	for _, tt := range tests {
		pvcscaler := &PVCscaler{
			k8sClient:    &k8s.Client{ClientSet: fakeClient},
			namespaces:   tt.namespace,
			storageClass: tt.storageClass,
		}

		t.Run(tt.name, func(t *testing.T) {
			err := pvcscaler.getWorkloads(context.TODO(), tt.namespace, tt.storageClass)

			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, pvcscaler.workloads, 2)
			}
		})
	}
}

func TestDown(t *testing.T) {
	type k8sResource[T any] struct {
		resource T
	}

	fakeClient := NewFakeClient()
	createNamespace(fakeClient)
	createPVC(fakeClient, "nginx-pvc", namespace, "standard")
	createDeployment(fakeClient)

	tmpDir := t.TempDir()

	tests := []struct {
		name               string
		resource           k8sResource[any]
		namespace          []string
		storageClass       string
		expectedOutputFile string
		actualOutputFile   string
		error              bool
	}{
		{
			name:               "Down deployment for all namespaces",
			resource:           k8sResource[any]{resource: createDeployment(fakeClient)},
			namespace:          []string{"all"},
			storageClass:       "standard",
			expectedOutputFile: "testdata/pvcscaler.json",
			actualOutputFile:   filepath.Join(tmpDir, "pvscaler.json"),
			error:              false,
		},
	}

	for _, tt := range tests {
		pvcscaler := &PVCscaler{
			k8sClient:    &k8s.Client{ClientSet: fakeClient},
			namespaces:   tt.namespace,
			storageClass: tt.storageClass,
		}

		t.Run(tt.name, func(t *testing.T) {
			err := pvcscaler.Down(t.Context(), tt.actualOutputFile)

			if tt.error {
				assert.Error(t, err)
			} else {
				expected, err := os.ReadFile(tt.expectedOutputFile)
				assert.NoError(t, err)

				actual, err := os.ReadFile(tt.actualOutputFile)
				assert.NoError(t, err)
				assert.JSONEq(t, string(expected), string(actual))
			}
		})
	}
}

func TestUp(t *testing.T) {
	type k8sResource[T any] struct {
		resource T
	}

	fakeClient := NewFakeClient()
	createNamespace(fakeClient)

	tests := []struct {
		name      string
		inputFile string
		resource  k8sResource[any]
		error     bool
	}{
		{
			name:      "Up deployment",
			inputFile: "testdata/pvcscaler.json",
			resource:  k8sResource[any]{resource: createDeployment(fakeClient)},
			error:     false,
		},
		{
			name:      "No deployment found",
			inputFile: "testdata/noWorkloadFoundjson",
			resource:  k8sResource[any]{resource: nil},
			error:     true,
		},
		{
			name:      "No input file found",
			inputFile: "testdata/notfound.json",
			error:     true,
		},
	}

	for _, tt := range tests {
		pvcscaler := &PVCscaler{
			k8sClient: &k8s.Client{ClientSet: fakeClient},
		}

		t.Run(tt.name, func(t *testing.T) {
			err := pvcscaler.Up(t.Context(), tt.inputFile)

			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
