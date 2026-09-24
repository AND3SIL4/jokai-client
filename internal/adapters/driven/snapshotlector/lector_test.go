package snapshotlector

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"jokai-client/internal/domain"
)

func escribirSnapshot(t *testing.T, raiz, nombre, contenido string) string {
	t.Helper()
	ruta := filepath.Join(raiz, nombre)
	if err := os.WriteFile(ruta, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
	return ruta
}

func TestLeerSnapshotAusente(t *testing.T) {
	l := NuevoLector(t.TempDir(), "estado.json")
	_, err := l.Leer(t.Context())
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(strings.ToLower(eu.Causa), "consolidador") {
		t.Errorf("la causa debe mencionar al consolidador: %q", eu.Causa)
	}
	if eu.Sugerencia == "" {
		t.Error("debe haber sugerencia")
	}
}

func TestLeerSnapshotMalformado(t *testing.T) {
	raiz := t.TempDir()
	escribirSnapshot(t, raiz, "estado.json", `{"version_formato": `)
	l := NuevoLector(raiz, "estado.json")
	_, err := l.Leer(t.Context())
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(strings.ToLower(eu.Causa), "snapshot") {
		t.Errorf("la causa debe mencionar el snapshot: %q", eu.Causa)
	}
}

func TestLeerSnapshotVersionMayor(t *testing.T) {
	raiz := t.TempDir()
	escribirSnapshot(t, raiz, "estado.json", `{"version_formato": 99, "generado_en": "2026-09-22T14:00:00Z", "maquinas": [], "solicitudes": [], "ejecuciones": []}`)
	l := NuevoLector(raiz, "estado.json")
	_, err := l.Leer(t.Context())
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "nueva") || !strings.Contains(eu.Causa, "consolidador") {
		t.Errorf("la causa debe pedir actualizar el cliente: %q", eu.Causa)
	}
}

func TestLeerSnapshotValido(t *testing.T) {
	raiz := t.TempDir()
	escribirSnapshot(t, raiz, "estado.json", `{
  "version_formato": 1,
  "generado_en": "2026-09-22T14:00:00Z",
  "maquinas": [
    { "nombre": "BOTPC01", "ambiente": "prod",
      "ultimo_latido": "2026-09-22T13:59:00Z",
      "bots": ["facturacion", "nomina"] }
  ],
  "solicitudes": [
    { "id": "550e8400-e29b-41d4-a716-446655440000", "bot_nombre": "facturacion",
      "bot_ambiente": "prod", "maquina": "BOTPC01", "usuario": "JOKAI\\ana",
      "fecha": "2026-09-22T14:00:00Z", "estado": "SOLICITADA" }
  ],
  "ejecuciones": [
    { "id": "550e8400-e29b-41d4-a716-446655440001", "solicitud_id": "550e8400-e29b-41d4-a716-446655440000",
      "estado": "EN_PROCESO", "inicio": null, "fin": null, "codigo": null, "mensaje": null }
  ]
}`)
	l := NuevoLector(raiz, "estado.json")
	snap, err := l.Leer(t.Context())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if snap.VersionFormato != domain.VersionFormatoActual {
		t.Errorf("VersionFormato = %d", snap.VersionFormato)
	}
	if len(snap.Maquinas) != 1 {
		t.Fatalf("se esperaba 1 máquina, hay %d", len(snap.Maquinas))
	}
	m := snap.Maquinas[0]
	if m.Nombre != "BOTPC01" || m.Ambiente.String() != "prod" {
		t.Errorf("máquina incorrecta: %+v", m)
	}
	if m.UltimoLatido == nil || m.UltimoLatido.Year() != 2026 {
		t.Errorf("ultimo_latido no mapeado: %+v", m.UltimoLatido)
	}
	if len(m.Bots) != 2 {
		t.Errorf("bots = %v", m.Bots)
	}

	if len(snap.Solicitudes) != 1 {
		t.Fatalf("se esperaba 1 solicitud, hay %d", len(snap.Solicitudes))
	}
	s := snap.Solicitudes[0]
	if s.ID != "550e8400-e29b-41d4-a716-446655440000" || s.BotNombre != "facturacion" ||
		s.Maquina != "BOTPC01" || s.Usuario != `JOKAI\ana` {
		t.Errorf("solicitud incorrecta: %+v", s)
	}
	if s.Estado != domain.EstadoSolicitada {
		t.Errorf("estado = %q", s.Estado)
	}
	want := time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC)
	if !s.Fecha.Equal(want) {
		t.Errorf("fecha = %v, se esperaba %v", s.Fecha, want)
	}

	if len(snap.Ejecuciones) != 1 {
		t.Fatalf("se esperaba 1 ejecución, hay %d", len(snap.Ejecuciones))
	}
	e := snap.Ejecuciones[0]
	if e.ID != "550e8400-e29b-41d4-a716-446655440001" || e.Estado != domain.EstadoEnProceso {
		t.Errorf("ejecución incorrecta: %+v", e)
	}
}

func TestLeerSnapshotEstadoDesconocido(t *testing.T) {
	raiz := t.TempDir()
	escribirSnapshot(t, raiz, "estado.json", `{
  "version_formato": 1,
  "generado_en": "2026-09-22T14:00:00Z",
  "maquinas": [],
  "solicitudes": [
    { "id": "550e8400-e29b-41d4-a716-446655440000", "bot_nombre": "facturacion",
      "bot_ambiente": "prod", "maquina": "BOTPC01", "usuario": "JOKAI\\ana",
      "fecha": "2026-09-22T14:00:00Z", "estado": "RARO" }
  ],
  "ejecuciones": []
}`)
	l := NuevoLector(raiz, "estado.json")
	_, err := l.Leer(t.Context())
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
}