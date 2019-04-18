package persister

import (
	"encoding/json"
	"log"

	"github.com/ChainTex/server-go/tomochain"
	"github.com/influxdata/influxdb1-client/v2"
)

const (
	path   = "./persister/db/market.db"
	database = "chaintex_db"
	point_name = "market_info"
)

// BoltStorage storage for cache
type InfluxDb struct {
	httpClient *NewHTTPClient
}

// NewInfluxClient make bolt instance
func NewInfluxStorage() (*InfluxDb, error) {
	config = HTTPConfig{
		Addr:     os.Getenv("INFLUXDB_ADDR"),
		Username: os.Getenv("INFLUXDB_USER"),
		Password: os.Getenv("INFLUXDB_USER_PASSWORD"),
	}

	httpClient, err := NewHTTPClient(config)
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}

	return &InfluxDb{
		httpClient: httpClient,
	}, nil
}

// StoreGeneralInfo store market info
func (marketDB *InfluxDb) StoreGeneralInfo(mapInfo map[string]*tomochain.TokenGeneralInfo) error {
	var err error
	bp, _ := NewBatchPoints(BatchPointsConfig{Database: database})
	
	tags := map[string]string{}

	bp.AddPoint(NewPoint(
		point_name,
		tags,
		mapInfo,
		time.Now(),
	))

	err := marketDB.httpClient.Write(bp)

	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

// GetGeneralInfo store market info
func (bs *BoltStorage) GetGeneralInfo(mapToken map[string]tomochain.Token) (map[string]*tomochain.TokenGeneralInfo, error) {
	var err error
	result := make(map[string]*tomochain.TokenGeneralInfo)
	err = bs.marketDB.View(func(tx *bolt.Tx) error {
		var errV error
		b := tx.Bucket([]byte(bucket))
		if errV = b.ForEach(func(k, v []byte) error {
			var tokenInfo tomochain.TokenGeneralInfo
			errLoop := json.Unmarshal(v, &tokenInfo)
			if errLoop != nil {
				return errLoop
			}
			result[string(k)] = &tokenInfo
			return nil
		}); errV != nil {
			log.Println(errV.Error())
			return errV
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	return result, nil
}
