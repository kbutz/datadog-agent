// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build orchestrator

package k8s

import (
	"sort"
	"testing"
	"time"

	model "github.com/DataDog/agent-payload/v5/process"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/orchestrator/processors"

	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestExtractRole(t *testing.T) {
	creationTime := metav1.NewTime(time.Date(2021, time.April, 16, 14, 30, 0, 0, time.UTC))

	tests := map[string]struct {
		input             rbacv1.Role
		labelsAsTags      map[string]string
		annotationsAsTags map[string]string
		expected          model.Role
	}{
		"standard": {
			input: rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"annotation": "my-annotation",
					},
					CreationTimestamp: creationTime,
					Labels: map[string]string{
						"app": "my-app",
					},
					Name:            "role",
					Namespace:       "namespace",
					ResourceVersion: "1234",
					UID:             types.UID("e42e5adc-0749-11e8-a2b8-000c29dea4f6"),
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{""},
						Resources: []string{"nodes", "pods", "services"},
						Verbs:     []string{"get", "patch", "list"},
					},
					{
						APIGroups: []string{"batch"},
						Resources: []string{"cronjobs", "jobs"},
						Verbs:     []string{"get", "list", "watch"},
					},
					{
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"rolebindings"},
						Verbs:     []string{"create"},
					},
				},
			},
			labelsAsTags: map[string]string{
				"app": "application",
			},
			annotationsAsTags: map[string]string{
				"annotation": "annotation_key",
			},
			expected: model.Role{
				Metadata: &model.Metadata{
					Annotations:       []string{"annotation:my-annotation"},
					CreationTimestamp: creationTime.Unix(),
					Labels:            []string{"app:my-app"},
					Name:              "role",
					Namespace:         "namespace",
					ResourceVersion:   "1234",
					Uid:               "e42e5adc-0749-11e8-a2b8-000c29dea4f6",
				},
				Rules: []*model.PolicyRule{
					{
						ApiGroups: []string{""},
						Resources: []string{"nodes", "pods", "services"},
						Verbs:     []string{"get", "patch", "list"},
					},
					{
						ApiGroups: []string{"batch"},
						Resources: []string{"cronjobs", "jobs"},
						Verbs:     []string{"get", "list", "watch"},
					},
					{
						ApiGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"rolebindings"},
						Verbs:     []string{"create"},
					},
				},
				Tags: []string{
					"application:my-app",
					"annotation_key:my-annotation",
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			pctx := &processors.K8sProcessorContext{
				LabelsAsTags:      tc.labelsAsTags,
				AnnotationsAsTags: tc.annotationsAsTags,
			}
			actual := ExtractRole(pctx, &tc.input)
			sort.Strings(actual.Tags)
			sort.Strings(tc.expected.Tags)
			assert.Equal(t, &tc.expected, actual)
		})
	}
}
