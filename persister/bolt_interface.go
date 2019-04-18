package persister

import "github.com/ChainTex/server-go/tomochain"

type BoltInterface interface {
	StoreGeneralInfo(map[string]*tomochain.TokenGeneralInfo) error
	GetGeneralInfo(map[string]tomochain.Token) (map[string]*tomochain.TokenGeneralInfo, error)
}
