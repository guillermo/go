package rfc2047

import (
	"fmt"
	"testing"
)

type TestCase struct {
	input, out string
}

func (c TestCase) Test(t *testing.T) {
	r := &Rfc2047{S: c.input}
	result := r.String()
	if result != c.out {
		t.Errorf("(%d) Should be '%s':\ninput:  '%s'\noutput: '%s'", len(r.Errors), c.out, c.input, result)
	}
	if len(r.Errors) != 0 {
		for _, e := range r.Errors {
			t.Error(e)
		}
	}
}

func TestRfc2047(t *testing.T) {
	TestCase{"=?ISO-8859-1?Q?a?=", "a"}.Test(t)
	TestCase{"=?ISO-8859-1?Q?a?= b", "a b"}.Test(t)

	//Spaces between tokens should be removed
	// This one is failing
	TestCase{"=?ISO-8859-1?Q?a?= \n =?ISO-8859-1?Q?b?=", "ab"}.Test(t)

	TestCase{"=?ISO-8859-1?Q?a_b?=", "a b"}.Test(t)
	TestCase{"=?ISO-8859-1?Q?a_b?=", "a b"}.Test(t)

	TestCase{"=?US-ASCII?Q?Keith_Moore?= <moore@cs.utk.edu>", "Keith Moore <moore@cs.utk.edu>"}.Test(t)
	TestCase{"=?ISO-8859-1?Q?Keld_J=F8rn_Simonsen?= <keld@dkuug.dk>", "Keld Jørn Simonsen <keld@dkuug.dk>"}.Test(t)
	TestCase{"=?ISO-8859-1?Q?Andr=E9?= Pirard <PIRARD@vm1.ulg.ac.be>", "André Pirard <PIRARD@vm1.ulg.ac.be>"}.Test(t)
	TestCase{"=?ISO-8859-1?B?SWYgeW91IGNhbiByZWFkIHRoaXMgeW8=?=\n=?ISO-8859-2?B?dSB1bmRlcnN0YW5kIHRoZSBleGFtcGxlLg==?=", "If you can read this yo\nu understand the example."}.Test(t)
	TestCase{"=?ISO-8859-1?Q?Olle_J=E4rnefors?= <ojarnef@admin.kth.se>", "Olle Järnefors <ojarnef@admin.kth.se>"}.Test(t)
	TestCase{"=?ISO-8859-1?Q?Patrik_F=E4ltstr=F6m?= <paf@nada.kth.se>", "Patrik Fältström <paf@nada.kth.se>"}.Test(t)
	TestCase{"Nathaniel Borenstein <nsb@thumper.bellcore.com>\n(=?iso-8859-8?b?7eXs+SDv4SDp7Oj08A==?=)", "Nathaniel Borenstein <nsb@thumper.bellcore.com>\n(םולש ןב ילטפנ)"}.Test(t)

}

func ExampleConvert() {
	fmt.Println(Convert("=?ISO-8859-1?Q?Patrik_F=E4ltstr=F6m?="))

	// Output:
	// Patrik Fältström
}

func ExampleRfc2047_String() {
	s := &Rfc2047{S: "=?ISO-666-1?Q?The_ch=65rset_is_weird?="}
	fmt.Println(s.String())
	if len(s.Errors) > 0 {
		for _, e := range s.Errors {
			fmt.Println(e)
		}
	}
	// Output:
	// The cherset is weird
	// Transcoding: [101] : character set "iso-666-1" not found
}
