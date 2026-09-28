// usbictl es el binario único de operación de USBI-Anon (M3.4 del plan de
// maduración): un solo ejecutable con subcomandos, no varios binarios
// descartables. Los cinco subcomandos comparten internal/config,
// internal/crypto e internal/repository, y un "descartable" que termina
// tocando la base de producción nunca se descarta en la práctica — mejor que
// lo sea a propósito, documentado y con pruebas.
package main

import (
	"fmt"
	"os"

	"github.com/altair/usbi-anon-backend/internal/config"
)

func main() {
	config.LoadEnvironment()

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "migrate":
		err = runMigrate(os.Args[2:])
	case "secrets":
		err = runSecrets(os.Args[2:])
	case "admin":
		err = runAdmin(os.Args[2:])
	case "doctor":
		err = runDoctor(os.Args[2:])
	case "pgconf":
		err = runPgconf(os.Args[2:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "usbictl: subcomando desconocido %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "usbictl:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usbictl — operación de USBI-Anon (M3 del plan de maduración)

Uso: usbictl <subcomando> [flags]

Subcomandos:
  migrate up|down [N]|force <version>|version   Aplica/revierte migraciones (rol usbi_migrate)
  secrets init [--rotate <clave>] [--out <archivo>]   Genera secretos de máquina
  admin create                                   Da de alta el primer administrador
  doctor                                         Verifica los tres pools y la matriz de permisos
  pgconf render                                  Rellena las plantillas de configuración de Postgres
`)
}
