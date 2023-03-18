package file

import (
	"os"
	"testing"
)

func TestEmptyFile(t *testing.T) {
	file := "/tmp/testfile"
	f, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	out := EmptyFile(file)
	if !out {
		t.Fatal("File not empty")
	}
	err = os.Remove(file)
	if err != nil {
		t.Fatal(err)
	}
}
func TestExists(t *testing.T) {
	file := "/tmp/testexists"
	err := os.WriteFile(file, []byte("hello world"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	size, out := Exists(file)
	if !out || size == 0 {
		t.Fatalf("error not corect %v %d", out, size)
	}
	err = os.Remove(file)
	if err != nil {
		t.Fatal(err)
	}
}
func TestCreateIfDoesNotExistInvalidOrEmpty(t *testing.T) {

}
