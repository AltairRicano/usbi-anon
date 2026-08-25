// cmd/hash_password es el binario standalone que pide el rediseño de
// identidad (plan/04_Rediseno_identidad_gustos.md §2) para el bootstrap del
// primer admin: sin base de datos, sin red, imprime el hash Argon2id de un
// password dado — el mismo internal/crypto.HashPassword que usa el servidor,
// así que el hash resultante es válido para pegar directo en
// backend/sql/01_seed_primer_admin.sql.
//
// Extensión pragmática, no pedida literalmente por el plan pero necesaria
// para que ese seed sea ejecutable: con --seal se imprime TAMBIÉN un UUID
// nuevo y el sello HMAC de accounts.privacy_acceptance_hash. La razón es que
// ni gen_random_uuid() ni hmac() (ambas de pgcrypto) están disponibles —la
// extensión se eliminó a propósito en F5, no queda un solo campo cifrado en
// el sistema— así que no hay forma de generarlos en SQL puro. Pedirle al
// operador que replique un HMAC-SHA256 exacto a mano (p. ej. con openssl,
// cuidando que ni un byte de más se cuele en la clave) es mucho más
// propenso a error que reutilizar aquí la misma función
// (internal/crypto.GenerateHMAC) que el servidor usará después para
// verificarlo.
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
