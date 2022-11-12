#!/bin/bash

readonly CLASSIC=CertficateClassic.sketch
readonly GOLDEN=CertficateGolden.sketch
readonly KOFOLALOVER=CertficateKofola.sketch

readonly SKETCH_PDF_EXPORT_NAME="A4.pdf"
readonly SKETCH_TOOL_PATH="/Applications/Sketch.app/Contents/MacOS/sketchtool"

readonly JAVA_BIN_PATH="/Users/vojtechjungmann/Java/bin/java"
readonly PKCS12_PASSWORD="REDACTED"

CERTIFICATE=$CLASSIC

# Nápověda
help() { echo "Použití: $0 [-f lidi.txt] [-g (Zlatý certifikát)] [-k (Kofola certifikát)]  [-d (Smazání temp a out) [-h (Nápověda)]]" 1>&2; exit 1; }

while getopts "f:gkdh" options; do
    case "${options}" in
        # Název textového souboru se jmény ve formátu [Jmeno Přijmení\n]
        f)
            f=${OPTARG}
            ;;
        # Generovat zlatý certifikát
        g)
            CERTIFICATE=${GOLDEN}
            ;;
        # Generovat certifikát pro organizátory
        k)
            CERTIFICATE=${KOFOLALOVER}
            ;;
        # Smazání souborů
        d)
            rm -rf out temp
            exit 1
            ;;
        # Nápověda
        h)
            help
            ;;
    esac
done

# Přepínač -f je prázdný
if [ -z "${f}" ]; then
    help
fi

# Kontrola argumentu přepínače f
if [ ! -f "${f}" ]; then
    echo "${f} není soubor!" 1>&2
    exit 1
fi

mkdir -p temp
mkdir -p out/_signed

while read name; do
  # Print jména
  echo "$name"
  # name_array[0] -> Jméno name_array[1] -> Přijmení
  name_array=($name)
  # Parametry pro úpravu patch.json na jméno účastníka
  jq_parameters="'.operations[0].value.string = \"${name_array[0]} ${name_array[1]}\"'" 
  cmd="cat patch.json | jq ${jq_parameters} > temp/patch.json"
  # Spuštění příkazu
  eval $cmd
  # Zkopírování certifikátu do temp
  cp "template/${CERTIFICATE}" temp/Certficate.sketch
  # Změna jména v certifikátu
  $SKETCH_TOOL_PATH patch temp/Certficate.sketch temp/patch.json
  # Export v certifikátu do PDF
  $SKETCH_TOOL_PATH export artboards temp/Certficate.sketch
  # Uložení certifikátu
  mv $SKETCH_PDF_EXPORT_NAME "out/Certifikát_${name_array[1]}_${name_array[0]}.pdf"
  # Podepsání certifikátu
  cmd="${JAVA_BIN_PATH} -jar cert/JSignPdf.jar -d out/_signed --out-suffix \"\" -kst PKCS12 -ksf cert/hackdays_certificate.p12 -ksp ${PKCS12_PASSWORD} out/Certifikát_${name_array[1]}_${name_array[0]}.pdf -c registrace@hackdays.eu -r \"Pro ověření autenticity HackDays certifikátu\""
  eval $cmd
done <$f

rm -rf temp