package validator

import (
	"testing"
)

func TestIsRealName(t *testing.T) {
	validTest := []string{"Yussif", "Marcos", "D'angelo", "李", "静", "राजेश", "يوسف"}
	inValidTest := []string{"yUssif", "Brown45", "Ruben_", " "}

	for _, value := range validTest {
		if !IsRealName(value) {
			t.Fatalf("Testing %s. Expected %t got %t", value, true, false)
		}
	}

	for _, value := range inValidTest {
		if IsRealName(value) {
			t.Fatalf("Testing %s. Expected %t got %t", value, false, true)
		}
	}
}

func TestIsUserName(t *testing.T) {
	validTest := []string{"Hamza_Jalil", "brown.5", "50serwaa", "smith25", "2222", "22.ussif"}
	inValidTest := []string{"bl", "somereallylong_user_name_tofail", "yuusif mohammed", "_ruben", "Zamani__25", "jay._5", "fati_", "22.22.22.", "a____._____b"}
	for _, value := range validTest {
		if !IsUserName(value) {
			t.Fatalf("Testing %s; Expected %t got %t", value, true, false)
		}
	}

	for _, value := range inValidTest {
		if IsUserName(value) {
			t.Fatalf("Testing %s; Expected %t got %t", value, false, true)
		}
	}
}

func TestIsCoupleName(t *testing.T) {
	validTest := []string{"jennifer&ruben", "brandi25&zainab", "yussif_&_kylie", "jalil.and.malil", "waterAndfire"}
	inValidTest := []string{"jelo", "sky&_.blue", "clown____and____jon", "_water.and.fire", "fire_", "somethingverylong_&_somethingextralongshitthatmakesyoucry", "you and me", " and "}
	for _, value := range validTest {
		if !IsCoupleName(value) {
			t.Fatalf("IsCoupleName: testing %s. Expected %t got %t", value, true, false)
		}
	}

	for _, value := range inValidTest {
		if IsCoupleName(value) {
			t.Fatalf("IsCoupleName: testing %s. Expected %t got %t", value, false, true)
		}
	}
}

func TestIsEmail(t *testing.T) {
	validTest := []string{}
	inValidTest := []string{}
	for _, value := range validTest {
		if !IsEmail(value) {
			t.Fatalf("IsEmail: testing %s. Expected %t got %t", value, true, false)
		}
	}

	for _, value := range inValidTest {
		if IsEmail(value) {
			t.Fatalf("IsEmail: testing %s. Expected %t got %t", value, false, true)
		}
	}
}

func TestBio(t *testing.T) {
	valid := `This is my bio from the ages of the gods and 
              that is the reason I want to tell you that you suck`
	invalid := "The NBER’s Business Cycle Dating Committee maintains a chronology of US business cycles. The chronology identifies the dates of peaks and troughs that frame economic recessions and expansions. A recession is the period between a peak of economic activity and its subsequent trough, or lowest point. Between trough and peak, the economy is in an expansion. Expansion is the normal state of the economy; most recessions are brief. However, the time that it takes for the economy to return to its previous peak level of activity or its previous trend path may be quite extended. According to the NBER chronology, the most recent peak occurred in February 2020. The most recent trough occurred in April 2020."

	if !IsBio(valid) {
		t.Fatalf("Testing: %s; exptected %t got %t", valid, true, false)
	}

	if IsBio(invalid) {
		t.Fatalf("Testing: %s; exptected %t got %t", invalid, false, true)
	}
}

func TestPassword(t *testing.T) {
	password := "somethingpassword"
	if !IsPassword(password) {
		t.Fatalf("Testing: %s; exptected %t got %t", password, true, false)
	}
	password = "some"
	if IsPassword(password) {
		t.Fatalf("Testing: %s; exptected %t got %t", password, false, true)
	}
}

func TestCaption(t *testing.T) {
	valid := `This is my bio from the ages of the gods and 
	that is the reason I want to tell you that you suck`
	invalid := `The NBER's Business Cycle Dating Committee maintains a chronology of US business cycles.
	 The chronology identifies the dates of peaks and troughs that frame economic recessions and
	  expansions. A recession is the period between a peak of economic activity and its subsequent t
	  rough, or lowest point. Between trough and peak, the economy is in an expansion. Expansion is 
	  the normal state of the economy; most recessions are brief. However, the time that it takes for 
	  the economy to return to its previous peak level of activity or its previous trend path may be
	   quite extended. According to the NBER chronology, the most recent peak occurred in February
	    2020. The most recent trough occurred in April 2020.`

	if !IsCaption(valid) {
		t.Fatalf("Testing: %s; exptected %t got %t", valid, true, false)
	}

	if IsCaption(invalid) {
		t.Fatalf("Testing: %s; exptected %t got %t", invalid, false, true)
	}
}

func TestIsWebsite(t *testing.T) {
	validTest := []string{"https://google.us.edi?34535/534534?dfg=g&fg", "http://RegExr.com?2rjl6", "gskinner.com/products/spl", "_water.and.fire", "blue.onlyfans.fire", "mongodb+srv://mongod:pasfword@a-free-cluster.zqbas.mongodb.net/somedb?retryWrites=true&w=majority"}
	inValidTest := []string{"jeloky&_.blueclown____and____jonwwwwwater.clownfire_somethingverylong_&_somethingextralongshitthatmakesyoucryyou and meand "}
	for _, value := range validTest {
		if !IsWebsite(value) {
			t.Fatalf("Testing %s. Expected %t got %t", value, true, false)
		}
	}

	for _, value := range inValidTest {
		if IsWebsite(value) {
			t.Fatalf("Testing %s. Expected %t got %t", value, false, true)
		}
	}
}
