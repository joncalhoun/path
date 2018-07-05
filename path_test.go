package path

import (
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBuilder_Set(t *testing.T) {
	var pb Builder
	var names []string
	seed := time.Now().UnixNano()
	t.Logf("Seed: %v", seed)
	rand := rand.New(rand.NewSource(seed))
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("%v", rand.Int())
		names = append(names, name)
		pb.Set(name, "/some-path")
		for _, testName := range names {
			if _, err := pb.StrictPath(testName, nil); err != nil {
				t.Errorf("Builder.Set(%v) didn't persist", testName)
			}
		}
	}
}

func TestBuilder_Path(t *testing.T) {
	var pb Builder
	pb.Set("show_dog", "/dogs/:id")
	tests := []struct {
		name, path, want string
		params           map[string]interface{}
	}{
		{"errors returns empty string", "fake_path", "", nil},
		{"existing paths work", "show_dog", "/dogs/123", map[string]interface{}{"id": 123}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := pb.Path(tc.path, tc.params)
			if got != tc.want {
				t.Errorf("Builder.Path(%v) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

// Like 90% of this functionality is tested in Test_replace, so all
// I really want to test is that we pass the correct argus into replace.
func TestBuilder_StrictPath(t *testing.T) {
	var pb Builder
	pb.Set("create_dog", "/dogs/")
	pb.Set("show_dog", "/dogs/:id")
	pb.Set("edit_dog", "/dogs/:id/edit")

	type args struct {
		name   string
		params map[string]interface{}
	}
	tests := []struct {
		name         string
		args         args
		ignoreParams bool
		want         string
		wantErr      error
	}{
		{
			name: "create_dog no params",
			args: args{name: "create_dog"},
			want: "/dogs/",
		},
		{
			name: "create_dog with params",
			args: args{
				name: "create_dog",
				params: map[string]interface{}{
					"age": 12,
				},
			},
			want:         "/dogs/?age=12",
			ignoreParams: false,
		},
		{
			name: "create_dog ignored params",
			args: args{
				name: "create_dog",
				params: map[string]interface{}{
					"age": 12,
				},
			},
			want:         "/dogs/",
			ignoreParams: true,
		},
		{
			name: "show_dog",
			args: args{
				name: "show_dog",
				params: map[string]interface{}{
					"id": 123,
				},
			},
			want: "/dogs/123",
		},
		{
			name: "error",
			args: args{
				name: "invalid_name",
			},
			want:    "",
			wantErr: ErrNotFound,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pb.IgnoreExtraParams = tc.ignoreParams
			got, err := pb.StrictPath(tc.args.name, tc.args.params)
			if err != tc.wantErr {
				t.Fatalf("Builder.StrictPath() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("Builder.StrictPath() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBuilder_init(t *testing.T) {
	var b Builder
	b.init()
	// This should not panic after we init
	b.paths["key"] = "value"
}

func Test_replace(t *testing.T) {
	type args struct {
		path   string
		params map[string]interface{}
		query  bool
	}
	tests := []struct {
		name      string
		args      args
		wantBase  string
		wantQuery url.Values
	}{
		{
			name: "no replacements",
			args: args{
				path:   "/some/path",
				params: nil,
				query:  false,
			},
			wantBase: "/some/path",
		},
		{
			name: "id replacement",
			args: args{
				path: "/widgets/:id",
				params: map[string]interface{}{
					"id": 123,
				},
				query: false,
			},
			wantBase: "/widgets/123",
		},
		{
			name: "query with no replacements",
			args: args{
				path: "/widgets/",
				params: map[string]interface{}{
					"id": 123,
				},
				query: true,
			},
			wantBase: "/widgets/",
			wantQuery: url.Values{
				"id": []string{"123"},
			},
		},
		{
			name: "query and replacements",
			args: args{
				path: "/widgets/:id/edit/:blah",
				params: map[string]interface{}{
					"id":   123,
					"blah": "dog",
					"name": "felix",
				},
				query: true,
			},
			wantBase: "/widgets/123/edit/dog",
			wantQuery: url.Values{
				"name": []string{"felix"},
			},
		},
		{
			name: "query replacements and missing param",
			args: args{
				path: "/widgets/:id/edit/:blah",
				params: map[string]interface{}{
					"id":   123,
					"name": "felix",
				},
				query: true,
			},
			wantBase: "/widgets/123/edit/:blah",
			wantQuery: url.Values{
				"name": []string{"felix"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := replace(tc.args.path, tc.args.params, tc.args.query)
			gotPieces := strings.SplitN(got, "?", 2)
			gotBase := gotPieces[0]
			if gotBase != tc.wantBase {
				t.Errorf("replace() = %v, want %v", gotBase, tc.wantBase)
			}
			if !tc.args.query {
				return
			}
			if len(gotPieces) != 2 {
				t.Fatalf("replace() query = %v, want %v", nil, tc.wantQuery)
			}
			gotQ, err := url.ParseQuery(gotPieces[1])
			if err != nil {
				t.Fatalf("url.ParseQuery(%v) err = %v, want %v", gotPieces[1], err, nil)
			}
			if !reflect.DeepEqual(gotQ, tc.wantQuery) {
				t.Errorf("replace() query = %v, want %v", gotQ, tc.wantQuery)
			}
		})
	}
}

func Test_key(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    string
		wantErr error
	}{
		{"valid key", ":id", "id", nil},
		{"invalid key", "id", "", errInvalidKey},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := key(tc.arg)
			if err != tc.wantErr {
				t.Errorf("key() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("key() = %v, want %v", got, tc.want)
			}
		})
	}
}
