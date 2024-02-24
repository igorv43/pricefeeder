package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/pricefeeder/feeder/priceprovider/sources"
	"github.com/NibiruChain/pricefeeder/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/joho/godotenv"
)

var defaultExchangeSymbolsMap = map[string]map[asset.Pair]types.Symbol{
	// k-yang: default disable Coingecko because they have aggressive rate limiting
	// sources.Coingecko: {
	// 	"ubtc:uusd":  "bitcoin",
	// 	"ueth:uusd":  "ethereum",
	// 	"uusdt:uusd": "tether",
	// 	"uusdc:uusd": "usd-coin",
	// 	"uatom:uusd": "cosmos",
	// },
	sources.Bitfinex: {
		"ubtc:uusd":  "tBTCUSD",
		"ueth:uusd":  "tETHUSD",
		"uusdc:uusd": "tUDCUSD",
		"uusdt:uusd": "tUSTUSD",
		"uatom:uusd": "tATOUSD",
	},
	sources.GateIo: {
		"ubtc:uusd":  "BTC_USDT",
		"ueth:uusd":  "ETH_USDT",
		"uusdc:uusd": "USDC_USDT",
		"uusdt:uusd": "USDT_USD",
		"uatom:uusd": "ATOM_USDT",
	},
	sources.Okex: {
		"ubtc:uusd":  "BTC-USDT",
		"ueth:uusd":  "ETH-USDT",
		"uusdc:uusd": "USDC-USDT",
		"uusdt:uusd": "USDT-USDC",
		"uatom:uusd": "ATOM-USDT",
	},
}

func MustGet() *Config {
	conf, err := Get()
	if err != nil {
		panic(fmt.Sprintf("config error! check the environment: %v", err))
	}

	if conf == nil {
		panic(errors.New("invalid config"))
	}

	return conf
}

// Get loads the configuration from the .env file and returns a Config struct or an error
// if the configuration is invalid.
func Get() (*Config, error) {
	_ = godotenv.Load() // .env is optional

	conf := new(Config)
	conf.ChainID = os.Getenv("CHAIN_ID")
	conf.GRPCEndpoint = os.Getenv("GRPC_ENDPOINT")
	conf.WebsocketEndpoint = os.Getenv("WEBSOCKET_ENDPOINT")
	conf.FeederMnemonic = os.Getenv("FEEDER_MNEMONIC")
	conf.EnableTLS = os.Getenv("ENABLE_TLS") == "true"
	overrideExchangeSymbolsMapJson := os.Getenv("EXCHANGE_SYMBOLS_MAP")
	overrideExchangeSymbolsMap := map[string]map[string]string{}
	err := json.Unmarshal([]byte(overrideExchangeSymbolsMapJson), &overrideExchangeSymbolsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EXCHANGE_SYMBOLS_MAP: invalid json")
	}

	conf.ExchangesToPairToSymbolMap = defaultExchangeSymbolsMap
	for exchange, symbolMap := range overrideExchangeSymbolsMap {
		conf.ExchangesToPairToSymbolMap[exchange] = map[asset.Pair]types.Symbol{}
		for nibiAssetPair, tickerSymbol := range symbolMap {
			conf.ExchangesToPairToSymbolMap[exchange][asset.MustNewPair(nibiAssetPair)] = types.Symbol(tickerSymbol)
		}
	}

	// datasource config map
	datasourceConfigMapJson := os.Getenv("DATASOURCE_CONFIG_MAP")
	datasourceConfigMap := map[string]json.RawMessage{}

	if datasourceConfigMapJson != "" {
		err = json.Unmarshal([]byte(datasourceConfigMapJson), &datasourceConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DATASOURCE_CONFIG_MAP: invalid json")
		}
	}
	conf.DataSourceConfigMap = datasourceConfigMap

	// optional validator address (for delegated feeders)
	valAddrStr := os.Getenv("VALIDATOR_ADDRESS")
	if valAddrStr != "" {
		valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
		if err == nil {
			conf.ValidatorAddr = &valAddr
		}
	}

	return conf, conf.Validate()
}

type Config struct {
	ExchangesToPairToSymbolMap map[string]map[asset.Pair]types.Symbol
	DataSourceConfigMap        map[string]json.RawMessage
	GRPCEndpoint               string
	WebsocketEndpoint          string
	FeederMnemonic             string
	ChainID                    string
	ValidatorAddr              *sdk.ValAddress
	EnableTLS                  bool
}

func (c *Config) Validate() error {
	if c.ChainID == "" {
		return fmt.Errorf("no chain id")
	}
	if c.FeederMnemonic == "" {
		return fmt.Errorf("no feeder mnemonic")
	}
	if c.WebsocketEndpoint == "" {
		return fmt.Errorf("no websocket endpoint")
	}
	if c.GRPCEndpoint == "" {
		return fmt.Errorf("no grpc endpoint")
	}
	return nil
}
