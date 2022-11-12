package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

// Údaje pro přihlášení k e-mailu
const HOST = "smtp.seznam.cz"
const PORT = 25
const EMAIL = "registrace@hackdays.eu"
const PASSWORD = "REDACTED"

const SIGNATURE = "\n\n" +
	"S pozdravem\n" +
	"Vojtěch Jungmann\n" +
	"za tým KB na Smíchovské SPŠaG"

// Obsah a nadpis zprávy pro klasický certifikát
const SUBJECT_CLASSIC = "Certifikát HackDays! 2022"
const BODY_CLASSIC = "Dobrý den,\nděkujeme za Vaší účast na kempu kybernetické bezpečnosti - HackDays! 2022." +
	" Do přílohy přikládáme digitálně podepsaný a ověřený certifikát o absolvování kurzu." + SIGNATURE

// Obsah a nadpis zprávy pro golden certifikát
const SUBJECT_GOLDEN = "Certifikát HackDays! 2022 | Zlatá edice"
const BODY_GOLDEN = "Dobrý den,\nza projevené znalosti oboru kybernetické bezpečnosti Vám navíc zasíláme zlatý „Golden“ certifikát" +
	"V příloze naleznete digitálně podepsané PDF" + SIGNATURE

const DELIMETER = "**=85541b8ef226c79d6e1872c048bc1552"

// Název odeslaného souboru
const CERTIFICATENAME = "Certifikát.pdf"

func main() {
	var people string
	var path string
	var goldenCertificate bool

	// -f cesta k souboru se jmény ve formátu Jméno Přijmení E-mail\nJméno Přijmení E-mail\n...
	flag.StringVar(&people, "f", "", "Cesta k souboru s údaji")
	// -c cesta ke složce s certifikáty
	flag.StringVar(&path, "c", "", "Složka s certifikáty")
	// -g odeslat zlatý certifikát
	flag.BoolVar(&goldenCertificate, "g", false, "Odeslat zlatý certifikát")

	// Zpracování argumentů
	flag.Parse()

	// Dekódování hesla z base64
	password, _ := base64.StdEncoding.DecodeString(PASSWORD)

	// Otevření souboru se jmény
	file, err := os.Open(people)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Přihlášení k SMTP
	auth := smtp.PlainAuth("", EMAIL, string(password), HOST)

	// Vytvoření e-mailové zprávy
	messageTemplate := fmt.Sprintf("From: %s\r\n", EMAIL)
	messageTemplate += fmt.Sprintf("To: %s\r\n", EMAIL)

	if !goldenCertificate {
		messageTemplate += "Subject: " + SUBJECT_CLASSIC + "\r\n"
	} else {
		messageTemplate += "Subject: " + SUBJECT_GOLDEN + "\r\n"
	}

	messageTemplate += "MIME-Version: 1.0\r\n"
	messageTemplate += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", DELIMETER)

	messageTemplate += fmt.Sprintf("\r\n--%s\r\n", DELIMETER)
	messageTemplate += "Content-Type: text/plain; charset=\"utf-8\"\r\n"
	messageTemplate += "Content-Transfer-Encoding: 7bit\r\n"

	if !goldenCertificate {
		messageTemplate += fmt.Sprintf("\r\n%s", BODY_CLASSIC+"\r\n")
	} else {
		messageTemplate += fmt.Sprintf("\r\n%s", BODY_GOLDEN+"\r\n")
	}

	// Čtení jmen řádek po řádku
	lineReader := bufio.NewScanner(file)
	for lineReader.Scan() {
		line := lineReader.Text()
		// Rozdělení rádku podle mezer na 3 části
		parts := strings.Fields(line)
		if len(parts) < 3 {
			log.Fatal("Špatný formát vstupních dat")
		}
		name := parts[0]
		surname := parts[1]

		email := parts[2]
		if len(parts) > 3 {
			email = parts[3]
		}

		// Je třetí část e-mail?
		_, err := mail.ParseAddress(email)
		if err != nil {
			log.Fatal(err)
		}

		// Adresa příjemce
		to := []string{
			email,
		}

		// Zkopírování messageTemplate do message, neefektivní, ale co, RAMky je hodně a tohle je zanedbatelné.... Lepší funkce existuje až v Go 1.18
		message := string([]byte(messageTemplate))

		// Přidání přílohy do zprávy
		message += fmt.Sprintf("\r\n--%s\r\n", DELIMETER)
		message += "Content-Type: application/octet-stream; charset=\"utf-8\"\r\n"
		message += "Content-Transfer-Encoding: base64\r\n"
		message += "Content-Disposition: attachment;filename=\"" + CERTIFICATENAME + "\"\r\n"

		filePath := filepath.Join(path, fmt.Sprintf("Certifikát_%s_%s.pdf", surname, name))
		certificateFile, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Panic(err)
		}
		// Encode souboru do Base64 a jeho vložení do e-mailu
		message += "\r\n" + base64.StdEncoding.EncodeToString(certificateFile)

		// Odeslání zprávy
		err = smtp.SendMail(fmt.Sprintf("%s:%d", HOST, PORT), auth, EMAIL, to, []byte(message))
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Odeslán e-mail na adresu %s, soubor %s", to[0], filePath)
		}

	}

}
