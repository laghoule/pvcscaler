package test

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	fake "k8s.io/client-go/kubernetes/fake"
)

const (
	Namespace        = "default"
	StorageClass     = "standard"
	DeploymentName   = "nginx-deployment"
	StatefulSetName  = "nginx-statefulset"
	DaemonSetName    = "nginx-daemonset"
	PodName          = "nginx"
	PodImage         = "nginx"
	PvcName          = "nginx-pvc"
	ClaimName        = "nginx-pvc"
	StorageClassName = "standard"
)

// K8sResource wraps a Kubernetes resource for table-driven tests.
type K8sResource[T any] struct {
	Resource T
}

func int32PTR(i int32) *int32 { return &i }

// NewFakeClient returns an in-memory Kubernetes clientset for tests.
func NewFakeClient() kubernetes.Interface {
	return fake.NewSimpleClientset()
}

// CreateNamespace creates the shared test namespace.
func CreateNamespace(c kubernetes.Interface) *corev1.Namespace {
	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: Namespace,
		},
	}
	ns, _ := c.CoreV1().Namespaces().Create(context.TODO(), nsObj, metav1.CreateOptions{})
	return ns
}

// CreateDeployment creates a Deployment that mounts the shared PVC.
func CreateDeployment(c kubernetes.Interface) *appsv1.Deployment {
	depObj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeploymentName,
			Namespace: Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: PvcName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: ClaimName,
								},
							},
						},
					},
				},
			},
			Replicas: int32PTR(1),
		},
	}
	dep, _ := c.AppsV1().Deployments(Namespace).Create(context.TODO(), depObj, metav1.CreateOptions{})
	return dep
}

// CreateStatefulSet creates a StatefulSet with a volume claim template using
// StorageClassName.
func CreateStatefulSet(c kubernetes.Interface) *appsv1.StatefulSet {
	scName := StorageClassName
	stsObj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      StatefulSetName,
			Namespace: Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32PTR(1),
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: PvcName,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName:       ClaimName,
						StorageClassName: &scName,
					},
				},
			},
		},
	}
	sts, _ := c.AppsV1().StatefulSets(Namespace).Create(context.Background(), stsObj, metav1.CreateOptions{})
	return sts
}

// CreateDaemonSet creates a DaemonSet without persistent volumes.
func CreateDaemonSet(c kubernetes.Interface) *appsv1.DaemonSet {
	dsObj := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DaemonSetName,
			Namespace: Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  PodName,
							Image: PodImage,
						},
					},
				},
			},
		},
	}
	ds, _ := c.AppsV1().DaemonSets(Namespace).Create(context.TODO(), dsObj, metav1.CreateOptions{})
	return ds
}

// CreatePVC creates a PVC with the given name, namespace and storage class.
func CreatePVC(c kubernetes.Interface, name, namespace, storageClass string) *corev1.PersistentVolumeClaim {
	pvcObj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
		},
	}
	pvc, _ := c.CoreV1().PersistentVolumeClaims(namespace).Create(context.TODO(), pvcObj, metav1.CreateOptions{})
	return pvc
}

// CreateNamespaceNamed creates a namespace with the given name.
func CreateNamespaceNamed(c kubernetes.Interface, name string) *corev1.Namespace {
	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	ns, _ := c.CoreV1().Namespaces().Create(context.TODO(), nsObj, metav1.CreateOptions{})
	return ns
}

// CreateStorageClass creates a StorageClass, optionally marked as the cluster
// default.
func CreateStorageClass(c kubernetes.Interface, name string, defaultClass bool) *storagev1.StorageClass {
	annotations := map[string]string{}
	if defaultClass {
		annotations["storageclass.kubernetes.io/is-default-class"] = "true"
	}

	scObj := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
		},
		Provisioner: "kubernetes.io/no-provisioner",
	}
	sc, _ := c.StorageV1().StorageClasses().Create(context.TODO(), scObj, metav1.CreateOptions{})
	return sc
}

// CreateDeploymentNamed creates a Deployment named name that mounts the given
// PVC claim names.
func CreateDeploymentNamed(c kubernetes.Interface, name, namespace string, claimNames []string) *appsv1.Deployment {
	volumes := make([]corev1.Volume, 0, len(claimNames))
	for i, claimName := range claimNames {
		volumes = append(volumes, corev1.Volume{
			Name: PvcName + "-" + string(rune('a'+i)),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claimName,
				},
			},
		})
	}

	depObj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32PTR(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: volumes,
				},
			},
		},
	}
	dep, _ := c.AppsV1().Deployments(namespace).Create(context.TODO(), depObj, metav1.CreateOptions{})
	return dep
}
