#!/bin/bash

# Preverimo, če imamo podane vse potrebne argumente
if [ "$#" -ne 4 ]; then
    echo "Uporaba: $0 <pot_do_testne_datoteke> <pot_do_glavne_zip_datoteke> <izhodna_pot> <ime_modula>"
    exit 1
fi

# Definiranje poti
test_file_path=$1
main_zip_file=$2
output_path=$3
module_name=$4

# Ustvarimo izhodno mapo, ce ta se ne obstaja
mkdir -p "$output_path"

# Ekstraktamo glavni ZIP, ki vsebuje ZIP-e studentov
unzip -q "$main_zip_file" -d "$output_path"

# Poiscemo mapo, ki se je ustvarila znotraj izhodne mape po ekstrakciji
main_folder=$(find "$output_path" -mindepth 1 -maxdepth 1 -type d)

# Ekstrakcija ZIP datotek za vsakega studenta znotraj glavne mape
for student_zip in "$main_folder"/*.zip; do
    unzip -q "$student_zip" -d "$output_path"
done

# Odstranimo ZIP datoteke studentov po ekstrakciji
rm "$main_folder"/*.zip

# Odstranimo tudi glavno mapo, ki je bila ustvarjena znotraj izhodne mape
rm -rf "$main_folder"

# Inicializacija CSV datoteke
csv_file="$output_path/results.csv"
echo "Ime in Priimek,Studentski Mail,Vpisna Stevilka,Rezultat" > "$csv_file"

# Iteracija skozi vse mape studentov
for student_folder in "$output_path"/*/; do
    # Preverimo, da ne gre za glavno mapo
    if [ "$student_folder" != "$main_folder/" ]; then
        student_id=$(basename "$student_folder")
        
        # Razdelimo ime mape v dele (Ime Priimek, Student Email, Vpisna Stevilka)
        IFS='=' read -r ime_priimek student_email vpisna_stevilka <<< "${student_id%%_*}"
        
        # Če vpisna stevilka ni prisotna, uporabimo _
        if [ -z "$vpisna_stevilka" ]; then
            vpisna_stevilka="_"
        fi
        
        # Kopiramo vse datoteke iz testne mape v studentsko mapo, razen main.go
        for file in "$test_file_path"/*; do
            if [ "$(basename "$file")" != "main.go" ]; then
                cp -r "$file" "$student_folder"
            fi
        done
        
        # Gremo v mapo studenta
        cd "$student_folder" || continue
        
        # Inicializiramo go mod z podanim imenom modula
        go mod init "$module_name"
        
        # Uporabimo go mod tidy za urejanje odvisnosti
        go mod tidy
        
        # Shranimo rezultat testa v test.res
        go test -v > test.res
        
        # Preverimo, ali je test uspesno opravljen ali ne
        if grep -q "FAIL" test.res; then
            test_result=0
        else
            test_result=1
        fi
        
        # Dodamo podatke v CSV datoteko
        echo "$ime_priimek,$student_email,$vpisna_stevilka,$test_result" >> "../results.csv"
        
        # Vrnemo se v izhodno mapo
        cd - > /dev/null
    fi
done

echo "Testiranje koncano. Rezultati so shranjeni v mapah studentov in v datoteki $csv_file."