package flag

import (
	"testing"
)

func str(val *string) string {
	if val == nil {
		return "(null)"
	}
	return *val
}

func TestString(t *testing.T) {
	var fs FlagSet

	str1 := fs.String("s1", "", "usage")
	str2 := fs.String("s2", "", "usage")
	bool1 := fs.Bool("b", false, "usage")

	if err := fs.Parse([]string{"-s1", "ahaha", "-b", "ihihi", "-s2", "ufufu"}); err != nil {
		t.Fatal(err.Error())
	}
	if result, expect := str(str1), "ahaha"; result != expect {
		t.Fatalf("expect %#v, but (*FlagSet) String() returns %#v", expect, result)
	}
	if result, expect := str(str2), "ufufu"; result != expect {
		t.Fatalf("expect %#v, but (*FlagSet) String() returns %#v", expect, result)
	}
	if args := fs.Args(); len(args) != 1 || args[0] != "ihihi" {
		t.Fatalf("(*FlagSet) Args() returns %#v", args)
	}
	if bool1 == nil {
		t.Fatal("expect not nil, but (*FlagSet) Bool() returns nil")
	}
	if *bool1 != true {
		t.Fatalf("expect %#v, but (*FlagSet) Bool() returns %#v", true, *bool1)
	}
}
