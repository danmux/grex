package app

import (
	"testing"
)

func Test_EmailRegex(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	PrepRest()

	if !ValidateEmail("dna@man.com") {
		t.Error("failed to pass a correct email")
	}

	if ValidateEmail("dna@mancom") {
		t.Error("didnt spot bad email")
	}

	if ValidateEmail("dnamancom") {
		t.Error("didnt spot bad email")
	}

	if ValidateEmail("dna@mancom") {
		t.Error("didnt spot bad email again ")
	}

	if ValidateEmail("dnaman.com") {
		t.Error("didnt spot bad email again ")
	}

	if !ValidateEmail("dna@man.co.uk") {
		t.Error("failed to pass a correct email")
	}

	if !ValidateEmail("dna@man.co.uk") {
		t.Error("failed to pass a correct email")
	}
}

func Test_TidyEmail(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	if TidyEmail("  ffdfs @ ban . Com d f Ff   ") != "ffdfs@ban.comdfff" {
		t.Error("tidy didnt tidy email properly")
	}

	email, valid := ValidateAndTidyEmail("  ffdfs @ ban . Com d f Ff   ")

	if !valid {
		t.Error("tidy validate messy email properly")
	}

	if email != "ffdfs@ban.comdfff" {
		t.Error("tidy email returned messy email")
	}
}

func Test_FldTidy(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	if TidyInput("  ffdfs d f ff   ") != "ffdfs d f ff" {
		t.Error("tidy didnt strip spaces properly")
	}
}

func Test_KeyTidy(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	if TidyKey("  ffdfs df   ff  FGG  ") != "ffdfs_df_ff_fgg" {
		t.Error("key tidy didnt tidy >" + TidyKey("  ffdfs df   ff  FGG  "))
	}

	if TidyKey("  aaA შოთ Aaa") != "aaa_aaa" {
		t.Error("key tidy didnt tidy >" + TidyKey("  aaA შოთ Aaa"))
	}

	if TidyKey("  aaA შოთ Bშოთaa") != "aaa_baa" {
		t.Error("key tidy didnt tidy >" + TidyKey("  aaA შოთ Aaa"))
	}
}
