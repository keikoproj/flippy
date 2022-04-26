package k8s

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestCreateJob(t *testing.T) {

	jobToCreate := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "CreateTestJobName",
			Namespace:   "TestNamespace",
			Annotations: map[string]string{},
		},
	}

	k8 := K8sWrapper{}

	err := k8.CreateJob(fake.NewSimpleClientset(), jobToCreate, "TestNamespace")
	if err != nil {
		t.Errorf("CreateJob test failed, expected no error. Actual error = %v", err)
		return
	}

	err = k8.CreateJob(fake.NewSimpleClientset(), jobToCreate, "TestNamespace1")
	if err == nil {
		t.Errorf("Negative CreateJob test failed. Expected error.")
		return
	}
}

func TestDeleteJob(t *testing.T) {

	k8 := K8sWrapper{}

	fakeClientSet := fake.NewSimpleClientset(
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "TestJob",
				Namespace:   "TestNamespace",
				Annotations: map[string]string{},
			},
		}, &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "TestJobWithPod",
				Namespace:   "TestNamespace",
				Annotations: map[string]string{},
			},
		})

	err := k8.DeleteJob(fakeClientSet, "TestNamespace", "TestJob")
	if err != nil {
		t.Errorf("DeleteJob test failed, expected no error. Actual error = %v", err)
		return
	}

	err = k8.DeleteJobWithPods(fakeClientSet, "TestNamespace", "TestJobWithPod")
	if err != nil {
		t.Errorf("DeleteJobWithPods test failed, expected no error. Actual error = %v", err)
		return
	}

	err = k8.DeleteJobWithPods(fakeClientSet, "TestNamespace", "TestNonExistingJob")
	if err == nil {
		t.Errorf("DeleteJob test failed for non existing job, expected error.")
		return
	}
}
