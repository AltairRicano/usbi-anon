package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// defaultSecretsPath vive fuera del árbol del repositorio a propósito
// (M3.6): un `git add -A` distraído no puede alcanzarlo ni por accidente.
// En el entorno de desarrollo de este repo, sin permisos de root, se puede
// sobreescribir con --out a una ruta propia.
const defaultSecretsPath = "/etc/usbi-anon/env"

type secretSpec struct {
	key   string
	bytes int
	doc   string
}

// machineSecrets son los secretos que NUNCA debe teclear una persona — se
// generan con crypto/rand y nadie los vuelve a escribir después (M3.6). La
// contraseña del primer administrador NO está aquí: es la única credencial
// que sí pide esta herramienta por consola (ver runSecrets).
var machineSecrets = []secretSpec{
	{key: "JWT_SECRET", bytes: 32, doc: "firma los JWT de sesión"},
	{key: "HMAC_SECRET", bytes: 32, doc: "sella refresh tokens y aceptación del aviso de privacidad"},
	{key: "DB_SUPERUSER_PASSWORD", bytes: 24, doc: "contraseña del superusuario postgres (solo bootstrap del clúster, ver backend/sql/00_init_cluster.sh)"},
	{key: "DB_APP_PASSWORD", bytes: 24, doc: "contraseña del rol usbi_app"},
	{key: "DB_MODERADOR_PASSWORD", bytes: 24, doc: "contraseña del rol usbi_moderador"},
	{key: "DB_MIGRATE_PASSWORD", bytes: 24, doc: "contraseña del rol usbi_migrate"},
	{key: "DB_DBMAINT_PASSWORD", bytes: 24, doc: "contraseña del rol usbi_dbmaint"},
}

// runSecrets implementa `usbictl secrets init [--rotate <clave>] [--out <archivo>]`.
func runSecrets(args []string) error {
	if len(args) == 0 || args[0] != "init" {
		return errors.New("uso: usbictl secrets init [--rotate <clave>] [--out <archivo>]")
	}
	fs := flag.NewFlagSet("secrets init", flag.ContinueOnError)
	rotateKey := fs.String("rotate", "", "rota solo esta clave, deja las demás intactas")
	outPath := fs.String("out", defaultSecretsPath, "archivo de destino (se crea con permisos 600)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	existing, existedBefore, err := readEnvFile(*outPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", *outPath, err)
	}

	if *rotateKey != "" {
		spec := findSecretSpec(*rotateKey)
		if spec == nil {
			return fmt.Errorf("clave desconocida para --rotate: %q (válidas: %s)", *rotateKey, secretKeyNames())
		}
		value, err := randomHex(spec.bytes)
		if err != nil {
			return err
		}
		existing[spec.key] = value
		if err := writeEnvFile(*outPath, existing); err != nil {
			return err
		}
		fmt.Printf("rotado %s en %s.\n", spec.key, *outPath)
		fmt.Println("Reinicia todo servicio que use esa credencial para que tome el valor nuevo.")
		return nil
	}

	// Idempotente: si el archivo ya existía, no lo pisa completo — solo
	// --rotate cambia una clave existente (M3.6).
	if existedBefore {
		fmt.Printf("%s ya existe con %d clave(s) — no se sobrescribe.\n", *outPath, len(existing))
		fmt.Println("Usa --rotate <clave> para cambiar una sola, o borra el archivo para regenerar todo.")
		return nil
	}

	for _, spec := range machineSecrets {
		value, err := randomHex(spec.bytes)
		if err != nil {
			return fmt.Errorf("generating %s: %w", spec.key, err)
		}
		existing[spec.key] = value
	}

	// La única credencial que una persona tiene que recordar y escribir
	// (M3.6) — validada por auth.Service.CreateAdminAccount cuando
	// `usbictl admin create` la use, no aquí, para no duplicar la política.
	password, err := adminBootstrapPassword()
	if err != nil {
		return err
	}
	existing["ADMIN_BOOTSTRAP_PASSWORD"] = password

	if err := writeEnvFile(*outPath, existing); err != nil {
		return err
	}
	fmt.Printf("secretos generados en %s (permisos 600).\n", *outPath)
	fmt.Println("Siguiente paso: usbictl admin create")
	return nil
}

func findSecretSpec(key string) *secretSpec {
	for i := range machineSecrets {
		if machineSecrets[i].key == key {
			return &machineSecrets[i]
		}
	}
	return nil
}

func secretKeyNames() string {
	names := make([]string, len(machineSecrets))
	for i, s := range machineSecrets {
		names[i] = s.key
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func randomHex(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// readEnvFile lee pares CLAVE=VALOR de un archivo estilo .env. Un archivo
// ausente no es error: devuelve un mapa vacío y existed=false.
func readEnvFile(path string) (values map[string]string, existed bool, err error) {
	values = make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return values, false, nil
		}
		return nil, false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, true, err
	}
	return values, true, nil
}

// writeEnvFile escribe `values` en `path` con permisos 600, creando el
// directorio contenedor si hace falta. Las claves salen ordenadas para que
// el archivo sea diffable si alguna vez se versiona por error.
func writeEnvFile(path string, values map[string]string) error {
	if dir := dirOf(path); dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("# Generado por usbictl secrets init — no versionar (ver .gitignore).\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\n", k, values[k])
	}

	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func dirOf(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx <= 0 {
		return ""
	}
	return path[:idx]
}
