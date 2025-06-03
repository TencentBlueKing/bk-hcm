/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package jwttoken

import (
	"testing"
)

func Test_defaultParser_GenerateToken(t *testing.T) {
	type args struct {
		userName   string
		workflowID string
		title      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				userName:   "name-x",
				workflowID: "workflow-y",
				title:      "title-z",
			},
			want:    "name-x/workflow-y/title-z",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &defaultParser{}
			got, err := p.GenerateToken(tt.args.userName, tt.args.workflowID, tt.args.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultParser_ParseToken(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		want2   string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				token: "name-x/workflow-y/title-z",
			},
			want:    "name-x",
			want1:   "workflow-y",
			want2:   "title-z",
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				token: "name-x/ticket-y",
			},
			want:    "",
			want1:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &defaultParser{}
			got, got1, got2, err := p.ParseToken(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseToken() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseToken() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("ParseToken() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_jwtParser_GenerateANDParseToken(t *testing.T) {
	type fields struct {
		SecretKey []byte
	}
	type args struct {
		userName   string
		workflowID string
		title      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   string
		want2   string
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				SecretKey: []byte("ef60bae75b5a86777b0e039dad51ef1c1d969d1918977d453aa8ae150543874a"),
			},
			args: args{
				userName:   "name-x",
				workflowID: "workflow-y",
				title:      "title-z",
			},
			want:    "name-x",
			want1:   "workflow-y",
			want2:   "title-z",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &jwtParser{
				SecretKey: tt.fields.SecretKey,
			}
			token, err := p.GenerateToken(tt.args.userName, tt.args.workflowID, tt.args.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, got1, got2, err := p.ParseToken(token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseToken() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseToken() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("ParseToken() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
