package security

import (
	"fmt"
	"net"
	"net/http"
)

func CheckClientSubnet(IP *string, trustedSubnet *string) (bool, error) {
	parsedIP := net.ParseIP(*IP)
	if parsedIP == nil {
		return false, fmt.Errorf("error while parsing client's IP: %s", *IP)
	}

	_, trustedNet, err := net.ParseCIDR(*trustedSubnet)
	if err != nil {
		return false, fmt.Errorf("error while parsing trusted subnet's CIDR: %s", *trustedSubnet)
	}

	return trustedNet.Contains(parsedIP), nil
}

func SubnetCheckerMiddleware(trustedSubnet *string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {

		return func(w http.ResponseWriter, r *http.Request) {
			if *trustedSubnet != "" {
				IP := r.Header.Get("X-Real-IP")
				if IP == "" {
					http.Error(w, "X-Real-IP header is not found", http.StatusForbidden)
					return
				}
				isTrusted, err := CheckClientSubnet(&IP, trustedSubnet)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if !isTrusted {
					http.Error(w, "Client IP-address bot from trusted subnet", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		}
	}
}
