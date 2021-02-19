package main

import (
	"os"
	"testing"
)

const TESTDB_FILE = "testdb.sqlite"
const TESTDB_TYPE = "sqlite3"
const TESTCONF_TEMPLATE = "config.example.json"
const TESTCONF_FILE = "config.json"

const CONF_NIL = 0
const CONF_GLOBAL = 1
const CONF_DB = 2
const CONF_READY = 3

func InitTesting(step int) {
	// Setup configuration file
	if step >= CONF_NIL {
		_, err := os.Stat(TESTCONF_FILE)
		if err != nil && os.IsNotExist(err) {
			os.Link(TESTCONF_TEMPLATE, TESTCONF_FILE)
		}
	}
	// Setup global configuration
	if step >= CONF_GLOBAL {
		var emptyConfiguration Configuration
		if configuration == emptyConfiguration {
			LoadConfiguration()
		}
	}
	// Setup database configuration
	if step >= CONF_DB {
		if dbString != TESTDB_FILE {
			os.Remove(TESTDB_FILE)
			dbType = TESTDB_TYPE
			dbString = TESTDB_FILE
		}
	}
	// Setup database
	if step >= CONF_READY {
		CheckDB()
	}
}

func TestLoadConfiguration(t *testing.T) {
	InitTesting(CONF_NIL)

	LoadConfiguration()

	var emptyConfiguration Configuration
	if configuration == emptyConfiguration {
		t.Error("expected configuration to be valid after loading")
	}
}

func TestGetDB(t *testing.T) {
	InitTesting(CONF_DB)

	db := GetDB()
	if db == nil {
		t.Error("expected valid database")
	}
	defer db.Close()

	err := db.Ping()
	if err != nil {
		t.Error("expected pingable database")
	}
}

func TestCheckDB(t *testing.T) {
	InitTesting(CONF_DB)

	db := GetDB()
	defer db.Close()
	_, err := db.Query("SELECT 1 FROM pastebin LIMIT 1")
	if err == nil {
		t.Error("expected SELECT on non existent table to fail")
	} else if err.Error() != "no such table: pastebin" {
		t.Error("did not expect database error : ", err)
	}

	CheckDB()

	res, err := db.Query("SELECT 1 FROM pastebin LIMIT 1")
	if err != nil && err.Error() != "no such table: pastebin" {
		t.Error("expected database to be populated after CheckDB")
	} else if err != nil {
		t.Error("did not expect database error : ", err)
	} else if res.Err() != nil {
		t.Error("did not expect database row error : ", res.Err())
	}
}

func TestValidPasteId(t *testing.T) {
	InitTesting(CONF_READY)

	paste := "abcde"
	if !ValidPasteId(paste) {
		t.Errorf("expected %s to be a valid paste ID", paste)
	}
	paste = "favicon.ico"
	if ValidPasteId(paste) {
		t.Errorf("expected %s to be an invalid paste ID", paste)
	}
	paste = "1234567890123456789012345678901"
	if ValidPasteId(paste) {
		t.Errorf("expected %s to be a invalid paste ID", paste)
	}
}

func TestGenerateName(t *testing.T) {
	InitTesting(CONF_READY)

	name, err := GenerateName()
	if err != nil {
		t.Error("did not expect error : ", err)
	}
	if !ValidPasteId(name) {
		t.Errorf("expected GenerateName to output a valid paste ID : %s", name)
	}
	name2, _ := GenerateName()
	if name == name2 {
		t.Errorf("expected %s and %s to be 2 different GenerateName results", name, name2)
	}

	db := GetDB()
	defer db.Close()
	db.Exec("INSERT INTO pastebin (id) VALUES (?)", name)
	name3, _ := GenerateName()
	if name == name3 {
		t.Errorf("expected %s and %s to be 2 different GenerateName results", name, name3)
	}
}

func TestSha1(t *testing.T) {
	InitTesting(CONF_READY)

	input := ""
	expected := "2jmj7l5rSw0yVb_vlWAYkK_YBwk="
	output := Sha1(input)
	if output != expected {
		t.Errorf("expected sha1('') == %s got %s instead", expected, output)
	}

	input = "abcde123456_$^ù*ø"
	expected = "FEpsLSpz-VoymSaM2efFIX4vV44="
	output = Sha1(input)
	if output != expected {
		t.Errorf("expected sha1(%s) == %s go %s instead", input, expected, output)
	}
}

