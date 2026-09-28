// Package migrations empotra los archivos *.sql de este directorio en el
// binario (M3.3 del plan de maduración): quien despliega el binario despliega
// también sus migraciones exactas, sin depender de que el árbol de archivos
// viaje aparte ni de que coincida con la versión del código.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
