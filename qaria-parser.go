package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

/**
 * programma a linea di comando per lettura da pagine html dei dati di inquinamento
 *
 */

// Misura struct per contenere i dati di una misura
type Misura struct {
	Inq        string
	Data       string
	Valore     float64
	StazioneID int
	ComuneID   int
}

// ToCSV presenta i dati della Misura in formato CSV
func (m Misura) ToCSV() string {
	return strconv.Itoa(m.StazioneID) + "," + m.Inq + "," + m.Data + "," + strconv.FormatFloat(m.Valore, 'g', 3, 64)
}

// Valori struct per contenere i dati letti dal file html
type Valori struct {
	Inq    string
	Valori []float64
	Date   []string
}

func main() {
	// directory dove salvare il file html
	dir := flag.String("d", ".", "outputdir")

	flag.Parse()

	log.Printf(fmt.Sprintf("directory lettura = %s", *dir))

	files, err := filepath.Glob(*dir + "/*.html")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		misure := leggiFileEEstraiDati(f)
		linee := misureToCSV(misure)
		writeLines(linee, strings.Replace(f, ".html", ".txt", -1))

	}
}
func misureToCSV(misuras []Misura) []string {
	var res []string
	for i := 0; i < len(misuras); i++ {
		m := misuras[i]
		res = append(res, m.ToCSV())
	}
	return res
}

func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

// EstraiStringaDati torna un array di stringhe nella forma "PM10 =  [45,47,91,86,68,25,23,28,33,26 ]"
func estraiStringaDati(s string) []string {
	var res []string
	var i int
	i = strings.Index(s, "dati_")
	for i > -1 {
		var sDati string
		sDati = s[i+5:]
		j := strings.Index(sDati, "]")
		sDati = sDati[:j+1]
		if !strings.HasPrefix(sDati, "\" ") {
			res = append(res, sDati)
		}
		s = s[i+len(sDati):]
		i = strings.Index(s, "dati_")
	}
	return res
}

func estraiFloats(valoriAsString string) []float64 {
	var result []float64

	subs := strings.Trim(valoriAsString[1+strings.Index(valoriAsString, "["):strings.Index(valoriAsString, "]")], " ")
	valoriStr := strings.Split(subs, ",")
	for i := 0; i < len(valoriStr); i++ {
		d, _ := strconv.ParseFloat(valoriStr[i], 64)
		result = append(result, d)
	}
	return result
}

// StrReplace oldnew è la mappa con valoreVecchio => valoreNuovo
func StrReplace(s string, oldnew map[string]string) string {
	var res string
	res = s
	for key, value := range oldnew {
		res = strings.Replace(res, key, value, -1)
	}
	return res
}

func estraiDate(valoriAsString string) []string {
	var result []string

	if strings.Trim(valoriAsString, " ") == "" {
		return result
	}
	subs := strings.Trim(valoriAsString[1+strings.Index(valoriAsString, "["):strings.Index(valoriAsString, "]")], " ")
	valoriStr := strings.Split(subs, ",")

	// valori da sosituire nell stringa della data
	oldnew := map[string]string{
		"<br/>":  "",
		"<b>":    "",
		"<br />": " ",
		"</b>":   "",
	}

	for i := 0; i < len(valoriStr); i++ {
		s2 := StrReplace(valoriStr[i], oldnew)
		result = append(result, convertiData(strings.Trim(s2, "'")))
	}
	return result
}

func costruisciValori(stringheDati []string) []Valori {
	var result []Valori

	for i := 0; i < len(stringheDati); i++ {
		s := stringheDati[i]
		inq := s[:strings.Index(s, " ")]

		valore := Valori{Inq: strings.Trim(inq, " "), Valori: estraiFloats(s)}
		result = append(result, valore)
	}
	return result
}

// torna una stringa del tipo " [ '<b>31<br />Ott</b>','<b>1<br />Nov</b>','<b>2<br />Nov</b>','<b>3<br />Nov</b>','<b>4<br />Nov</b>','<b>5<br />Nov</b>','<b>6<br />Nov</b>','<b>7<br />Nov</b>','<b>8<br />Nov</b>','<b>9<br />Nov</b>' ]"
func estraiStringaDate(s string) string {
	var res string
	var i int
	i = strings.Index(s, "var ticks =")
	if i > -1 {
		var sDati string
		sDati = s[i+len("var ticks ="):]
		j := strings.Index(sDati, "]")
		res = sDati[:j+1]
	}
	return strings.Trim(res, " ")
}

// nel file html, la data ha la forma 1 Ott
// dobbiamo convertirla in 20161201
func convertiData(dataHTML string) string {
	pezzi := strings.Split(dataHTML, " ")
	mese := "01"
	switch pezzi[1] {
	case "Gen":
		mese = "01"
	case "Feb":
		mese = "02"
	case "Mar":
		mese = "03"
	case "Apr":
		mese = "04"
	case "Mag":
		mese = "05"
	case "Giu":
		mese = "06"
	case "Lug":
		mese = "07"
	case "Ago":
		mese = "08"
	case "Set":
		mese = "09"
	case "Ott":
		mese = "10"
	case "Nov":
		mese = "11"
	case "Dic":
		mese = "12"
	}

	// FIXME non posso schiantare 2017!!
	if len(pezzi[0]) < 2 {
		return "2017" + mese + "0" + pezzi[0]
	}
	return "2017" + mese + pezzi[0]
}

func leggiFileEEstraiDati(filename string) []Misura {
	var res []Misura
	log.Println("lettura di ", filename)
	re := regexp.MustCompile("[0-9]+")
	matchSlice := re.FindAllString(filename, -1)
	stazID, _ := strconv.Atoi(matchSlice[1])
	buf := bytes.NewBuffer(nil)

	f, _ := os.Open(filename) // Error handling elided for brevity.
	io.Copy(buf, f)           // Error handling elided for brevity.
	f.Close()

	fullhtml := string(buf.Bytes())

	/*
		var dati_NO2 =  [137,123,111,113,94,71,90,84,101,89 ] ;
		var colors_NO2 =  ['#ffde33' ,'#ffde33' ,'#ffde33' ,'#ffde33' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#ffde33' ,'#40b38c'  ] ;


		var dati_CO =  [1,1,1.4,1.3,1.2,1.1,1.1,1.2,1.1,1.3 ] ;
		var colors_CO =  ['#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c'  ] ;


		var dati_C6H6 =  [0.5,0.5,0.8,2.1,1.5,0.2,1.4,1.6,0.7,1.2 ] ;
		var colors_C6H6 =  ['#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c' ,'#40b38c'  ] ;


		// tick Array
		var ticks = [ '<b>31<br />Ott</b>','<b>1<br />Nov</b>','<b>2<br />Nov</b>','<b>3<br />Nov</b>','<b>4<br />Nov</b>','<b>5<br />Nov</b>','<b>6<br />Nov</b>','<b>7<br />Nov</b>','<b>8<br />Nov</b>','<b>9<br />Nov</b>' ] ;   // ['-10', '-9', '-8',];
	*/

	dateArrayStr := estraiStringaDate(fullhtml)
	dateArray := estraiDate(dateArrayStr)

	stringheDati := estraiStringaDati(fullhtml)
	valoriArray := costruisciValori(stringheDati)
	for i := 0; i < len(valoriArray); i++ {
		val := valoriArray[i]
		val.Date = dateArray
		for j := 0; j < len(val.Valori); j++ {
			res = append(res, Misura{Inq: val.Inq, Data: val.Date[j], Valore: val.Valori[j], StazioneID: stazID, ComuneID: 161})
		}
	}
	return res
}