func TestDurationFromExpiry(t *testing.T) {
	InitTesting(CONF_READY)

	expiry := ""
	expected := 175200.0
	dur := DurationFromExpiry(expiry)
	if dur.Hours() != expected {
		t.Errorf("expected DurationFromExpiry(%s) == %f got %f", expiry, expected, dur.Hours())
	}

	expiry = "P100Y"
	dur = DurationFromExpiry(expiry)
	if dur.Hours() != expected {
		t.Errorf("expected DurationFromExpiry(%s) == %f got %f", expiry, expected, dur.Hours())
	}

	expiry = "PT5M"
	expected = 5
	dur = DurationFromExpiry(expiry)
	if dur.Minutes() != expected {
		t.Errorf("expected DurationFromExpiry(%s) == %f got %f", expiry, expected, dur.Minutes())
	}
}

func TestSave(t *testing.T) {
	InitTesting(CONF_READY)

	content := ""
	expiry := ""
	res, err := Save(content, expiry)
	if err != nil {
		t.Error("did not expect error ", err)
	}

	if res.Size != len(content) {
		t.Errorf("expected input len %d == output len %d", len(content), res.Size)
	}

	content = "ABCdef123_i"
	res, err = Save(content, expiry)
	res2, err2 := Save(content, expiry)
	if err2 != nil {
		t.Error("did not expect error on second Save() ", err)
	}
	if res.ID != res2.ID {
		t.Errorf("expected two identical Save() to give same paste ID : %s != %s", res.ID, res2.ID)
	}
	if res.URL != res2.URL {
		t.Errorf("expected two identical Save() to give same URL : %s != %s", res.URL, res2.URL)
	}
	if res.Sha1 != res2.Sha1 {
		t.Errorf("expected two identical Save() to give same hash : %s != %s", res.Sha1, res2.Sha1)
	}
	if res.Size != res2.Size {
		t.Errorf("expected two identical Save() to give same size : %d != %d", res.Size, res2.Size)
	}
	if res.Delkey != res2.Delkey {
		t.Errorf("expected two identical Save() to give same delkey : %s != %s", res.Delkey, res2.Delkey)
	}

	content = "ABCdef123_I"
	res3, _ := Save(content, expiry)
	if res == res3 {
		t.Error("expected two Save() with different content to be different")
	}
}

func TestGetPaste(t *testing.T) {
	InitTesting(CONF_READY)

	pasteId := "favicon.ico"
	paste, err := GetPaste(pasteId)
	if err == nil {
		t.Errorf("expected invalid paste id %s to return error", pasteId)
	}
	if paste != "" {
		t.Errorf("expected invalid paste id %s to return nil paste", pasteId)
	}

	pasteId, err = GenerateName()
	if err != nil {
		t.Error("GenerateName() failed for some reason ", err)
	}

	paste, err = GetPaste(pasteId)
	if err == nil {
		t.Errorf("expected unknown paste id %s to return error", pasteId)
	}
	if paste != "" {
		t.Errorf("expected unknown paste id %s to return nil paste", pasteId)
	}

	content := "testcontent"
	expiry := "PT5M"
	res, err := Save(content, expiry)
	if err != nil {
		t.Error("Save() failed for some reason ", err)
	}

	paste, err = GetPaste(res.ID)
	if err != nil {
		t.Error("did not expect freshly saved paste to give error ", err)
	}
	if paste != content {
		t.Errorf("expected '%s' saved content to return identical, got '%s' instead", content, paste)
	}
}

func TestHeavyUsage(t *testing.T) {
	InitTesting(CONF_READY)

	for i := 0; i < 1000; i++ {
		content := "testcontent"
		expiry := "PT5M"
		res, err := Save(content, expiry)
		if err != nil {
			t.Error("Save() failed for some reason ", err)
		}

		if i%2 == 0 {
			paste, err := GetPaste(res.ID)
			if err != nil {
				t.Error("GetPaste() failed for some reason ", err)
			}

			if paste != content {
				t.Errorf("expected content %s == paste %s", content, paste)
			}
		}
	}
}

func BenchmarkSave(b *testing.B) {
	InitTesting(CONF_READY)

	expiry := "PT5M"
	for i := 0; i < b.N; i++ {
		content := "testcontent" + string(i)
		_, err := Save(content, expiry)
		if err != nil {
			b.Error("Save() failed for some reason ", err)
		}
	}
}

func BenchmarkGetPaste(b *testing.B) {
	InitTesting(CONF_READY)

	content := "testcontent"
	expiry := "PT5M"
	res, err := Save(content, expiry)
	if err != nil {
		b.Error("Save() failed for some reason ", err)
	}

	for i := 0; i < b.N; i++ {
		_, err = GetPaste(res.ID)
		if err != nil {
			b.Error("GetPaste() failed for some reason ", err)
		}
	}
}
