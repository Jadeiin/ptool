package impersonateutil

import (
	"fmt"
	"slices"

	"github.com/Noooste/azuretls-client"
	tls "github.com/Noooste/utls"
)

func ApplyJA3(session *azuretls.Session, ja3, navigator string) error {
	if err := session.ApplyJa3(ja3, navigator); err != nil {
		return err
	}
	getSpec := session.GetClientHelloSpec
	var groups []tls.CurveID
	hasKeyShare := false
	for _, extension := range getSpec().Extensions {
		switch extension := extension.(type) {
		case *tls.SupportedCurvesExtension:
			groups = extension.Curves
		case *tls.KeyShareExtension:
			hasKeyShare = true
		}
	}
	if !hasKeyShare {
		return nil
	}
	shares, err := selectKeyShares(groups)
	if err != nil {
		return err
	}
	session.GetClientHelloSpec = func() *tls.ClientHelloSpec {
		spec := getSpec()
		for _, extension := range spec.Extensions {
			if keyShare, ok := extension.(*tls.KeyShareExtension); ok {
				keyShare.KeyShares = slices.Clone(shares)
			}
		}
		return spec
	}
	return nil
}

func selectKeyShares(groups []tls.CurveID) ([]tls.KeyShare, error) {
	// uTLS retains one classical private key and one hybrid private key. Its
	// hybrid path also needs the classical key to use X25519.
	hybrid := tls.CurveID(0)
	classical := tls.CurveID(0)
	for _, group := range groups {
		switch group {
		case tls.X25519MLKEM768, tls.X25519Kyber768Draft00:
			if hybrid != 0 && hybrid != group {
				return nil, fmt.Errorf("multiple hybrid key-exchange groups (%d and %d) are unsupported", hybrid, group)
			}
			hybrid = group
		case tls.X25519, tls.CurveP256, tls.CurveP384, tls.CurveP521:
			if classical == 0 {
				classical = group
			}
		}
	}
	if hybrid != 0 {
		if !slices.Contains(groups, tls.X25519) {
			return nil, fmt.Errorf("hybrid key-exchange group %d requires X25519 (29) in JA3 supported groups", hybrid)
		}
		classical = tls.X25519
	}
	// Offer shares in the same order as their entries in supported_groups.
	var shares []tls.KeyShare
	for _, group := range groups {
		switch group {
		case tls.GREASE_PLACEHOLDER:
			shares = append(shares, tls.KeyShare{Group: group, Data: []byte{0}})
		case classical, hybrid:
			if group != 0 && !slices.ContainsFunc(shares, func(share tls.KeyShare) bool { return share.Group == group }) {
				shares = append(shares, tls.KeyShare{Group: group})
			}
		}
	}
	return shares, nil
}
