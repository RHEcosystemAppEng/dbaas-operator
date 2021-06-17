/*
Copyright 2021.

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
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
	"time"
)

func Test_decode(t *testing.T) {
	lastTransitionTimeString := "2021-06-18T20:03:20Z"
	lastTransitionTime, _ := time.Parse(time.RFC3339, lastTransitionTimeString)

	type args struct {
		input  interface{}
		output interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "inventory_status",
			args: args{
				input: map[string]interface{}{
					"type": "MongoDB",
					"conditions": []map[string]interface{}{
						{
							"lastTransitionTime": lastTransitionTimeString,
							"message":            "Secret not found",
							"reason":             "InputError",
							"status":             "False",
							"type":               "SpecSynced",
						},
					},
				},
				output: &v1alpha1.DBaaSInventoryStatus{},
			},
			want: &v1alpha1.DBaaSInventoryStatus{
				Type: "MongoDB",
				Conditions: []metav1.Condition{
					{
						LastTransitionTime: metav1.Time{Time: lastTransitionTime},
						Message:            "Secret not found",
						Reason:             "InputError",
						Status:             "False",
						Type:               "SpecSynced",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := decode(tt.args.input, tt.args.output); (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decode(tt.args.input, tt.args.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.output, tt.want) {
				t.Errorf("decode() got = %v, want %v", tt.args.output, tt.want)
			}
		})
	}
}
