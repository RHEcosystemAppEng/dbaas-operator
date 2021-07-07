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
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mitchellh/mapstructure"
)

func StringToAPITimeHookFunc(layout string) mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(metav1.Time{}) {
			return data, nil
		}

		if parsedTime, err := time.Parse(layout, data.(string)); err != nil {
			return metav1.Time{}, err
		} else {
			return metav1.Time{Time: parsedTime}, nil
		}
	}
}

func decode(input interface{}, output interface{}) error {
	config := mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
			StringToAPITimeHookFunc(time.RFC3339),
		),
		Result: output,
	}

	if decoder, err := mapstructure.NewDecoder(&config); err != nil {
		return err
	} else if err := decoder.Decode(input); err != nil {
		return err
	}
	return nil
}
