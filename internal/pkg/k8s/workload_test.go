package k8s

import (
	"testing"

	"github.com/laghoule/pvcscaler/internal/pkg/test"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func createDeploymentWorkload(c kubernetes.Interface) *Workload {
	test.CreateDeployment(c)
	return &Workload{
		Name:      "nginx-deployment",
		Namespace: namespace,
		Kind:      "Deployment",
		Replicas:  1,
	}
}

func createStatefulsetWorkload(c kubernetes.Interface) *Workload {
	test.CreateStatefulSet(c)
	return &Workload{
		Name:      "nginx-statefulset",
		Namespace: namespace,
		Kind:      "StatefulSet",
		Replicas:  1,
	}
}

func TestGetWorkloads(t *testing.T) {
	c, err := NewTestClient(t)
	if err != nil {
		t.Error(err)
	}

	test.CreateDeployment(c.ClientSet)
	createPVC(c.ClientSet, "nginx-pvc", namespace, "standard")

	wloads, err := c.GetWorkloads(namespace, "standard")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(wloads))
	assert.Equal(t, "nginx-deployment", wloads[0].Name)
	assert.Equal(t, namespace, wloads[0].Namespace)
}

func TestGetDeploymentWorkloads(t *testing.T) {
	c, err := NewTestClient(t)
	if err != nil {
		t.Error(err)
	}

	createPVC(c.ClientSet, "nginx-pvc", namespace, "standard")

	tests := []struct {
		name         string
		dep          *appsv1.Deployment
		storageClass string
		expected     []Workload
		error        bool
	}{
		{
			name:         "deployment with pvc",
			dep:          test.CreateDeployment(c.ClientSet),
			storageClass: "standard",
			expected: []Workload{
				{
					Name:      "nginx-deployment",
					Namespace: namespace,
					Kind:      "Deployment",
					Replicas:  1,
				},
			},
			error: false,
		},
		{
			name:         "deployment with pvc and wrong storage class",
			dep:          test.CreateDeployment(c.ClientSet),
			storageClass: "not found",
			expected:     []Workload{},
			error:        false,
		},
		{
			name: "deployment with no pvc",
			dep: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{},
				},
			},
			expected: []Workload{},
			error:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := c.GetDeploymentWorkloads(tt.dep.Namespace, tt.storageClass)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestGetStatefulSetWorkloads(t *testing.T) {
	c, err := NewTestClient(t)
	if err != nil {
		t.Error(err)
	}

	test.CreateStatefulSet(c.ClientSet)
	createPVC(c.ClientSet, "nginx-pvc", namespace, "standard")

	wloads, err := c.GetStatefulSetWorkloads(namespace, "standard")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(wloads))
	assert.Equal(t, "nginx-statefulset", wloads[0].Name)
	assert.Equal(t, namespace, wloads[0].Namespace)
}

func TestGetDeploymentWorkloadsNoDuplicate(t *testing.T) {
	c, err := NewTestClient(t)
	assert.NoError(t, err)

	test.CreatePVC(c.ClientSet, "pvc-a", namespace, "standard")
	test.CreatePVC(c.ClientSet, "pvc-b", namespace, "standard")
	test.CreateDeploymentNamed(c.ClientSet, "nginx-deployment", namespace, []string{"pvc-a", "pvc-b"})

	wloads, err := c.GetDeploymentWorkloads(namespace, "standard")
	assert.NoError(t, err)
	assert.Len(t, wloads, 1)
}

func TestGetStatefulSetWorkloadsDefaultStorageClass(t *testing.T) {
	c, err := NewTestClient(t)
	assert.NoError(t, err)

	test.CreateStorageClass(c.ClientSet, "standard", true)

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "sts-default-sc", Namespace: namespace},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32PTR(2),
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "data"},
					Spec:       corev1.PersistentVolumeClaimSpec{},
				},
			},
		},
	}
	_, err = c.ClientSet.AppsV1().StatefulSets(namespace).Create(c.Context, sts, metav1.CreateOptions{})
	assert.NoError(t, err)

	wloads, err := c.GetStatefulSetWorkloads(namespace, "standard")
	assert.NoError(t, err)
	assert.Len(t, wloads, 1)
	assert.Equal(t, int32(2), wloads[0].Replicas)
}

func TestGetStatefulSetWorkloadsWithExistingPVCVolume(t *testing.T) {
	c, err := NewTestClient(t)
	assert.NoError(t, err)

	createPVC(c.ClientSet, "shared-pvc", namespace, "standard")

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "sts-existing-pvc", Namespace: namespace},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32PTR(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "shared-pvc",
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = c.ClientSet.AppsV1().StatefulSets(namespace).Create(c.Context, sts, metav1.CreateOptions{})
	assert.NoError(t, err)

	wloads, err := c.GetStatefulSetWorkloads(namespace, "standard")
	assert.NoError(t, err)
	assert.Len(t, wloads, 1)
	assert.Equal(t, "sts-existing-pvc", wloads[0].Name)
}
