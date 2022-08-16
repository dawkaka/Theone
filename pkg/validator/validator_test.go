package validator

import (
	"os"
	"testing"
	"time"
)

func TestIsRealName(t *testing.T) {
	validTest := []string{"Yussif", "Marcos", "D'angelo", "李", "静", "राजेश", "يوسف"}
	inValidTest := []string{"yUssif", "Brown45", "Ruben_", " ", "Sharp Brown", "ShaTta"}

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
	validTest := []string{"yousiph77@gmail.com", "what@blue.com", "somethign@outlook.com", "see@yahoo.net"}
	inValidTest := []string{"youiosw!gmail.com", "youi djlfsf@brown.com", "shit.@gmail.com", "al;a@gmail.com"}
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

func TestIsValidPastDate(t *testing.T) {
	pastDate := time.Date(2006, time.March, 16, 0, 0, 0, 0, time.Local)
	futureDate := time.Date(2025, time.January, 33, 3, 3, 3, 3, time.Local)

	if !IsValidPastDate(pastDate) {
		t.Fatalf("Testing %s. Expected %t got %t", pastDate, true, false)
	}

	if IsValidPastDate(futureDate) {
		t.Fatalf("Testing %s. Expected %t got %t", futureDate, true, false)
	}
}

func TestIsPronouns(t *testing.T) {
	valid := []string{"she/her", "he/him", "they/them", "water/fire", "blue/green"}
	inValid := []string{"she.her", "he", "/him", "they/", "sometthing/ "}

	for _, val := range valid {
		if !IsPronouns(val) {
			t.Fatalf("Testing %s. Expected %t got %t", val, true, false)
		}
	}
	for _, val := range inValid {
		if IsPronouns(val) {
			t.Fatalf("Testing %s. Expected %t got %t", val, false, true)
		}
	}
}

func TestIsValidSetting(t *testing.T) {
	setting := "language"
	for _, val := range Settings[setting] {
		if !IsValidSetting(setting, val) {
			t.Fatalf("Testing %s and %s. Expected %t got %t", setting, val, true, false)
		}
	}
	value := "jp"
	if IsValidSetting(setting, value) {
		t.Fatalf("Testing %s and %s. Expected %t got %t", setting, value, false, true)
	}
	value = "en_US"
	if IsValidSetting(setting, value) {
		t.Fatalf("Testing %s and %s. Expected %t got %t", setting, value, false, true)
	}
	setting = "Occupation"
	value = "ar"
	if IsValidSetting(setting, value) {
		t.Fatalf("Testing %s and %s. Expected %t got %t", setting, value, false, true)
	}

}

func TestIsSupportedFileType(t *testing.T) {
	image, err := os.Open("/home/dawkaka/Pictures/yz series.jpg")
	if err != nil {
		panic(err)
	}
	pdf, err := os.Open("/home/dawkaka/Documents/Data-Structures-and-Algorithms-in-Java-6th-Edition.pdf")
	if err != nil {
		panic(err)
	}
	pdfBuffer := make([]byte, 512)
	imgBuffer := make([]byte, 512)

	pdf.Read(pdfBuffer)
	image.Read(imgBuffer)

	if i, val := IsSupportedFileType(imgBuffer); !val {
		t.Fatalf("Testing %s. Expected %t got %t", i, true, false)
	}

	if i, val := IsSupportedFileType(pdfBuffer); val {
		t.Fatalf("Testing %s. Expected %t got %t", i, false, true)
	}
}

func TestIs18Plus(t *testing.T) {
	validAge, err := time.Parse("2006-01-02", "1993-03-16")
	if err != nil {
		panic(err)
	}
	if !Is18Plus(validAge) {
		t.Errorf("Testing %v. expected %t got %t", validAge, true, false)
	}

	validAge, err = time.Parse("2006-01-02", "2004-08-11")
	if err != nil {
		panic(err)
	}
	if !Is18Plus(validAge) {
		t.Errorf("Testing %v. expected %t got %t", validAge, true, false)
	}

	validAge, err = time.Parse("2006-01-02", "2005-10-16")
	if err != nil {
		panic(err)
	}
	if Is18Plus(validAge) {
		t.Errorf("Testing %v. expected %t got %t", validAge, false, true)
	}
}
