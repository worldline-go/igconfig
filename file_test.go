package igconfig

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileBadData(t *testing.T) {
	const funcName = "TestFileBadData"
	const fileName = "/tmp/TestFileBadData.cfg"
	const fileData = "age=haha"

	if err := ioutil.WriteFile(fileName, []byte(fileData), 0644); err != nil {
		t.Errorf("%s could not write temporary fileName '%s'", funcName, fileName)
		return
	}
	defer os.Remove(fileName)

	var c testConfig
	data := localData{userStruct: &c, fileName: fileName}
	data.loadDefaults()

	if data.loadFile() != nil {
		t.Errorf("%s failed to read configuration fileName '%s'", funcName, fileName)
	}
	if len(data.messages) == 0 {
		t.Errorf("%s failed to check for parsing error", funcName)
	}
}

func TestFileOverwriteDefault(t *testing.T) {
	const funcName = "TestFileOverwriteDefault"
	const fileName = "/tmp/TestFileBadData.cfg"
	const fileData = "Name="

	if err := ioutil.WriteFile(fileName, []byte(fileData), 0644); err != nil {
		t.Errorf("%s could not write temporary fileName '%s'", funcName, fileName)
		return
	}
	defer os.Remove(fileName)

	var c testConfig
	data := localData{userStruct: &c, fileName: fileName}
	data.loadDefaults()

	if data.loadFile() != nil {
		t.Errorf("%s failed to read configuration fileName '%s'", funcName, fileName)
	}
	if len(data.messages) != 0 {
		t.Errorf("%s got parsing error(s)", funcName)
	}

	if c.Name != "" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "")
	}
}

func TestFileSimple(t *testing.T) {
	const funcName = "TestFileSimple"
	const fileName = "/tmp/TestFileSimple.cfg"
	const fileData = "age=28\nsalary=1800.00\nNAME=Jantje"

	if err := ioutil.WriteFile(fileName, []byte(fileData), 0644); err != nil {
		t.Errorf("%s could not write temporary fileName '%s'", funcName, fileName)
		return
	}
	defer os.Remove(fileName)

	var c testConfig
	data := localData{userStruct: &c, fileName: fileName}
	data.loadDefaults()

	if data.loadFile() != nil {
		t.Errorf("%s could not load config fileName '%s'", funcName, fileName)
	}
	if len(data.messages) != 0 {
		t.Errorf("%s got parsing error(s)", funcName)
	}

	if c.Name != "Jantje" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "Jantje")
	}
	if c.Age != 28 {
		t.Errorf("%s age mismatch; got: %d; want: %d", funcName, c.Age, 28)
	}
	if c.Salary != 1800.0 {
		t.Errorf("%s salary mismatch; got: %.2f; want: %.2f", funcName, c.Salary, 1800.0)
	}
	if c.Host != "localhost" {
		t.Errorf("%s host mismatch; got: %s; want: %s", funcName, c.Host, "localhost")
	}
	if c.Port != 8080 {
		t.Errorf("%s port mismatch; got: %d; want: %d", funcName, c.Port, 8080)
	}
	if c.Secure != false {
		t.Errorf("%s secure mismatch; got: %t; want: %t", funcName, c.Secure, false)
	}
}

func TestFileComplex(t *testing.T) {
	const funcName = "TestFileComplex"
	const fileName = "/tmp/TestFileComplex.cfg"
	const fileData = "// Age\nage=28\n#Salary\nsalary=1800.00  //it's ok ! \n\nsettle_name=Jantje ## Name of subject ##\n\n\n"

	if ioutil.WriteFile(fileName, []byte(fileData), 0644) != nil {
		t.Errorf("%s could not write temporary fileName '%s'", funcName, fileName)
		return
	}
	defer os.Remove(fileName)

	var c testConfig
	data := localData{userStruct: &c, fileName: fileName}
	data.loadDefaults()

	if data.loadFile() != nil {
		t.Errorf("%s could not load config fileName '%s'", funcName, fileName)
	}
	if len(data.messages) != 0 {
		t.Errorf("%s got parsing error(s)", funcName)
	}

	if c.Name != "Jantje" {
		t.Errorf("%s name mismatch; got: %s; want: %s", funcName, c.Name, "Jantje")
	}
	if c.Age != 28 {
		t.Errorf("%s age mismatch; got: %d; want: %d", funcName, c.Age, 28)
	}
	if c.Salary != 1800.0 {
		t.Errorf("%s salary mismatch; got: %.2f; want: %.2f", funcName, c.Salary, 1800.0)
	}
	if c.Host != "localhost" {
		t.Errorf("%s host mismatch; got: %s; want: %s", funcName, c.Host, "localhost")
	}
	if c.Port != 8080 {
		t.Errorf("%s port mismatch; got: %d; want: %d", funcName, c.Port, 8080)
	}
	if c.Secure != false {
		t.Errorf("%s secure mismatch; got: %t; want: %t", funcName, c.Secure, false)
	}
}
