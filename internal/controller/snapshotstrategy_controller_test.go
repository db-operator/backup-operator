/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kindarocksv1alpha1 "github.com/db-operator/backup-operator/api/v1alpha1"
)

var _ = Describe("SnapshotStrategy Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		snapshotstrategy := &kindarocksv1alpha1.SnapshotStrategy{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind SnapshotStrategy")
			err := k8sClient.Get(ctx, typeNamespacedName, snapshotstrategy)
			if err != nil && errors.IsNotFound(err) {
				resource := &kindarocksv1alpha1.SnapshotStrategy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: kindarocksv1alpha1.SnapshotStrategySpec{
						PostgresDumpScript: "postgres_dump.sh",
						MysqlDumpScript:    "mysql_dump.sh",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &kindarocksv1alpha1.SnapshotStrategy{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance SnapshotStrategy")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			workdir, err := os.MkdirTemp("", "snashotStrategy")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(workdir)
			controllerReconciler := &SnapshotStrategyReconciler{
				Client:        k8sClient,
				Scheme:        k8sClient.Scheme(),
				ScriptsFolder: workdir,
			}

			postgresDump := []byte("echo 1")
			err = os.WriteFile(fmt.Sprintf("%s/postgres_dump.sh", workdir), postgresDump, 0777)
			Expect(err).NotTo(HaveOccurred())

			mysqlDump := []byte("echo 1")
			err = os.WriteFile(fmt.Sprintf("%s/mysql_dump.sh", workdir), mysqlDump, 0777)
			Expect(err).NotTo(HaveOccurred())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(snapshotstrategy.Status.ScriptsVerified).To(Equal(true))

		})

		It("should fail because scripts don't exist reconcile the resource", func() {
			By("Reconciling the created resource")
			workdir, err := os.MkdirTemp("", "snashotStrategy")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(workdir)
			controllerReconciler := &SnapshotStrategyReconciler{
				Client:        k8sClient,
				Scheme:        k8sClient.Scheme(),
				ScriptsFolder: workdir,
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())

			Expect(snapshotstrategy.Status.ScriptsVerified).To(Equal(false))

		})
	})
})
