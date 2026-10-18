package packagetool_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Ambloplites/packagetool"
)

//TODO: test ContentsFS

func TestPackage(t *testing.T) {
	type file struct {
		Name string `json:"name"`
		Hash string `json:"hash"`
	}
	var testcases []struct {
		file
		Encrypted bool   `json:"encrypted"`
		Files     []file `json:"files"`
	}

	data, err := os.ReadFile("testdata/packages.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &testcases); err != nil {
		t.Fatal(err)
	}

	hash := sha256.New()
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			testFile, err := os.Open("testdata/" + testcase.Name)
			if err != nil {
				t.Fatal("failed to open test package:", err)
			}
			defer testFile.Close()

			reader, err := packagetool.NewPackageReader(testFile, 0, packagetool.PackageReaderSettings{
				DecryptMAR: testcase.Encrypted,
			})
			if err != nil {
				t.Fatal("failed to create package reader:", err)
			}

			pkg, err := packagetool.NewPackage(reader)
			if err != nil {
				t.Fatal("failed to create package:", err)
			}
			if pkg.FileCount() != len(testcase.Files) {
				t.Fatal("invalid number of files:", pkg.FileCount())
			}

			// test contents
			contents, err := pkg.ContentsFS()
			if err != nil {
				t.Fatal("failed to create fs for contents: ", err)
			}
			if _, err := contents.Open("allfiles.lst"); err != nil {
				t.Fatal(err)
			}

			for _, file := range testcase.Files {
				f, err := pkg.Open(file.Name)
				if err != nil {
					t.Fatal(err)
				}

				hash.Reset()
				if _, err := io.Copy(hash, f); err != nil {
					t.Fatal("failed to read file:", err)
				}
				got := hex.EncodeToString(hash.Sum(nil))
				if !strings.EqualFold(got, file.Hash) {
					t.Fatalf("invalid hash: filename=%s, want=%s, got=%s", file.Name, file.Hash, got)
				}
			}
		})
	}
}
