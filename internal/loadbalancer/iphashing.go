package loadbalancer

import (
	"hash/fnv"
	"net/http"
)

type IPHashing struct{}

func (i *IPHashing) NextBackend(backends []*Backend, req *http.Request) *Backend {
	if len(backends) == 0 {
		return nil
	}

	clientIP := req.RemoteAddr

	h := fnv.New32a()

	h.Write([]byte(clientIP))
	hashValue := h.Sum32()

	index := int(hashValue) % len(backends)

	attempts := 0

	for !backends[index].GetAlive() && attempts < len(backends) {
		index = (index + 1) % len(backends)
		attempts++
	}

	if attempts == len(backends) {
		return nil
	}

	return backends[index]
}
