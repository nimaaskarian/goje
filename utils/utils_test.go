package utils

import (
	"os"
	"testing"
)

type ExpandTest struct {
	not_expanded string;
	expanded string;
}

func TestExpandUser(t *testing.T) {
	expanduser, err := NewExpandUser()
	if err != nil {
		t.Logf("getting user's homedir failed. ignoring expanduser test")
		return
	}
	home,_ := os.UserHomeDir()
	if expanduser.home != home {
		t.Fatalf("expanduser.home (%v) not equal home (%v)", expanduser.home, home)
	}
	test_items := []ExpandTest{
		{
			"~",
			expanduser.home,
		},
		{
			"~/",
			expanduser.home+"/",
		},
		{
			"~/Documents/",
			expanduser.home+"/Documents/",
		},
		{
			"~~/Documents/",
			"~~/Documents/",
		},
		{
			"~/~/Documents/",
			expanduser.home+"/~/Documents/",
		},
		{
			"~/.config/goje/config.toml",
			expanduser.home+"/.config/goje/config.toml",
		},
		{
			"/.config/~goje/config.toml",
			"/.config/~goje/config.toml",
		},
	}
	
	for _,item := range test_items {
		if path:=expanduser.Expand(item.not_expanded); path != item.expanded {
			t.Fatalf("Expanding %q to %q failed. path=%v err=%v", item.not_expanded, item.expanded, path, err)
		}
	}
}
