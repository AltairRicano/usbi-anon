package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	pgconf "github.com/altair/usbi-anon-backend/deploy/postgres"
)

type pgconfValues struct {
	TotalRAMMB           int
	SharedBuffersMB      int
	EffectiveCacheSizeMB int
	WorkMemMB            int
	MaintenanceWorkMemMB int
	MaxConnections       int
	NetworkCIDR          string
}

// runPgconf implementa `usbictl pgconf render` (M3.7): toma las plantillas
// del repositorio (deploy/postgres/*.tmpl) y las rellena con los datos del
// servidor de destino — hace que el mismo repositorio sirva en un VPS de
// 1 GB o de 16 GB sin editar nada a mano. Conveniente, no imprescindible:
// con los .conf ya renderizados el despliegue funciona igual sin volver a
// correr esto.
func runPgconf(args []string) error {
	if len(args) == 0 || args[0] != "render" {
		return errors.New("uso: usbictl pgconf render [--total-ram-mb N] [--network CIDR] [--out-dir DIR]")
	}
	fs := flag.NewFlagSet("pgconf render", flag.ContinueOnError)
	// Default 1024 MB / red 172.28.0.0/16: el VPS de referencia del plan de
	// maduración (1 vCPU, 1 GB RAM, 20 GB disco) y la subred que
	// docker-compose.yml asigna a este proyecto.
	totalRAM := fs.Int("total-ram-mb", 1024, "RAM total del VPS en MB (se reparte entre db/api/web)")
	network := fs.String("network", "172.28.0.0/16", "CIDR de la red interna de docker compose")
	outDir := fs.String("out-dir", "./deploy/postgres/rendered", "directorio de salida")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if *totalRAM < 256 {
		return fmt.Errorf("--total-ram-mb=%d es demasiado bajo para correr Postgres+api+web", *totalRAM)
	}

	values := computePgconfValues(*totalRAM, *network)

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", *outDir, err)
	}

	for _, name := range []string{"postgresql.conf.tmpl", "pg_hba.conf.tmpl"} {
		dst := filepath.Join(*outDir, strings.TrimSuffix(name, ".tmpl"))
		if err := renderTemplate(name, dst, values); err != nil {
			return err
		}
		fmt.Printf("  %s\n", dst)
	}
	fmt.Printf("plantillas de Postgres renderizadas en %s (total-ram-mb=%d, network=%s)\n", *outDir, *totalRAM, *network)
	return nil
}

// computePgconfValues reparte totalRAMMB entre los tres servicios del
// compose (db/api/web) con una proporción conservadora fija para Postgres
// (~35%) — calibrado para el VPS de referencia del plan de maduración.
// Recalcular con --total-ram-mb en cuanto se confirme el tamaño real
// contratado (una de las decisiones abiertas del plan).
func computePgconfValues(totalRAMMB int, network string) pgconfValues {
	const postgresShare = 35 // % del total asignado a Postgres

	postgresMB := totalRAMMB * postgresShare / 100
	if postgresMB < 64 {
		postgresMB = 64
	}

	maxConnections := 20
	sharedBuffers := postgresMB / 4
	effectiveCache := postgresMB * 3 / 4
	workMem := sharedBuffers / maxConnections
	if workMem < 2 {
		workMem = 2
	}
	maintenanceWorkMem := postgresMB / 8
	if maintenanceWorkMem < 16 {
		maintenanceWorkMem = 16
	}

	return pgconfValues{
		TotalRAMMB:           totalRAMMB,
		SharedBuffersMB:      sharedBuffers,
		EffectiveCacheSizeMB: effectiveCache,
		WorkMemMB:            workMem,
		MaintenanceWorkMemMB: maintenanceWorkMem,
		MaxConnections:       maxConnections,
		NetworkCIDR:          network,
	}
}

func renderTemplate(srcName, dstPath string, values pgconfValues) error {
	tmplBytes, err := pgconf.FS.ReadFile(srcName)
	if err != nil {
		return fmt.Errorf("reading template %s: %w", srcName, err)
	}
	tmpl, err := template.New(srcName).Parse(string(tmplBytes))
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", srcName, err)
	}
	f, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dstPath, err)
	}
	defer f.Close()
	return tmpl.Execute(f, values)
}
