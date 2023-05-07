/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	projv1 "github.com/zachfi/proj/api/v1"
)

// CodeProjectReconciler reconciles a CodeProject object
type CodeProjectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Tracer trace.Tracer
}

//+kubebuilder:rbac:groups=proj.zachfi,resources=codeprojects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=proj.zachfi,resources=codeprojects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=proj.zachfi,resources=codeprojects/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CodeProject object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CodeProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	log := log.FromContext(ctx)

	attributes := []attribute.KeyValue{
		attribute.String("req", req.String()),
		attribute.String("namespace", req.Namespace),
	}

	ctx, span := r.Tracer.Start(ctx, "Reconcile", trace.WithAttributes(attributes...))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "complete")
		}
		span.End()
	}()

	var project projv1.CodeProject
	if err = r.Get(ctx, req.NamespacedName, &project); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dir, err := os.MkdirTemp("/tmp", fmt.Sprintf("proj_maintenance_%s_%s", project.Spec.Owner, project.Spec.Repo))
	githubURL := fmt.Sprintf("https://github.com/%s/%s.git", project.Spec.Owner, project.Spec.Repo)
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:   githubURL,
		Depth: 1,
	})

	head, err := repo.Head()
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to read HEAD")
	}

	var status projv1.CodeProjectStatus
	status.Hash = head.Hash().String()
	project.Status = status
	log.Info("updating project status", "project", req.Name, "status", fmt.Sprintf("%+v", project.Status))
	if err := r.Status().Update(ctx, &project); err != nil {
		log.Error(err, "unable to update ManagedNode status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CodeProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&projv1.CodeProject{}).
		Complete(r)
}

func (r *CodeProjectReconciler) clone() error {
	var err error

	// if i.dir == "" {
	// 	var dir string
	// 	dir, err = os.MkdirTemp("/tmp", fmt.Sprintf("keeper_%s_%s", i.Username, i.Name))
	// 	if err != nil {
	// 		return fmt.Errorf("failed to create temp dir: %w", err)
	// 	}
	// 	i.dir = dir
	// }
	//
	// githubURL := fmt.Sprintf("https://github.com/%s/%s.git", i.Username, i.Name)
	//
	// // Clones the repository into the given dir, just as a normal git clone does
	// i.gitRepo, err = git.PlainClone(i.dir, false, &git.CloneOptions{
	// 	URL:   githubURL,
	// 	Depth: 10,
	// })

	return err

	// if err != nil {
	// 	i.SetStatus(Failed)
	// 	return fmt.Errorf("clone failed: %w", err)
	// }
	// i.SetStatus(Done)

	// // Prints the content of the CHANGELOG file from the cloned repository
	// changelog, err := os.Open(filepath.Join(dir, "CHANGELOG"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// io.Copy(os.Stdout, changelog)
	// // Output: Initial changelog
}
