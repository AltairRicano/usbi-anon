// cmd/hash_password es el binario standalone para el bootstrap del primer admin:
// sin base de datos ni red, imprime el hash Argon2id de una contraseña dada
// utilizando internal/crypto.HashPassword (el mismo que usa el servidor),
// de modo que el hash resultante sea válido para backend/sql/01_seed_primer_admin.sql.
//
// Con --seal se imprime también un UUID nuevo y el sello HMAC de accounts.privacy_acceptance_hash.
// Dado que ni gen_random_uuid() ni hmac() están disponibles en SQL puro sin pgcrypto,
// este binario genera estos valores reutilizando internal/crypto.GenerateHMAC para garantizar
// coherencia exacta con la verificación del servidor.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/google/uuid"
)

func main() {
	seal := flag.Bool("seal", false, "además del hash, imprime un UUID nuevo y el sello HMAC de privacy_acceptance_hash")
	version := flag.String("version", "", "privacy_notice_version a sellar (requerido con -seal)")
	secret := flag.String("secret", "", "HMAC_SECRET a usar para el sello (requerido con -seal; el mismo que usará el servidor)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "uso: hash_password [-seal -version V -secret S] <password>")
		os.Exit(2)
	}
	password := flag.Arg(0)

	hash, err := crypto.HashPassword(password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error hasheando password:", err)
		os.Exit(1)
	}
	fmt.Println("password_hash:", hash)

	if !*seal {
		return
	}
	if *version == "" || *secret == "" {
		fmt.Fprintln(os.Stderr, "error: -seal requiere -version y -secret")
		os.Exit(2)
	}

	id := uuid.Must(uuid.NewV7())
	acceptedAt := time.Now().UTC()
	payload := []byte(id.String() + "|" + *version + "|" + acceptedAt.Format(time.RFC3339Nano))
	sealHash := crypto.GenerateHMAC(payload, []byte(*secret))

	fmt.Println("id:                        ", id)
	fmt.Println("privacy_notice_version:    ", *version)
	fmt.Println("privacy_notice_accepted_at:", acceptedAt.Format(time.RFC3339Nano))
	fmt.Printf("privacy_acceptance_hash:    \\x%x\n", sealHash)
}
