package domain

const (
	VersionFormatoActual = 1

	CarpetaInbox      = "inbox"
	CarpetaEstado     = "estado"
	CarpetaProcesados = "procesados"

	SufijoComando = ".cmd"
	SufijoEstado  = ".st"
	SufijoTomando = ".tomando"
	SufijoTemp    = ".tmp"
)

func EsUUIDv4(id string) bool {
	if len(id) != 36 || id[14] != '4' {
		return false
	}
	for i := 0; i < len(id); i++ {
		switch i {
		case 8, 13, 18, 23:
			if id[i] != '-' {
				return false
			}
		default:
			if !esHex(id[i]) {
				return false
			}
		}
	}
	return true
}

func esHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}