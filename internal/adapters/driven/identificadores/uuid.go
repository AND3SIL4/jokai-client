package identificadores

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Generador struct{}

func Nuevo() *Generador {
	return &Generador{}
}

func (g *Generador) NuevoID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("no se pudo generar el identificador: %w", err)
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	enc := hex.EncodeToString(bytes[:])
	return enc[0:8] + "-" + enc[8:12] + "-" + enc[12:16] + "-" + enc[16:20] + "-" + enc[20:32], nil
}