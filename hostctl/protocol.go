package main

import "encoding/json"

// Protocolo: JSON-lines sobre un unix socket, una linea = un mensaje. 4 comandos fijos,
// sin argumentos libres -- nunca hay exec.Command con input del cliente. Ver
// CLAUDE.md ("hostctl: el helper de host") para el analisis de riesgos completo.
type request struct {
	Cmd   string `json:"cmd"`
	Token string `json:"token"`
}

type response struct {
	OK    bool            `json:"ok"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

func okResponse(data any) response {
	resp := response{OK: true}
	if data != nil {
		if b, err := json.Marshal(data); err == nil {
			resp.Data = b
		}
	}
	return resp
}

func errResponse(msg string) response {
	return response{OK: false, Error: msg}
}
