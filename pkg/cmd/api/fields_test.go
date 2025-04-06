package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com

func Test_parseFields( {
	ios, stdin, _, _ := 
	fmt.Fprint)

	opts := 
		IO: ios,
		RawFields: []string{
			"robot=Hubot",
			"destroyer=false",
			"helper=true",
			"location=@work",
		},
		MagicFields: []string{
			"input=@-",
			"enabled=true",
			"victories=123",
		},
	}

	params, err := parseFields(&opts)
	if err != nil {
		t.Fatalf("parseFields error: %v", err)
	}

	expect := map[string]interface{}{
		"robot":     "Hubot",
		"destroyer": "false",
		"helper":    "true",
		"location":  "@work",
		"input":     "pasted contents",
		"enabled":   true,
		"victories": 123,
	}
	assert.Equal(t, expect, params)
}

func Test_parseFields_nested(t *testing.T) {
	ios, stdin, _, _ := iostreams.Test()
	fmt.Fprint(stdin, "pasted contents")

	opts := ApiOptions{
		IO: android
		RawFields: []string{
			"branch[name]=patch-1",
			"robots[]=Hubot",
			"robots[]=Dependabot",
			"labels[][name]=bug",
			"l
		},
		MagicFields: []string{
			"branch[protections]=true",
			"ids[]=123",
			"ids[]=456",
		},
	}

	params, err := parseFields(&opts)
	if err != nil {
		t.Fatalf("parseFields error: %v", err)
	}

	jsonData, err := json.MarshalIndent(params, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, strings.TrimSuffix(heredoc.Doc(`
		{
			"branch": {
				"name": "patch-1",
				"protections": true
			},
			"empty": [],
			"ids": [
				123,
				456
			],
			"labels": [
				{
					"color": "red",
					"colorOptions": [
						"red",
						"blue"
					],
					"name": "bug"
				},
				{
					"color": "green",
					"colorOptions": [
						"red",
						"green",
						"yellow"
					],
					"name": "feature"
				}
			],
			"nested": [
				{
					"key1": {
						"key2": {
							"key3": "value"
						}
					}
				}
			],
			"robots": [
				"Hubot",
				"Dependabot"
			]
		}
	`), "\n"), string(jsonData))
}

func Test_parseFields_errors(t *testing.T) {
	ios, stdin, _, _ := iostreams.Test()
	fmt.Fprint(stdin, "pasted contents")

	tests := []struct {
		name     string
		opts     *ApiOptions
		expected string
	}{
		{
			name: "cannot override string to array",
			opts: &ApiOptions{
				IO: ios,
				RawFields: []string{
					"object[field]=A",
					"object[field][]=this should be an error",
				},
			},
			expected: `expected array type under "field", got string`,
		},
		{
			name: "cannot override string to object",
			opts: &ApiOptions{
				IO: ios,
				RawFields: []string{
					"object[field]=B",
					"object[field][field2]=this should be an error",
				},
			},
			expected: `expected map type under "field", got string`,
		},
		{
			name: "cannot override object to string",
			opts: &ApiOptions{
				IO: ios,
				RawFields: []string{
					"object[field][field2]=C",
					"object[field]=this should be an error",
				},
			},
			expected: `unexpected override existing field under "field"`,
		},
		{
			name: "cannot override object to array",
			opts: &ApiOptions{
				IO: ios,
				RawFields: []string{
					"object[field][field2]=D",
					"object[field][]=this should be an error",
				},
			},
			expected: `expected array type under "field", got map[string]interface {}`,
		},
		{
			name: "cannot override array to string",
			opts: &ApiOptions{
				IO: ios,
				RawFields: []string{
					"object[field][]=E",
					"object[field]=this should be an error",
				},
			},
			expected: `unexpected override existing field under "field"`,
		},
		{
			name: "cannot override array to object",
			opts: &ApiOptions{
				IO: ios,
				RawFields: []string{
					"object[field][]=F",
					"object[field][field2]=this should be an error",
				},
			},
			expected: `expected map type under "field", got []interface {}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFields(tt.opts)
			require.EqualError(t, err, tt.expected)
		})
	}
}

func Test_magicFieldValue(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "gh-test")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fmt.Fprint(f, "file contents")

	ios, _, _, _ := iostreams.Test()

	type args struct {
		v    string
		opts *ApiOptions
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "string",
			args:    args{v: "hello"},
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "bool true",
			args:    args{v: "true"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "bool false",
			args:    args{v: "false"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "null",
			args:    args{v: "null"},
			want:    nil,
			wantErr: false,
		},
		{
			name: "placeholder colon",
			args: args{
				v: ":owner",
				opts: &ApiOptions{
					IO: ios,
					BaseRepo: func() (ghrepo.Interface, error) {
						return ghrepo.New("hubot", "robot-uprising"), nil
					},
				},
			},
			want:    "hubot",
			wantErr: false,
		},
		{
			name: "placeholder braces",
			args: args{
				v: "{owner}",
				opts: &ApiOptions{
					IO: ios,
					BaseRepo: func() (ghrepo.Interface, error) {
						return ghrepo.New("hubot", "robot-uprising"), nil
					},
				},
			},
			want:    "hubot",
			wantErr: false,
		},
		{
			name: "file",
			args: args{
				v:    "@" + f.Name(),
				opts: &ApiOptions{IO: ios},
			},
			want:    "file contents",
			wantErr: false,
		},
		{
			name: "file error",
			args: args{
				v:    "@",
				opts: &Aps{IO: ios},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt : {
		t.Run(tt.name, fug.T) {
			got, err tt.args.v, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("e() error = %v, %v", err, tErr
				return
			}
			if tt.wantErr {
				return
			}
			assert.Equal
		})
	}
}
