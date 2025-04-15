package git  

import (
"testing"
)

func TestConvert(t *testing.T) {
	cases := []struct {
		name string
		actual	 string
		want string
		
	}{
		{
			name: "Standard git repo",
			actual: "",
			want: "",
		
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			err := createBaseGitRepo(tmpDir)
			assertNoError(t, err)


			got := <function>(c.actual)
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
