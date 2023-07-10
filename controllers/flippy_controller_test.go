package controllers

import (
	"gotest.tools/v3/assert"
	"os"
	"testing"
	"time"
)

//
//import (
//	"context"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/types"
//	"reflect"
//	ctrl "sigs.k8s.io/controller-runtime"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//	"testing"
//)
//
//func TestFlippyReconciler_Reconcile(t *testing.T) {
//	type fields struct {
//		Client client.Client
//		Scheme *runtime.Scheme
//	}
//	type args struct {
//		ctx context.Context
//		req ctrl.Request
//	}
//
//	testNamespace := types.NamespacedName{
//		Namespace: HappyPreCondition,
//		Name:      HappyPreCondition,
//	}
//
//	//testData := reconcile.Request{testNamespace}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    ctrl.Result
//		wantErr bool
//	}{
//		{"HappyCase", fields{
//			Client: nil,
//			Scheme: nil,
//		}, args{
//			ctx: context.TODO(),
//			req: ctrl.Request{testNamespace},
//		}, ctrl.Result{}, false},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := &FlippyReconciler{
//				Client: tt.fields.Client,
//				Scheme: tt.fields.Scheme,
//			}
//			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("Reconcile() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func TestReturnControllerWithRequeueUnset(t *testing.T) {
	result, _ := ReturnControllerWithRequeue()
	assert.Equal(t, result.RequeueAfter, time.Duration(0)*time.Second)
}

func TestReturnControllerWithRequeueSetEmpty(t *testing.T) {
	os.Setenv("REQUEUE_AFTER_HOUR", "")

	result, _ := ReturnControllerWithRequeue()
	assert.Equal(t, result.RequeueAfter, time.Duration(0)*time.Second)

	os.Unsetenv("REQUEUE_AFTER_HOUR")
}

func TestReturnControllerWithRequeueSetGarbage(t *testing.T) {
	os.Setenv("REQUEUE_AFTER_HOUR", "Garbage")

	result, _ := ReturnControllerWithRequeue()
	assert.Equal(t, result.RequeueAfter, time.Duration(0)*time.Second)

	os.Unsetenv("REQUEUE_AFTER_HOUR")
}

func TestReturnControllerWithRequeueSetZero(t *testing.T) {
	os.Setenv("REQUEUE_AFTER_HOUR", "0")

	result, _ := ReturnControllerWithRequeue()
	assert.Equal(t, result.RequeueAfter, time.Duration(0)*time.Second)

	os.Unsetenv("REQUEUE_AFTER_HOUR")
}

func TestReturnControllerWithRequeueSetNonZeror(t *testing.T) {
	os.Setenv("REQUEUE_AFTER_HOUR", "8760")

	result, _ := ReturnControllerWithRequeue()
	assert.Equal(t, result.RequeueAfter, time.Duration(8760)*time.Hour)

	os.Unsetenv("REQUEUE_AFTER_HOUR")
}
