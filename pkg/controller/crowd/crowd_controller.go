package crowd

import (
	"context"
	"k8s.io/apimachinery/pkg/util/intstr"

	appv1alpha1 "github.com/example-inc/app-operator/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Constants for hello-stateful StatefulSet & Volumes
const (
	DiskSize            = 1 * 1000 * 1000 * 1000
	AppVolumeName       = "app"
	AppVolumeMountPath  = "/usr/share/hello"
	HostProvisionerPath = "/tmp/hostpath-provisioner"
	AppImage            = "quay.io/shall3790/crowd:3.7.0"
	AppContainerName    = "crowd"
	ImagePullPolicy     = corev1.PullAlways
)

var log = logf.Log.WithName("controller_crowd")
var Replicas int32 = 1

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Crowd Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCrowd{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("crowd-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Crowd
	err = c.Watch(&source.Kind{Type: &appv1alpha1.Crowd{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Crowd
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.Crowd{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCrowd implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCrowd{}

// ReconcileCrowd reconciles a Crowd object
type ReconcileCrowd struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Crowd object and makes changes based on the state read
// and what is in the Crowd.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCrowd) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Crowd")

	// Fetch the Crowd instance
	instance := &appv1alpha1.Crowd{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new StatefulSet object
	stateful_set := newStatefulSetForCr(instance)

	// new Service
	service := newService(instance)

	// Set Crowd instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, stateful_set, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this Pod already exists
	//found := &corev1.Pod{}
	found := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: stateful_set.Name, Namespace: stateful_set.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new stateful_set", "StatefulSet.Namespace", stateful_set.Namespace, "StatfulSet.Name", stateful_set.Name)
		err = r.client.Create(context.TODO(), stateful_set)
		if err != nil {
			return reconcile.Result{}, err
		}

		// stateful_set created successfully - don't requeue
		// return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)

	// Check if this service already exists
	foundSvc := &corev1.Service{}
	errSvc := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundSvc)
	if errSvc != nil && errors.IsNotFound(errSvc) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		errSvc = r.client.Create(context.TODO(), service)
		if errSvc != nil {
			return reconcile.Result{}, errSvc
		}

		// Service created successfully - don't requeue
		//return reconcile.Result{}, nil
	} else if errSvc != nil {
		return reconcile.Result{}, errSvc
	} else {
		// Service already exists
		reqLogger.Info("Service already exists", "Service.Namespace", foundSvc.Namespace, "Service.Name", foundSvc.Name)
	}

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *appv1alpha1.Crowd) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

func newStatefulSetForCr(cr *appv1alpha1.Crowd) *appsv1.StatefulSet {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector:    labelSelector(labels),
			ServiceName: cr.Name,
			Replicas:    &cr.Spec.Size,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  cr.Name,
							Image: AppImage,
							Ports: []corev1.ContainerPort{{
								ContainerPort: 8095,
								Name:          "standardhttp",
							}},
						},
					},
				},
			},
		},
	}
}

func newService(cr *appv1alpha1.Crowd) *corev1.Service {
	labels := labelsForHelloStateful(cr.Name)
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Name:     "standardhttp",
				Protocol: "TCP",
				Port:     8095,
				TargetPort: intstr.IntOrString{
					Type:   0,
					IntVal: 0,
					StrVal: "8095",
				},
			}},
		},
	}
	return service
}
func labelsForHelloStateful(name string) map[string]string {
	return map[string]string{"app": name}
}

func labelSelector(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{MatchLabels: labels}
}
