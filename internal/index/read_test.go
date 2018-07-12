package index

import (
	"io/ioutil"
	"os"
	"testing"
)

var postFiles = map[string]string{
	"file0": "",
	"file1": "Tester Code Search",
	"file2": "Tester Code Project Hosting",
	"file3": "Tester Web Search",
}

func tri(x, y, z byte) uint32 {
	return uint32(x)<<16 | uint32(y)<<8 | uint32(z)
}

func TestTrivialPosting(t *testing.T) {
	f, _ := ioutil.TempFile("", "index-test")
	defer os.Remove(f.Name())
	out := f.Name()
	buildIndex(out, nil, postFiles)
	ix := Open(out)
	if l := ix.PostingList(tri('S', 'e', 'a')); !equalList(l, []uint32{1, 3}) {
		t.Errorf("PostingList(Sea) = %v, want [1 3]", l)
	}
	if l := ix.PostingList(tri('T', 'e', 's')); !equalList(l, []uint32{1, 2, 3}) {
		t.Errorf("PostingList(Tes) = %v, want [1 2 3]", l)
	}
	if l := ix.PostingAnd(ix.PostingList(tri('S', 'e', 'a')), tri('T', 'e', 's')); !equalList(l, []uint32{1, 3}) {
		t.Errorf("PostingList(Sea&Tes) = %v, want [1 3]", l)
	}
	if l := ix.PostingAnd(ix.PostingList(tri('T', 'e', 's')), tri('S', 'e', 'a')); !equalList(l, []uint32{1, 3}) {
		t.Errorf("PostingList(Tes&Sea) = %v, want [1 3]", l)
	}
	if l := ix.PostingOr(ix.PostingList(tri('S', 'e', 'a')), tri('T', 'e', 's')); !equalList(l, []uint32{1, 2, 3}) {
		t.Errorf("PostingList(Sea|Tes) = %v, want [1 2 3]", l)
	}
	if l := ix.PostingOr(ix.PostingList(tri('T', 'e', 's')), tri('S', 'e', 'a')); !equalList(l, []uint32{1, 2, 3}) {
		t.Errorf("PostingList(Tes|Sea) = %v, want [1 2 3]", l)
	}
}

func equalList(x, y []uint32) bool {
	if len(x) != len(y) {
		return false
	}
	for i, xi := range x {
		if xi != y[i] {
			return false
		}
	}
	return true
}
